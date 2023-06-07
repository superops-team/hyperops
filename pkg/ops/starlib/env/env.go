package env

import (
	"github.com/superops-team/hyperops/pkg/environment"
	"github.com/superops-team/hyperops/pkg/ops/util"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
)

const Name = "env"
const ModuleName = "env.star"

var Module = &starlarkstruct.Module{
	Name: "env",
	Members: starlark.StringDict{
		"get": starlark.NewBuiltin("env.get", Get),
		"set": starlark.NewBuiltin("env.set", Set),
	},
}

// Get get env
func Get(thread *starlark.Thread, _ *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	params, err := util.GetParser(args, kwargs)
	if err != nil {
		return starlark.None, err
	}
	key, err := params.GetStringByName("key")
	if err != nil {
		key, err = params.GetString(0)
		if err != nil {
			return starlark.None, err
		}
	}

	e := environment.NewEnvStorage()

	val := e.Get(key)
	return starlark.String(val), nil
}

// Set set env
func Set(thread *starlark.Thread, _ *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	params, err := util.GetParser(args, kwargs)
	if err != nil {
		return starlark.None, err
	}
	key, err := params.GetStringByName("key")
	if err != nil {
		key, err = params.GetString(0)
		if err != nil {
			return starlark.None, err
		}
	}
	val, err := params.GetStringByName("val")
	if err != nil {
		val, err = params.GetString(1)
		if err != nil {
			return starlark.None, err
		}
	}

	e := environment.NewEnvStorage()

	e.Set(key, val)
	return starlark.None, nil
}
