package metric

import (
	"testing"

	"github.com/superops-team/hyperops/pkg/ops/starlib/testdata"
	"github.com/superops-team/hyperops/pkg/ops/starlib/time"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarktest"
)

func TestMetric(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}
	thread := &starlark.Thread{Load: testdata.NewModuleLoader(Module, time.Module)}
	starlarktest.SetReporter(thread, t)
	_, err := starlark.ExecFile(thread, "testdata/metric.star", nil, nil)
	if err != nil {
		t.Error(err)
	}
}
