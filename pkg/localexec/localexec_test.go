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

/*
func TestExecBashCmdsMemoryLimit(t *testing.T) {
	var size = int64(1024 * 1024 * 50)
	//in localexec dir
	dir := ""
	//dd if=/dev/zero of=50M.file bs=1M count=50 in dir: testdata
	f, err := os.Create("50M.file")
	if err != nil {
		t.Errorf(err.Error())
	}
	defer f.Close()
	if err := f.Truncate(size); err != nil {
		t.Errorf(err.Error())
	}
	cmd := "cat 50M.file"
	res, err := ExecBatchCmdS(10*time.Second, dir, cmd, &Conf{
		Memory: 1,
		Cpu:    60,
		Name:   "hyperops",
	})
	fmt.Println(res.Code)
	if err != nil {
		t.Errorf(err.Error())
	}
	defer os.Remove("50M.file")
}

func TestExecBashCmdsCpuLimit(t *testing.T) {
	dir := ""
	_, err := ExecBatchCmdS(10*time.Second, dir, "cat /dev/zero > /dev/null", &Conf{
		Memory: 100,
		Cpu:    60,
		Name:   "hyperops",
	})
	if err != nil {
		t.Error(err)
	}
}
*/

func TestExecRealTime(t *testing.T) {
	dir := ""
	res, err := ExecBatchCmdS(10*time.Second, dir, "sleep 1 && echo hello&&sleep 1", &Conf{
		IsPrint: true,
	})
	if err != nil {
		t.Error(err)
	}
	fmt.Printf("res: %v\n", res)
}

func TestExecRealTimeErr(t *testing.T) {
	dir := ""
	res, err := ExecBatchCmdS(10*time.Second, dir, "echo 124 && exit 1", &Conf{
		IsPrint: true,
	})
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
	res, err := ExecBatchCmdS(10*time.Second, dir, "echoxjlkjk", &Conf{
		IsPrint: true,
	})
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

/*
func TestExecRealTimeWhileError(t *testing.T) {
	dir := ""
	res, err := ExecBatchCmdS(10*time.Second, dir, "sleep 1 && ec hello&&sleep 1", &Conf{
		Memory:  100,
		Cpu:     60,
		Name:    "hyperops",
		IsPrint: true,
	})
	if err != nil {
		t.Error(err)
	}
	fmt.Printf("res: %v\n", res)
}
*/

func TestExecWithNilConfig(t *testing.T) {
	dir := ""
	res, err := ExecBatchCmdS(10*time.Second, dir, "sleep 1 && echo hello&&sleep 1", nil)
	if err != nil {
		t.Error(err)
	}
	fmt.Printf("res: %v\n", res)
}
