package context

import (
	"errors"
	"fmt"
	"os"
	"path"
	"sync"
	"time"

	"github.com/superops-team/hyperops/pkg/environment"
	"github.com/superops-team/hyperops/pkg/metrics"
	"github.com/superops-team/hyperops/pkg/ops/event"
	"go.starlark.net/starlark"
)

var (
	onceTM     sync.Once
	instanceTM *TaskManager

	ErrSuspendFailed        = errors.New("error suspend job failed, reason not found job, maybe has already finished, or task was not running on this worker")
	ErrRecoveryFailed       = errors.New("error recovery job failed, reason not found job, maybe has already finished, or task was not running on this worker")
	ErrSuspendIsRecovring   = errors.New("error suspend job failed, is recovering")
	ErrSuspendIsHanging     = errors.New("error suspend job failed, is hanging")
	ErrSuspendIsPreHanging  = errors.New("error suspend job failed, is prehanging")
	ErrRecoveryIsRecovring  = errors.New("error recovery job failed, is recovering")
	ErrRecoveryIsNotHanging = errors.New("error recovery job failed, is not hanging")
	ErrHangTimeout          = errors.New("error hang task dutaiton > 24h, exit with timeout error")
	ErrTaskKill             = errors.New("error task was killed when hanging")
)

// 任务状态定义
type TaskStatus string

var (
	PendingStatus    TaskStatus = "pending"
	RunningStatus    TaskStatus = "running"
	HangingStatus    TaskStatus = "hanging"
	PreHangingStatus TaskStatus = "prehanging"
	FinishedStatus   TaskStatus = "finished"
)

// Task 任务描述
type Task struct {
	ID         string
	status     TaskStatus
	recovering bool
	thread     *starlark.Thread
	RecoveryCh chan string
	eventsCh   chan event.Event
	hangTime   time.Time
}

// TrigerEvent 变更状态后自动触发事件
func (t *Task) TrigerEvent(curStatus TaskStatus) {
	// 保证任务存在后执行避免死锁
	preStatus := t.status
	if string(preStatus) == "" {
		preStatus = PendingStatus
	}
	if t.eventsCh != nil {
		t.eventsCh <- event.MakeEvent(event.ETTask, t.ID, event.TaskEvent{
			ID:   t.ID,
			From: string(preStatus),
			To:   string(curStatus),
		})
	}
}

// TrigerDataEvent 触发数据事件,收到事件后进行归档操作
func (t *Task) TrigerDataEvent(data map[string]interface{}) {
	if t.eventsCh != nil {
		t.eventsCh <- event.MakeEvent(event.ETData, t.ID, event.DataEvent{
			ID:   t.ID,
			Data: data,
		})
	}
}

func (t *Task) GetStatus() TaskStatus {
	return t.status
}

// TaskManager 任务管理器,管理全局任务状态
type TaskManager struct {
	sync.Mutex
	tasks map[string]*Task
}

// NewTaskManager 创建任务管理器
func NewTaskManager() *TaskManager {
	onceTM.Do(func() {
		instanceTM = &TaskManager{
			tasks: make(map[string]*Task),
		}
	})
	return instanceTM
}

//判断路径是否存在
func PathExists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	if os.IsExist(err) {
		return true
	}
	return false
}

// Add 添加task
func (t *TaskManager) Add(taskid string, thread *starlark.Thread, eventsCh chan event.Event) {
	t.Lock()
	defer t.Unlock()
	env := environment.NewEnvStorage()
	pwdpath := env.Get("PWD")
	workdir := "./"
	if len(pwdpath) != 0 && len(thread.Name) != 0 {
		workdir = path.Join(pwdpath, thread.Name)
	}
	if !PathExists(workdir) {
		_ = os.MkdirAll(workdir, 0755)
	}
	_, ok := t.tasks[taskid]
	if ok {
		return
	}
	task := &Task{
		ID:         taskid,
		thread:     thread,
		recovering: false,
		RecoveryCh: make(chan string, 1),
		eventsCh:   eventsCh,
	}
	task.TrigerEvent(RunningStatus)
	task.status = RunningStatus

	t.tasks[taskid] = task
}

// Delete 删除task
func (t *TaskManager) Delete(taskid string, dict starlark.StringDict) {
	t.Lock()
	defer t.Unlock()
	var iskeep bool
	task, ok := t.tasks[taskid]
	if !ok {
		return
	}
	if task != nil {
		task.TrigerEvent(FinishedStatus)
	}
	thread := task.thread
	v, err := Call(thread, dict, "ctx.get_config", []interface{}{"HYPEROPS_WORKSPACE_KEEP"}, nil)
	if err == nil {
		iskeep, _ = v.(bool)
	}
	if !iskeep {
		env := environment.NewEnvStorage()
		pwd := env.Get("PWD")
		path := path.Join(pwd, thread.Name)
		if PathExists(path) {
			os.RemoveAll(path)
		}
	}
	values, err := Call(thread, dict, "ctx.values", nil, nil)
	if err == nil {
		evData := values.(map[string]interface{})
		// 当事件非空时才触发
		if len(evData) > 0 {
			task.TrigerDataEvent(evData)
		}
	}

	delete(t.tasks, taskid)
}

// Get 获取task
func (t *TaskManager) Get(taskid string) *Task {
	t.Lock()
	defer t.Unlock()
	task, ok := t.tasks[taskid]
	if !ok {
		return nil
	}
	return task
}

// Suspend 暂停指定task
func (t *TaskManager) Suspend(taskid string) error {
	t.Lock()
	defer t.Unlock()
	task, ok := t.tasks[taskid]
	if !ok {
		return ErrSuspendFailed
	}
	if task.status == PreHangingStatus {
		return ErrSuspendIsPreHanging
	}
	if task.status == HangingStatus {
		return ErrSuspendIsHanging
	}
	if task.recovering {
		return ErrSuspendIsRecovring
	}
	task.hangTime = time.Now()
	metrics.HangGouge.WithLabelValues("hanging").Inc()
	task.status = PreHangingStatus
	return nil
}

// Recovery 恢复指定task
func (t *TaskManager) Recovery(taskid string) error {
	t.Lock()
	defer t.Unlock()
	task, ok := t.tasks[taskid]
	if !ok {
		return ErrRecoveryFailed
	}
	if task.status != PreHangingStatus && task.status != HangingStatus {
		return ErrRecoveryIsNotHanging
	}
	if task.recovering {
		return ErrRecoveryIsRecovring
	}
	task.recovering = true
	task.RecoveryCh <- "recovery"
	return nil
}

func (t *TaskManager) RecoveryOver(taskid string) error {
	t.Lock()
	defer t.Unlock()
	task, ok := t.tasks[taskid]
	if !ok {
		return ErrRecoveryFailed
	}
	task.status = RunningStatus
	metrics.WorkDuration.WithLabelValues("hyperops", "hanging").Observe(time.Since(task.hangTime).Seconds())
	metrics.HangGouge.WithLabelValues("hanging").Dec()
	task.recovering = false
	return nil
}

func (t *TaskManager) StartHanging(taskid string) error {
	t.Lock()
	defer t.Unlock()
	task, ok := t.tasks[taskid]
	if !ok {
		return ErrRecoveryFailed
	}

	task.status = HangingStatus
	return nil
}

// Kill kill a task
func (t *TaskManager) Kill(taskid string) {
	_ = t.Recovery(taskid)
	t.Lock()
	defer t.Unlock()
	task, ok := t.tasks[taskid]
	if !ok {
		return
	}
	if task.thread != nil {
		task.thread.Cancel(fmt.Sprintf("cancel %s by task manager", task.ID))
	}
}

// GetAll 获取所有执行中的tasks快照
func (t *TaskManager) GetAll() map[string]*Task {
	t.Lock()
	defer t.Unlock()
	result := make(map[string]*Task, len(t.tasks))
	for k, v := range t.tasks {
		result[k] = v
	}
	return result
}

func GetTaskManager() *TaskManager {
	return NewTaskManager()
}
