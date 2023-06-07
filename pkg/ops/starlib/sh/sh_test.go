package sh

import (
	"testing"

	"github.com/superops-team/hyperops/pkg/ops/starlib/testdata"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarktest"
)

func TestNewModule(t *testing.T) {
	thread := &starlark.Thread{Load: testdata.NewModuleLoader(Module)}
	starlarktest.SetReporter(thread, t)

	// Execute test file
	_, err := starlark.ExecFile(thread, "testdata/exec.star", nil, nil)
	if err != nil {
		t.Error(err)
	}
}
