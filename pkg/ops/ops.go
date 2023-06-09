package ops

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"runtime/debug"
	"sync"
	"time"

	"github.com/superops-team/hyperops/pkg/metrics"
	localctx "github.com/superops-team/hyperops/pkg/ops/context"
	"github.com/superops-team/hyperops/pkg/ops/event"
	"github.com/superops-team/hyperops/pkg/ops/starlib"
	"github.com/superops-team/hyperops/pkg/ops/starlib/sh"
	"go.starlark.net/resolve"
	"go.starlark.net/starlark"
)

const (
	defaultContextName = "hyperops_context"
)

// ModuleLoader 模块加载声明
type ModuleLoader func(thread *starlark.Thread, module string) (starlark.StringDict, error)

// DefaultModuleLoader 默认模块加载
var DefaultModuleLoader = func(thread *starlark.Thread, module string) (dict starlark.StringDict, err error) {
	return starlib.Loader(thread, module)
}

// Runtime ops解析引擎运行时
type Runtime struct {
	sync.Mutex
	ctx          context.Context
	target       *Target
	globals      starlark.StringDict
	ctxConfig    map[string]interface{}
	ctxSecrects  map[string]string
	EventsCh     chan event.Event
	output       io.Writer
	moduleLoader ModuleLoader
	thread       *starlark.Thread
	predeclared  starlark.StringDict
}

func (r *Runtime) SetThread(thread *starlark.Thread) {
	r.thread = thread
}

// GetThread 获取starlark thread
func (r *Runtime) GetThread() *starlark.Thread {
	return r.thread
}

// hyperopsPrint 提供异步定制化的输出
func (r *Runtime) hyperopsPrint(thread *starlark.Thread, msg string) {
	sm := localctx.NewSecretsManager()
	safeMsg := sm.SafeReplace(msg)

	// 提供流式实时数据到外部使用者，保证执行过程会同步打印
	if r.EventsCh != nil {
		payload := event.PrintEvent{
			ID:  thread.Name,
			Msg: safeMsg,
		}
		r.EventsCh <- event.MakeEvent(event.ETPrint, thread.Name, payload)
	}

	// 统一将所有的输出记录到output，为外部提供存档记录
	_, err := r.output.Write([]byte(safeMsg + "\n"))
	if err != nil {
		fmt.Println(err.Error())
	}
}

// ExecScript 执行脚本
func ExecScript(ctx context.Context, target *Target, opts ...func(o *ExecOpts)) (err error) {
	// Recover from errors.
	now := time.Now()

	defer func() {
		latency := time.Since(now)
		if r := recover(); r != nil {
			fmt.Printf("running hyperops script panic reason: %v \n  statck: %v", r, debug.Stack())
		}
		if err != nil {
			metrics.WorkCount.WithLabelValues(target.ScriptPath, "failed").Inc()
			metrics.WorkDuration.WithLabelValues(target.ScriptPath, "failed").Observe(latency.Seconds())
		} else {
			metrics.WorkCount.WithLabelValues(target.ScriptPath, "succeed").Inc()
			metrics.WorkDuration.WithLabelValues(target.ScriptPath, "succeed").Observe(latency.Seconds())
		}
	}()
	o := &ExecOpts{}
	DefaultExecOpts(o)
	for _, opt := range opts {
		if opt == nil {
			return fmt.Errorf("nil option passed to ExecScript")
		}
		opt(o)
	}

	resolve.AllowFloat = o.AllowFloat
	resolve.AllowSet = o.AllowSet
	resolve.AllowLambda = o.AllowLambda
	resolve.AllowNestedDef = o.AllowNestedDef
	resolve.AllowGlobalReassign = o.AllowGlobalReassign

	// 增加错误处理内置函数
	r := &Runtime{
		ctx:          ctx,
		EventsCh:     o.EventsCh,
		ctxConfig:    o.Locals,
		ctxSecrects:  o.Secrets,
		target:       target,
		output:       o.OutputWriter,
		moduleLoader: o.ModuleLoader,
		predeclared: starlark.StringDict{
			"sh":    localctx.AddBuiltin("sh", sh.Exec),                // 将sh提升为一级内置函数，无需导入
			"sleep": localctx.AddBuiltin("sleep", SleepFn),             // 将sleep函数提升为内置，无需导入
			"ctx":   localctx.NewContext(o.Locals, o.Secrets).Struct(), // 每个实例绑定运行时上下文，用于记录该实例的各种状态
		},
	}
	// 收敛所有的print的逻辑，避免使用的时候混淆, 尽最大可能保证和python内置的一致性体验
	// 后续开发包也一样会遵守该原则
	thread := &starlark.Thread{Load: r.moduleLoader, Print: r.hyperopsPrint} // replace SafePrint to hyperopsPrint for only one place to print is more easy to use for two

	// 如果传递了jobID那么上下文可以绑定到jobid上
	ctxName := defaultContextName
	jobID, ok := r.ctxConfig["job_id"]
	if ok {
		ctxName, _ = jobID.(string)
	}
	thread.Name = ctxName
	r.SetThread(thread)

	// for outside manager all tasks
	tm := localctx.NewTaskManager()
	tm.Add(ctxName, thread, r.EventsCh)

	// add timeout when exec time exceeded
	go func() {
		for {
			select {
			case <-time.After(o.Timeout):
				if thread != nil {
					thread.Cancel(fmt.Sprintf("exec %s timeout %s", thread.Name, o.Timeout))
				}
				return
			case <-ctx.Done():
				return
			}
		}
	}()

	if len(target.ScriptContent) > 0 {
		evalName := filepath.Base(target.ScriptPath)
		thread.SetLocal(evalName, evalName)
		r.globals, err = starlark.ExecFile(thread, evalName, target.ScriptContent, r.predeclared)
	} else {
		r.globals, err = starlark.ExecFile(thread, target.ScriptPath, nil, r.predeclared)
	}
	tm.Delete(ctxName, r.predeclared)
	if err != nil {
		if evalErr, ok := err.(*starlark.EvalError); ok {
			return fmt.Errorf(evalErr.Backtrace())
		}
		return err
	}
	return err
}
