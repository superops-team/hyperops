package util

import (
	"fmt"
	"reflect"

	"github.com/pkg/errors"
	"go.starlark.net/starlark"
)

// ConvertToStarlark provides an exported function for converting starlark values
func ConvertToStarlark(value interface{}) (starlark.Value, error) {
	return convertToStarlark(value)
}

func convertToStarlark(value interface{}) (starlark.Value, error) {
	if value == nil {
		return starlark.None, nil
	}
	switch v := value.(type) {
	case starlark.Value:
		return v, nil
	case errTuple:
		var starlarkErr starlark.Value = starlark.None
		if v.err != nil {
			starlarkErr = starlark.String(v.err.Error())
		}
		retVal, err := convertToStarlark(v.val)
		if err != nil {
			return starlark.None, fmt.Errorf("failed to convert contained value: %w", err)
		}
		return starlark.Tuple{retVal, starlarkErr}, nil
	case bool:
		return starlark.Bool(v), nil
	case int:
		return starlark.MakeInt(v), nil
	case int64:
		return starlark.MakeInt64(v), nil
	case uint:
		return starlark.MakeUint(v), nil
	case uint64:
		return starlark.MakeUint64(v), nil
	case float32:
		return starlark.Float(v), nil
	case float64:
		return starlark.Float(v), nil
	case string:
		return starlark.String(v), nil
	case error:
		return starlark.String(v.Error()), nil
	default:
		reflectV := reflect.ValueOf(value)
		switch reflectV.Kind() {
		case reflect.Slice:
			var elems []starlark.Value
			for i := 0; i < reflectV.Len(); i++ {
				val, err := convertToStarlark(reflectV.Index(i).Interface())
				if err != nil {
					return nil, errors.Wrapf(err, "failed to convert slice element %d", i)
				}
				elems = append(elems, val)
			}
			return starlark.NewList(elems), nil
		case reflect.Map:
			dict := starlark.NewDict(len(reflectV.MapKeys()))

			iter := reflectV.MapRange()
			for iter.Next() {
				key, err := convertToStarlark(iter.Key().Interface())
				if err != nil {
					return nil, errors.Wrapf(err, "failed to convert map key %s", key.String())
				}

				val, err := convertToStarlark(iter.Value().Interface())
				if err != nil {
					return nil, errors.Wrapf(err, "failed to convert map val %s: %s", key.String(), val.String())
				}
				if err = dict.SetKey(key, val); err != nil {
					return nil, err
				}
			}
			return dict, nil
		}
	}
	return nil, fmt.Errorf("cannot convert value %#v", value)
}
