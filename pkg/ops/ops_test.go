package ops

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	localctx "github.com/superops-team/hyperops/pkg/ops/context"
	"github.com/superops-team/hyperops/pkg/ops/event"
	"go.starlark.net/starlark"
)

func TestExecScriptWithEvent(t *testing.T) {
	ctx := context.Background()

	output := &bytes.Buffer{}

	eventCh := make(chan event.Event)
	done := make(chan error)

	go func() {
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
					fmt.Printf("%s -> %s\n", payload.From, payload.To)
					output.Write([]byte(fmt.Sprintf("%s -> %s\n", payload.From, payload.To)))
				}

			case err := <-done:
				if err != nil {
					fmt.Println(err.Error())
				}
				return
			}
		}
	}()

	err := ExecScript(
		ctx,
		&Target{
			ScriptPath: "test.star",
		},
		AddEventsChannel(eventCh),
		SetLocals(map[string]interface{}{
			"x": "test_local_var",
			"y": "2",
		}),
		SetSecrets(map[string]string{
			"password1": "1233455",
			"password2": "1233455",
		}),
	)
	done <- err
	close(eventCh)
	if err != nil {
		t.Error(err.Error())
		return
	}
	content, err := ioutil.ReadAll(output)
	if err != nil {
		t.Fatal(err)
	}
	expect := "hello world"
	if !strings.Contains(string(content), expect) {
		t.Errorf("output mismatch. expected: '%s', got: '%s'", expect, string(content))
	}
	if !strings.Contains(string(content), "******") {
		t.Errorf("output mismatch. expected: '%s', got: '%s'", "1233455", string(content))
	}
	if !strings.Contains(string(content), "test_local_var") {
		t.Errorf("output mismatch. expected: '%s', got: '%s'", "test_local_var", string(content))
	}
}

func TestExecScript(t *testing.T) {
	ctx := context.Background()
	output := &bytes.Buffer{}
	err := ExecScript(
		ctx,
		&Target{
			ScriptPath: "test.star",
		},
		SetOutputWriter(output),
		SetLocals(map[string]interface{}{
			"x":                       "test_local_var",
			"y":                       "2",
			"HYPEROPS_WORKSPACE_KEEP": false,
		}),
		SetSecrets(map[string]string{
			"password1": "1233455",
			"password2": "1233455",
		}),
	)
	if err != nil {
		t.Error(err.Error())
		return
	}
	content, err := ioutil.ReadAll(output)
	if err != nil {
		t.Fatal(err)
	}
	expect := "hello world"
	if !strings.Contains(string(content), expect) {
		t.Errorf("output mismatch. expected: '%s', got: '%s'", expect, string(content))
	}
	if !strings.Contains(string(content), "******") {
		t.Errorf("output mismatch. expected: '%s', got: '%s'", "1233455", string(content))
	}
	if !strings.Contains(string(content), "test_local_var") {
		t.Errorf("output mismatch. expected: '%s', got: '%s'", "test_local_var", string(content))
	}
}

func TestExec(t *testing.T) {
	ctx := context.Background()
	outputBuffer := &bytes.Buffer{}
	err := ExecScript(
		ctx,
		&Target{
			ScriptContent: []byte(`
print("hello world")
print(ctx.get_config("x"))
print(ctx.get_secret("password1"))
			`),
		},
		SetOutputWriter(outputBuffer),
		SetLocals(map[string]interface{}{
			"x": "test_local_var1",
			"y": "2",
		}),
		SetSecrets(map[string]string{
			"password1": "1233455",
			"password2": "1233455",
		}),
	)
	if err != nil {
		t.Error(err.Error())
		return
	}
	output, err := ioutil.ReadAll(outputBuffer)
	if err != nil {
		t.Fatal(err)
	}
	expect := "hello world"
	if !strings.Contains(string(output), expect) {
		t.Errorf("output mismatch. expected: '%s', got: '%s'", expect, string(output))
	}

	if !strings.Contains(string(output), "******") {
		t.Errorf("output mismatch. expected: '%s', got: '%s'", "1233455", string(output))
	}
	if !strings.Contains(string(output), "test_local_var1") {
		t.Errorf("output mismatch. expected: '%s', got: '%s'", "test_local_var1", string(output))
	}
}

func TestExecWithHang(t *testing.T) {
	go func() {
		http.Handle("/metrics", promhttp.Handler())
		fmt.Println("listen failed")
		//然后访问http://localhost:8889/metrics
	}()
	tm := localctx.NewTaskManager()
	thread := &starlark.Thread{}
	taskid := "test"
	thread.Name = taskid
	ctx := context.Background()
	tm.Add(taskid, thread, nil)
	err := tm.Suspend(taskid)
	if err != nil {
		t.Error(err)
	}
	time.Sleep(5 * time.Second)
	fmt.Println("hang ended")
	err = tm.Recovery(taskid)
	if err != nil {
		t.Error(err)
	}
	err = tm.RecoveryOver(taskid)
	if err != nil {
		t.Error(err)
	}
	outputBuffer := &bytes.Buffer{}
	err = ExecScript(
		ctx,
		&Target{
			ScriptContent: []byte(`
print("hello world")
print(ctx.get_config("x"))
print(ctx.get_secret("password1"))
			`),
		},
		SetOutputWriter(outputBuffer),
		SetLocals(map[string]interface{}{
			"x": "test_local_var1",
			"y": "2",
		}),
		SetSecrets(map[string]string{
			"password1": "1233455",
			"password2": "1233455",
		}),
	)
	if err != nil {
		t.Error(err.Error())
		return
	}
	output, err := ioutil.ReadAll(outputBuffer)
	if err != nil {
		t.Fatal(err)
	}
	expect := "hello world"
	if !strings.Contains(string(output), expect) {
		t.Errorf("output mismatch. expected: '%s', got: '%s'", expect, string(output))
	}

	if !strings.Contains(string(output), "******") {
		t.Errorf("output mismatch. expected: '%s', got: '%s'", "1233455", string(output))
	}
	if !strings.Contains(string(output), "test_local_var1") {
		t.Errorf("output mismatch. expected: '%s', got: '%s'", "test_local_var1", string(output))
	}
	time.Sleep(30 * time.Second)
}
