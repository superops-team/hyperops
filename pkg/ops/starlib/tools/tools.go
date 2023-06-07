package tools

import (
	"strings"

	localctx "github.com/superops-team/hyperops/pkg/ops/context"
	"github.com/superops-team/hyperops/pkg/ops/util"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
)

const Name = "tools"
const ModuleName = "tools.star"

var Module = &starlarkstruct.Module{
	Name: "tools",
	Members: starlark.StringDict{
		"diff": localctx.AddBuiltin("tools.diff", DiffOfStr),
	},
}

func DiffOfStr(thread *starlark.Thread, _ *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var (
		origin string
		now    string
		err    error
	)

	params, err := util.GetParser(args, kwargs)
	if err != nil {
		return starlark.None, err
	}
	origin, err = params.GetString(0)
	if err != nil {
		return starlark.None, err
	}
	now, err = params.GetString(1)
	if err != nil {
		return starlark.None, err
	}
	originArr := strings.Split(origin, "\n")
	nowArr := strings.Split(now, "\n")

	diff := PPDiff(originArr, nowArr)

	return starlark.String(diff), nil
}
