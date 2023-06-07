package ops

import (
	"fmt"
	"time"

	"go.starlark.net/starlark"
)

// SleepFn implements built-in for sleep.
func SleepFn(t *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var dur string
	if err := starlark.UnpackPositionalArgs(b.Name(), args, kwargs, 1, &dur); err != nil {
		return nil, err
	}

	d, err := time.ParseDuration(dur)
	if err != nil {
		return nil, fmt.Errorf("<%v>: can not parse duration string `%s': %v", b.Name(), dur, err)
	}

	time.Sleep(d)

	return starlark.None, nil
}
