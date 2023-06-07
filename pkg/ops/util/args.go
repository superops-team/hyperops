package util

import (
	"fmt"

	"go.starlark.net/starlark"
)

// ArgParser easy parse starlark args and kwargs
type ArgParserInterface interface {
	RestrictKwargs(kwargs ...string) error

	GetParam(index int) (starlark.Value, error)

	GetString(index int) (string, error)
	GetStringByName(name string) (string, error)

	GetInt(index int) (int64, error)
	GetIntByName(name string) (int64, error)

	GetBool(index int) (bool, error)
	GetBoolByName(kwarg string) (bool, error)
}

type ArgParser struct {
	ArgParserInterface

	args   map[int]starlark.Value
	kwargs map[string]int
}

func (parser *ArgParser) RestrictKwargs(kwargs ...string) error {
	for kwarg := range parser.kwargs {
		valid := false
		for _, validKwarg := range kwargs {
			if kwarg == validKwarg {
				valid = true
			}
		}
		if !valid {
			return fmt.Errorf("%w: %q", ErrInvalidKwarg, kwarg)
		}
		valid = false
	}
	return nil
}

func GetParser(args starlark.Tuple, kwargs []starlark.Tuple) (*ArgParser, error) {
	parser := &ArgParser{
		args:   map[int]starlark.Value{},
		kwargs: map[string]int{},
	}

	for i, arg := range args {
		parser.args[i] = arg
	}
	for _, kwarg := range kwargs {
		if kwarg.Len() != 2 {
			return nil, ErrMalformattedKwarg
		}
		name, ok := starlark.AsString(kwarg[0])
		if !ok {
			return nil, ErrMalformattedKwarg
		}
		val := kwarg[1]

		index := len(parser.args)
		parser.args[index] = val
		parser.kwargs[name] = index
	}

	return parser, nil
}

func (parser *ArgParser) GetParam(index int) (starlark.Value, error) {
	val, ok := parser.args[index]
	if !ok {
		return nil, ErrMissingArg
	}
	return val, nil
}

func (parser *ArgParser) GetParamIndex(kwarg string) (int, error) {
	index, ok := parser.kwargs[kwarg]
	if !ok {
		return 0, fmt.Errorf("%w: %q", ErrMissingKwarg, kwarg)
	}
	return index, nil
}

func (parser *ArgParser) GetString(index int) (string, error) {
	val, err := parser.GetParam(index)
	if err != nil {
		return "", err
	}

	str, ok := val.(starlark.String)
	if !ok {
		return "", ErrInvalidArgType
	}

	return str.GoString(), nil
}
func (parser *ArgParser) GetStringByName(kwarg string) (string, error) {
	index, err := parser.GetParamIndex(kwarg)
	if err != nil {
		return "", err
	}
	return parser.GetString(index)
}

func (parser *ArgParser) GetInt(index int) (int64, error) {
	val, err := parser.GetParam(index)
	if err != nil {
		return 0, err
	}

	intVal, ok := val.(starlark.Int)
	if !ok {
		return 0, ErrInvalidArgType
	}

	realInt, isRepresentable := intVal.Int64()
	if !isRepresentable {
		return 0, ErrInvalidArgType
	}
	return realInt, nil
}
func (parser *ArgParser) GetIntByName(kwarg string) (int64, error) {
	index, err := parser.GetParamIndex(kwarg)
	if err != nil {
		return 0, err
	}
	return parser.GetInt(index)
}

func (parser *ArgParser) GetBool(index int) (bool, error) {
	val, err := parser.GetParam(index)
	if err != nil {
		return false, err
	}

	boolVal, ok := val.(starlark.Bool)
	if !ok {
		return false, ErrInvalidArgType
	}

	return bool(boolVal.Truth()), nil
}
func (parser *ArgParser) GetBoolByName(kwarg string) (bool, error) {
	index, err := parser.GetParamIndex(kwarg)
	if err != nil {
		return false, err
	}
	return parser.GetBool(index)
}

func (parser *ArgParser) GetListByName(kwarg string) ([]interface{}, error) {
	index, err := parser.GetParamIndex(kwarg)
	if err != nil {
		return nil, err
	}
	return parser.GetList(index)
}

func (parser *ArgParser) GetList(index int) ([]interface{}, error) {
	val, err := parser.GetParam(index)
	if err != nil {
		return nil, err
	}
	var v interface{}
	switch listVal := val.(type) {
	case starlark.Tuple:
		v, err = Unmarshal(listVal)
		if err != nil {
			return nil, err
		}
	case *starlark.List:
		v, err = Unmarshal(listVal)
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("got unknown type :%T", listVal)
	}
	listV, ok := v.([]interface{})
	if !ok {
		return nil, fmt.Errorf("got unknown type :%T", v)
	}
	return listV, err
}
