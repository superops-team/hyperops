package localexec

import (
	"fmt"
	_ "os"
	"testing"
	"time"
)

func TestLocalexec(t *testing.T) {
	ret, err := ExecCmd(5*time.Second, "/", "ls", "-a")
	fmt.Println(ret.Stdout)
	if err != nil {
		t.Errorf(err.Error())
	}

}

func TestExecRealTime(t *testing.T) {
	dir := ""
	res, err := ExecBatchCmdS(10*time.Second, dir, "sleep 1 && echo hello&&sleep 1")
	if err != nil {
		t.Error(err)
	}
	fmt.Printf("res: %v\n", res)
}

func TestExecRealTimeErr(t *testing.T) {
	dir := ""
	res, err := ExecBatchCmdS(10*time.Second, dir, "echo 124 && exit 1")
	if err != nil {
		t.Error(err)
	}
	if res.Code != 1 {
		t.Error(err)
	}
	fmt.Printf("res: %v\n", res)
}

func TestExecRealTimeCmdNotFound(t *testing.T) {
	dir := ""
	res, err := ExecBatchCmdS(10*time.Second, dir, "echoxjlkjk")
	if err != nil {
		t.Error(err)
	}
	if res.Code != 127 {
		t.Error(err)
	}
	if res.Stderr == "" {
		t.Error(err)
	}
	fmt.Printf("res: %v\n", res)
}

func TestExecWithNilConfig(t *testing.T) {
	dir := ""
	res, err := ExecBatchCmdS(10*time.Second, dir, "sleep 1 && echo hello&&sleep 1")
	if err != nil {
		t.Error(err)
	}
	fmt.Printf("res: %v\n", res)
}
