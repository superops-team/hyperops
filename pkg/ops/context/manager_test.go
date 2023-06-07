package context

import (
	"bytes"
	"fmt"
	"math/rand"
	"sync"
	"testing"
	"time"

	"github.com/superops-team/hyperops/pkg/ops/event"
	"github.com/superops-team/hyperops/pkg/ops/starlib/testdata"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
	"go.starlark.net/starlarktest"
)

var (
	eventCh chan event.Event
)

func run(thread *starlark.Thread, _ *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	rand.Seed(time.Now().Unix())
	fmt.Println("begin ----------->")
	time.Sleep(time.Duration(rand.Intn(5)+1) * time.Second)
	fmt.Println("finished ----------->")
	return starlark.None, nil
}

func hyperopsPrint(thread *starlark.Thread, msg string) {
	sm := NewSecretsManager()
	safeMsg := sm.SafeReplace(msg)

	// 提供流式实时数据到外部使用者，保证执行过程会同步打印
	if eventCh != nil {
		payload := event.PrintEvent{
			ID:  thread.Name,
			Msg: safeMsg,
		}
		eventCh <- event.MakeEvent(event.ETPrint, thread.Name, payload)
	}
}

type Arr struct {
	sync.Mutex
	arr []string
	len int
}

func (a *Arr) Add(item string) {

	a.Lock()
	defer a.Unlock()
	a.arr = append(a.arr, item)
	a.len += 1
}

func dbUpdateThread(eventCh chan event.Event, done chan error, dbArr *[]string) {
	output := &bytes.Buffer{}

	for {
		select {
		case ev := <-eventCh:
			if ev.Type == event.ETPrint {
				payload := ev.Payload.(event.PrintEvent)
				fmt.Println(payload.Msg)
				output.Write([]byte(payload.Msg))
			}
			if ev.Type == event.ETTask {
				payload := ev.Payload.(event.TaskEvent)
				fmt.Printf(" %s -> %s \n", payload.From, payload.To)
				output.Write([]byte(fmt.Sprintf(" %s -> %s\n", payload.From, payload.To)))
				*dbArr = append(*dbArr, payload.To)
			}
		case err := <-done:
			if err != nil {
				fmt.Println(err.Error())
			}
			return
		}
	}

}

func suspendOrRecovryThread(tm *TaskManager, taskid string, retArr *Arr) {
	for i := 1; i < 100; i++ {
		time.Sleep(1 * time.Second)
		flag := rand.Intn(2)
		if flag == 1 {
			err := tm.Suspend(taskid)
			if err == nil {
				retArr.Add("hanging")
			}
		} else {
			err := tm.Recovery(taskid)
			if err == nil {
				retArr.Add("running")
			}
		}
	}
}

func killThread(tm *TaskManager, taskid string, retArr *Arr) {
	for i := 1; i < 100; i++ {
		time.Sleep(1 * time.Second)
		index := rand.Intn(12) //1/6的概率结束
		if index >= 0 && index <= 4 {
			err := tm.Suspend(taskid)
			if err == nil {
				retArr.Add("hanging")
			}
		} else if index >= 5 && index <= 10 {
			err := tm.Recovery(taskid)
			if err == nil {
				retArr.Add("running")
			}
		} else {
			fmt.Println("Kill !!!!!!!!!!")
			tm.Kill(taskid)
		}

	}
}

func TestTaskManagerSuspendRecoveryWithSingleOrder(t *testing.T) {
	var module = &starlarkstruct.Module{
		Name: "sh",
		Members: starlark.StringDict{
			"run": AddBuiltin("sh.run", run),
		},
	}

	thread := &starlark.Thread{Load: testdata.NewModuleLoader(module), Print: hyperopsPrint}
	starlarktest.SetReporter(thread, t)

	taskid := "123456"
	thread.Name = taskid

	eventCh = make(chan event.Event)
	done := make(chan error)

	dbArr := []string{}
	retArr := &Arr{
		arr: []string{},
		len: 0,
	}

	go dbUpdateThread(eventCh, done, &dbArr)

	tm := NewTaskManager()
	tm.Add(taskid, thread, eventCh)

	go suspendOrRecovryThread(tm, taskid, retArr)
	go suspendOrRecovryThread(tm, taskid, retArr)
	go suspendOrRecovryThread(tm, taskid, retArr)
	go suspendOrRecovryThread(tm, taskid, retArr)
	go suspendOrRecovryThread(tm, taskid, retArr)

	_, err := starlark.ExecFile(thread, "testdata/exec_with_hang_with_single_order.star", nil, nil)
	tm.Delete(taskid, nil)

	done <- err

	if err != nil {
		t.Error(err)
	}

}

func TestTaskManagerSuspendRecovery(t *testing.T) {
	var module = &starlarkstruct.Module{
		Name: "sh",
		Members: starlark.StringDict{
			"run": AddBuiltin("sh.run", run),
		},
	}

	thread := &starlark.Thread{Load: testdata.NewModuleLoader(module), Print: hyperopsPrint}
	starlarktest.SetReporter(thread, t)

	taskid := "12345"
	thread.Name = taskid

	eventCh = make(chan event.Event, 2)
	done := make(chan error)

	dbArr := []string{}
	retArr := &Arr{
		arr: []string{},
		len: 0,
	}

	go dbUpdateThread(eventCh, done, &dbArr)

	tm := NewTaskManager()
	tm.Add(taskid, thread, eventCh)

	go suspendOrRecovryThread(tm, taskid, retArr)
	go suspendOrRecovryThread(tm, taskid, retArr)
	go suspendOrRecovryThread(tm, taskid, retArr)
	go suspendOrRecovryThread(tm, taskid, retArr)
	go suspendOrRecovryThread(tm, taskid, retArr)

	_, err := starlark.ExecFile(thread, "testdata/exec_with_hang.star", nil, nil)
	tm.Delete(taskid, nil)
	done <- err

	if err != nil {
		t.Error(err)
	}

	fmt.Println(retArr.arr)
	fmt.Println(dbArr)

	for index := 0; index < len(retArr.arr)-2; index++ {
		if retArr.arr[index] != dbArr[index+1] {
			t.Error("Exec error!!!")
			return
		}
	}

}

func TestTaskManagerKill(t *testing.T) {
	var module = &starlarkstruct.Module{
		Name: "sh",
		Members: starlark.StringDict{
			"run": AddBuiltin("sh.run", run),
		},
	}

	thread := &starlark.Thread{Load: testdata.NewModuleLoader(module), Print: hyperopsPrint}
	starlarktest.SetReporter(thread, t)

	taskid := "1234"
	thread.Name = taskid

	eventCh = make(chan event.Event, 2)
	done := make(chan error)

	dbArr := []string{}
	retArr := &Arr{
		arr: []string{},
		len: 0,
	}

	go dbUpdateThread(eventCh, done, &dbArr)

	tm := NewTaskManager()
	tm.Add(taskid, thread, eventCh)

	go killThread(tm, taskid, retArr)
	//go killThread(tm, taskid, retArr)
	//go killThread(tm, taskid, retArr)

	_, err := starlark.ExecFile(thread, "testdata/exec_with_hang.star", nil, nil)
	tm.Delete(taskid, nil)

	done <- err

	if err != nil {
		fmt.Println("Kill SUCCESS")
	}

	fmt.Println(retArr.arr)
	fmt.Println(dbArr)

	for index := 0; index < len(retArr.arr)-2; index++ {
		if retArr.arr[index] != dbArr[index+1] {
			t.Error("Exec error!!!")
			return
		}
	}
}
