package uuid

import (
	"testing"

	"go.starlark.net/starlark"
	"go.starlark.net/starlarktest"

	"github.com/superops-team/hyperops/pkg/ops/starlib/testdata"
)

func TestUUID(t *testing.T) {
	thread := &starlark.Thread{Load: testdata.NewModuleLoader(Module)}
	starlarktest.SetReporter(thread, t)

	// Execute test file
	_, err := starlark.ExecFile(thread, "testdata/test.star", nil, nil)
	if err != nil {
		t.Error(err)
	}
}
