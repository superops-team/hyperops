package context

import (
	"testing"

	"github.com/superops-team/hyperops/pkg/ops/starlib/testdata"
	"go.starlark.net/starlark"
)

func TestContext(t *testing.T) {
	thread := &starlark.Thread{Load: newLoader(), Print: SafePrint}

	// Execute test file
	_, err := starlark.ExecFile(thread, "testdata/test.star", nil, starlark.StringDict{
		"ctx": NewContext(
			map[string]interface{}{"foo": "bar"},
			map[string]string{"baz": "bat"},
		).Struct(),
	})
	if err != nil {
		t.Error(err)
	}
}

func TestMissingValue(t *testing.T) {
	thread := &starlark.Thread{}

	ctx := NewContext(nil, nil)
	val, err := ctx.getValue(thread, nil, starlark.Tuple{starlark.String("foo")}, nil)
	if val != starlark.None {
		t.Errorf("expected none return value")
	}

	expect := "value foo not set in context"
	if err.Error() != expect {
		t.Errorf("error message mismatch. expected: %s, got: %s", expect, err.Error())
	}
}

func TestMissingConfig(t *testing.T) {
	thread := &starlark.Thread{}
	ctx := NewContext(nil, nil)

	val, err := ctx.getConfig(thread, nil, starlark.Tuple{starlark.String("foo")}, nil)
	if val != starlark.None {
		t.Errorf("expected none return value")
	}

	expect := "no config provided"
	if err.Error() != expect {
		t.Errorf("error message mismatch. expected: %s, got: %s", expect, err.Error())
	}
}

func TestMissingSecrets(t *testing.T) {
	thread := &starlark.Thread{}
	ctx := NewContext(nil, nil)

	val, err := ctx.getSecret(thread, nil, starlark.Tuple{starlark.String("foo")}, nil)
	if val != starlark.None {
		t.Errorf("expected none return value")
	}

	expect := "no secrets provided"
	if err.Error() != expect {
		t.Errorf("error message mismatch. expected: %s, got: %s", expect, err.Error())
	}
}

func TestSetSecrets(t *testing.T) {
	thread := &starlark.Thread{}
	ctx := NewContext(nil, make(map[string]string))

	val, err := ctx.setSecret(thread, nil, starlark.Tuple{starlark.String("foo"), starlark.String("dao")}, nil)
	if val != starlark.None {
		t.Errorf("expected none return value")
	}

	if err != nil {
		t.Errorf("expected nil not err")
	}

	val, _ = ctx.getSecret(thread, nil, starlark.Tuple{starlark.String("foo")}, nil)
	if val != starlark.String("dao") {
		t.Errorf("expected none return value")
	}
}

// load implements the 'load' operation as used in the evaluator tests.
func newLoader() func(thread *starlark.Thread, module string) (starlark.StringDict, error) {
	return testdata.NewLoader(nil, "context_is_global_no_module_name_exists")
}
