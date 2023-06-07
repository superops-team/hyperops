package context

import (
	"fmt"
	"testing"
	"time"

	"github.com/superops-team/hyperops/pkg/localexec"
	"github.com/superops-team/hyperops/pkg/ops/util"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
)

var Module = &starlarkstruct.Module{
	Name: "shell",
	Members: starlark.StringDict{
		// "exec": starlark.NewBuiltin("shell.exec", Exec),
		"exec": AddBuiltin("shell.exec", Run),
	},
}

func TestCall(t *testing.T) {
	dict := starlark.StringDict{"shell": Module}
	res, err := Call(&starlark.Thread{}, dict, "shell.exec", []interface{}{"pwd"}, nil)
	if err != nil {
		t.Error(err)
	}
	fmt.Printf("res: %v\n", res)
}

func Run(thread *starlark.Thread, _ *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	params, err := util.GetParser(args, kwargs)
	if err != nil {
		return starlark.None, err
	}
	timeout, err := params.GetIntByName("timeout")
	if err != nil || timeout == 0 {
		timeout = 10
	}
	cmd, err := params.GetStringByName("cmd")
	if err != nil {
		// 如果获取失败默认第一个参数为cmd
		cmd, err = params.GetString(0)
		if err != nil {
			return starlark.None, err
		}
	}

	response, err := localexec.ExecBatchCmdS(time.Duration(timeout)*time.Second, "", cmd, nil)
	if response == nil && err != nil {
		return starlark.None, err
	}
	return starlarkstruct.FromStringDict(starlarkstruct.Default, starlark.StringDict{
		"code":   starlark.MakeInt(int(response.Code)),
		"stdout": starlark.String(response.Stdout),
		"stderr": starlark.String(response.Stderr),
	}), nil
}
