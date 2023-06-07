package context

import (
	"fmt"
	"strings"

	"github.com/superops-team/hyperops/pkg/ops/util"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
)

var (
	ErrFuncNotFound   = fmt.Errorf("the funcation in mod not found")
	ErrNotAFunc       = fmt.Errorf("only funcation can be called")
	ErrInvalidArgType = fmt.Errorf("invalid argument type provided to method")
)

//thread 表示执行该命令的starlark线程
//dict 表示 需要被执行的函数所在的dict 可以通过starlib.Loader或者ops.predecleared获取
//funcinfo 为执行函数的名称,形式为"mod.func",例如:"sh.exec",
//需要保证dict中存在有mod这个key,而对应的value是一个starlarkstruct.Module类型，且该module的members中含有这个function,并且保证传入的dict中包含有这个键值对
//args和kwargs为func所需要的参数，该函数会自行将其转化为starlark.Value类型
func Call(thread *starlark.Thread, dict starlark.StringDict, funcinfo string, args []interface{}, kwargs map[string]interface{}) (interface{}, error) {
	var (
		mod        string
		starArgs   starlark.Tuple
		starKwargs []starlark.Tuple
		err        error
		builtin    *starlark.Builtin
	)
	slice := strings.Split(funcinfo, ".")
	if len(slice) != 2 {
		return nil, fmt.Errorf("got invalid funcinfo,usage mod.func")
	}
	mod = slice[0]

	v, ok := dict[mod]
	if !ok {
		return nil, ErrFuncNotFound
	}
	switch funcList := v.(type) {
	case *starlarkstruct.Module:
		f := funcList.Members
		modfunc := f[funcinfo[len(mod)+1:]]
		if modfunc == nil {
			return nil, fmt.Errorf("func %s not found", funcinfo[len(mod)+1:])
		}
		builtin, ok = modfunc.(*starlark.Builtin)
		if !ok {
			return nil, ErrNotAFunc
		}
	case *starlarkstruct.Struct:
		res, err := funcList.Attr(funcinfo[len(mod)+1:])
		if err != nil {
			return nil, err
		}
		builtin = res.(*starlark.Builtin)
	}

	if args != nil {
		starArgs, err = GetArgs(args)
		if err != nil {
			return nil, err
		}
	}

	if kwargs != nil {
		starKwargs, err = GetKwargs(kwargs)
		if err != nil {
			return nil, err
		}
	}

	res, err := builtin.CallInternal(thread, starArgs, starKwargs)
	if err != nil {
		return nil, err
	}
	return util.Unmarshal(res)
}

func GetKwargs(kwargs map[string]interface{}) ([]starlark.Tuple, error) {
	a, err := util.Marshal(kwargs)
	if err != nil {
		return nil, err
	}
	dic, ok := a.(*starlark.Dict)
	if !ok {
		return nil, ErrInvalidArgType
	}

	return dic.Items(), nil
}

func GetArgs(args interface{}) (starlark.Tuple, error) {
	a, err := util.Marshal(args)
	if err != nil {
		return nil, err
	}
	lis, ok := a.(*starlark.List)
	if !ok {
		return nil, fmt.Errorf("got invaliid type in args,please use []interafce{} type")
	}
	var starArgs starlark.Tuple
	for i := 0; i < lis.Len(); i++ {
		starArgs = append(starArgs, lis.Index(i))
	}
	return starArgs, nil
}
