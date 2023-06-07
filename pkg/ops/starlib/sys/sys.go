package sys

import (
	"fmt"
	"os"
	"runtime"

	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
)

const Name = "sys"
const ModuleName = "sys.star"

var Module = &starlarkstruct.Module{
	Name: "sys",
	Members: starlark.StringDict{
		"os":         starlark.String(runtime.GOOS),
		"arch":       starlark.String(runtime.GOARCH),
		"platform":   starlark.String(fmt.Sprintf("%s_%s", runtime.GOOS, runtime.GOARCH)),
		"argv":       argv(),
		"executable": executable(),
	},
}

// List of commandline arguments that Tilt started with.
func argv() starlark.Value {
	values := []starlark.Value{}
	for _, arg := range os.Args {
		values = append(values, starlark.String(arg))
	}

	list := starlark.NewList(values)
	list.Freeze()
	return list
}

// Full path to the hyperops executable.
func executable() starlark.Value {
	e, err := os.Executable()
	if err != nil {
		return starlark.None
	}
	return starlark.String(e)
}
