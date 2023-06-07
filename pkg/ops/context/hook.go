package context

import (
	"context"
	"fmt"
	"time"

	"github.com/superops-team/hyperops/pkg/environment"
	"github.com/superops-team/hyperops/pkg/metrics"
	"github.com/rs/zerolog/log"
	"go.starlark.net/starlark"
)

const (
	DRYRUN_NAME        = "HYPEROPS_DRYRUN"
	HYPEROPS_FUNC_HOOK = "HYPEROPS_FUNC_HOOK"
)

type Function func(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error)

// isCancelled 判断外部是不是执行了取消操作
func isCancelled(thread *starlark.Thread) error {
	if ctx, ok := thread.Local(thread.Name).(context.Context); ok {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
	}
	return nil
}

// preRun 自定义lib能力执行前hook
func preRun(thread *starlark.Thread) error {
	err := isCancelled(thread)
	if err != nil {
		return err
	}
	tm := GetTaskManager()
	task := tm.Get(thread.Name)
	if task == nil {
		return nil
	}
	if task.status == PreHangingStatus {
		task.TrigerEvent(HangingStatus)
		_ = tm.StartHanging(thread.Name)
		select {
		case recoveryDesc := <-task.RecoveryCh:
			if recoveryDesc == "recovery" {
				task.TrigerEvent(RunningStatus)
				_ = tm.RecoveryOver(thread.Name)
				return nil
			}
			return nil
		// task max hang time is 24 hour
		case <-time.After(24 * time.Hour):
			task.TrigerEvent(RunningStatus)
			return ErrHangTimeout
		}
	}
	return nil
}

// PostRun 自定义lib执行后hook
func postRun(thread *starlark.Thread) {
}

// AddBuiltin hook starlark.NewBuiltin for add pre and post func when exec self func
func AddBuiltin(name string, f Function) *starlark.Builtin {
	wrapped := starlark.NewBuiltin(name, func(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
		start := time.Now()
		var err error
		defer func() {
			sm := NewSecretsManager()
			var msg string
			if err != nil {
				msg = fmt.Sprintf("fn=%s, args=%s, kwargs=%s, dur=%s, err=%s", name, args, kwargs, time.Since(start), err.Error())
			} else {
				msg = fmt.Sprintf("fn=%s, args=%s, kwargs=%s, dur=%s", name, args, kwargs, time.Since(start))
			}
			safeMsg := sm.SafeReplace(msg)
			env := environment.NewEnvStorage()
			if env.IsTrue(HYPEROPS_FUNC_HOOK) {
				thread.Print(thread, safeMsg)
			}
			status := "success"
			if err != nil {
				log.Error().Err(err).Msg(safeMsg)
				if !env.IsTrue(HYPEROPS_FUNC_HOOK) {
					thread.Print(thread, safeMsg)
				}
				status = "failed"
			}
			metrics.HyperFnCounter.WithLabelValues(name, status).Inc()
			metrics.HyperFnDurHis.WithLabelValues(name, status).Observe(float64(time.Since(start).Milliseconds()))
		}()
		err = preRun(thread)
		if err != nil {
			return starlark.None, err
		}
		res, err := f(thread, fn, args, kwargs)
		postRun(thread)
		return res, err
	})
	return wrapped
}
