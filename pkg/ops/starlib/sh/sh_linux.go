package sh

import (
	"os"
	"path"
	"time"

	"github.com/superops-team/hyperops/pkg/environment"
	"github.com/superops-team/hyperops/pkg/localexec"
	"github.com/superops-team/hyperops/pkg/ops/util"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
)

const Name = "shell"
const ModuleName = "shell.star"

var Module = &starlarkstruct.Module{
	Name: "shell",
	Members: starlark.StringDict{
		// "exec": starlark.NewBuiltin("shell.exec", Exec),
		"exec": starlark.NewBuiltin("shell.exec", Exec),
	},
}

func ensureWorkdir(jobid string) string {
	env := environment.NewEnvStorage()
	pwdpath := env.Get("PWD")
	workdir := "./"
	if len(pwdpath) != 0 && len(jobid) != 0 {
		workdir = path.Join(pwdpath, jobid)
		if err := os.MkdirAll(workdir, 0744); err != nil {
			return "./"
		}
	}
	return workdir
}

// Exec run local command in starlark
func Exec(thread *starlark.Thread, _ *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	params, err := util.GetParser(args, kwargs)
	if err != nil {
		return starlark.None, err
	}
	timeout, err := params.GetIntByName("timeout")
	if err != nil || timeout == 0 {
		timeout = 10
	}
	
	dir, err := params.GetStringByName("dir")
	if err != nil {
		if len(thread.Name) != 0 {
			dir = ensureWorkdir(thread.Name)
		} else {
			dir = "./"
		}
	}
	cmd, err := params.GetStringByName("cmd")
	if err != nil {
		// 如果获取失败默认第一个参数为cmd
		cmd, err = params.GetString(0)
		if err != nil {
			return starlark.None, err
		}
	}
	response, err := localexec.ExecBatchCmdS(time.Duration(timeout)*time.Second, dir, cmd)
	if response == nil && err != nil {
		return starlark.None, err
	}
	return starlarkstruct.FromStringDict(starlarkstruct.Default, starlark.StringDict{
		"code":   starlark.MakeInt(int(response.Code)),
		"stdout": starlark.String(response.Stdout),
		"stderr": starlark.String(response.Stderr),
	}), nil
}
