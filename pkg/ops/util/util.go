package util

import (
	"fmt"
	"os"
	"path"
	"strconv"
	"time"

	"github.com/superops-team/hyperops/pkg/environment"
	startime "go.starlark.net/lib/time"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
)

// AsString unquotes a starlark string value
func AsString(x starlark.Value) (string, error) {
	return strconv.Unquote(x.String())
}

// IsEmptyString checks is a starlark string is empty ("" for a go string)
// starlark.String.String performs repr-style quotation, which is necessary
// for the starlark.Value contract but a frequent source of errors in API
// clients. This helper method makes sure it'll work properly
func IsEmptyString(s starlark.String) bool {
	return s.String() == `""`
}

// Unmarshal decodes a starlark.Value into it's golang counterpart
func Unmarshal(x starlark.Value) (val interface{}, err error) {
	switch v := x.(type) {
	case starlark.NoneType:
		val = nil
	case starlark.Bool:
		val = v.Truth() == starlark.True
	case starlark.Int:
		var tmp int
		err = starlark.AsInt(x, &tmp)
		val = tmp
	case starlark.Float:
		if f, ok := starlark.AsFloat(x); !ok {
			err = fmt.Errorf("couldn't parse float")
		} else {
			val = f
		}
	case starlark.String:
		val = v.GoString()
	case startime.Time:
		val = time.Time(v)
	case *starlark.Dict:
		var (
			dictVal starlark.Value
			pval    interface{}
			kval    interface{}
			keys    []interface{}
			vals    []interface{}
			// key as interface if found one key is not a string
			ki bool
		)

		for _, k := range v.Keys() {
			dictVal, _, err = v.Get(k)
			if err != nil {
				return
			}

			pval, err = Unmarshal(dictVal)
			if err != nil {
				err = fmt.Errorf("unmarshaling starlark value: %w", err)
				return
			}

			kval, err = Unmarshal(k)
			if err != nil {
				err = fmt.Errorf("unmarshaling starlark key: %w", err)
				return
			}

			if _, ok := kval.(string); !ok {
				// found key as not a string
				ki = true
			}

			keys = append(keys, kval)
			vals = append(vals, pval)
		}

		// prepare result

		rs := map[string]interface{}{}
		ri := map[interface{}]interface{}{}

		for i, key := range keys {
			// key as interface
			if ki {
				ri[key] = vals[i]
			} else {
				rs[key.(string)] = vals[i]
			}
		}

		if ki {
			val = ri // map[interface{}]interface{}
		} else {
			val = rs // map[string]interface{}
		}
	case *starlark.List:
		var (
			i       int
			listVal starlark.Value
			iter    = v.Iterate()
			value   = make([]interface{}, v.Len())
		)

		defer iter.Done()
		for iter.Next(&listVal) {
			value[i], err = Unmarshal(listVal)
			if err != nil {
				return
			}
			i++
		}
		val = value
	case starlark.Tuple:
		var (
			i        int
			tupleVal starlark.Value
			iter     = v.Iterate()
			value    = make([]interface{}, v.Len())
		)

		defer iter.Done()
		for iter.Next(&tupleVal) {
			value[i], err = Unmarshal(tupleVal)
			if err != nil {
				return
			}
			i++
		}
		val = value
	case *starlark.Set:
		fmt.Println("errnotdone: SET")
		err = fmt.Errorf("sets aren't yet supported")
	case *starlarkstruct.Struct:
		dict := make(map[string]interface{})
		for _, n := range v.AttrNames() {
			tmp, _ := v.Attr(n)
			if tmp != nil {
				dict[n], err = Unmarshal(tmp)
				if err != nil {
					return
				}
			}
		}
		val = dict

	default:
		fmt.Println("errbadtype:", x.Type())
		err = fmt.Errorf("unrecognized starlark type: %s", x.Type())
	}
	return
}

// Marshal turns go values into starlark types
func Marshal(data interface{}) (v starlark.Value, err error) {
	switch x := data.(type) {
	case nil:
		v = starlark.None
	case bool:
		v = starlark.Bool(x)
	case string:
		v = starlark.String(x)
	case int:
		v = starlark.MakeInt(x)
	case int8:
		v = starlark.MakeInt(int(x))
	case int16:
		v = starlark.MakeInt(int(x))
	case int32:
		v = starlark.MakeInt(int(x))
	case int64:
		v = starlark.MakeInt64(x)
	case uint:
		v = starlark.MakeUint(x)
	case uint8:
		v = starlark.MakeUint(uint(x))
	case uint16:
		v = starlark.MakeUint(uint(x))
	case uint32:
		v = starlark.MakeUint(uint(x))
	case uint64:
		v = starlark.MakeUint64(x)
	case float32:
		v = starlark.Float(float64(x))
	case float64:
		v = starlark.Float(x)
	case time.Time:
		v = startime.Time(x)
	case []string:
		tem := starlark.Tuple{}
		for _, s := range x {
			tem = append(tem, starlark.String(s))
		}
		v = tem
	case []interface{}:
		var elems = make([]starlark.Value, len(x))
		for i, val := range x {
			elems[i], err = Marshal(val)
			if err != nil {
				return
			}
		}
		v = starlark.NewList(elems)
	case map[interface{}]interface{}:
		dict := &starlark.Dict{}
		var elem starlark.Value
		for ki, val := range x {
			var key starlark.Value
			key, err = Marshal(ki)
			if err != nil {
				return
			}
			elem, err = Marshal(val)
			if err != nil {
				return
			}
			if err = dict.SetKey(key, elem); err != nil {
				return
			}
		}
		v = dict
	case map[string]interface{}:
		dict := &starlark.Dict{}
		var elem starlark.Value
		for key, val := range x {
			elem, err = Marshal(val)
			if err != nil {
				return
			}
			if err = dict.SetKey(starlark.String(key), elem); err != nil {
				return
			}
		}
		v = dict
	case Marshaler:
		v, err = x.MarshalStarlark()
	default:
		return starlark.None, fmt.Errorf("unrecognized type: %#v", x)
	}
	return
}

// Unmarshaler is the interface use to unmarshal starlark custom types.
type Unmarshaler interface {
	// UnmarshalStarlark unmarshal a starlark object to custom type.
	UnmarshalStarlark(starlark.Value) error
}

// Marshaler is the interface use to marshal starlark custom types.
type Marshaler interface {
	// MarshalStarlark marshal a custom type to starlark object.
	MarshalStarlark() (starlark.Value, error)
}

func EnsureWorkdir(jobid string) string {
	env := environment.NewEnvStorage()
	pwdpath := env.Get("PWD")
	workdir := "./"
	if len(pwdpath) != 0 && len(jobid) != 0 {
		workdir = path.Join(pwdpath, jobid)
		if _, err := os.Stat(workdir); os.IsNotExist(err) {
			if err := os.MkdirAll(workdir, 0744); err != nil {
				return "./"
			}
		}
	}
	return workdir
}
