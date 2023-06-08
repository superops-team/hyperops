package zipfile

import (
	"io/ioutil"
	"testing"

	"github.com/superops-team/hyperops/pkg/ops/starlib/testdata"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarktest"
)

func TestFile(t *testing.T) {
	thread := &starlark.Thread{Load: testdata.NewModuleLoader(Module)}
	starlarktest.SetReporter(thread, t)


	zipBytes, err := ioutil.ReadFile("testdata/hello_world.zip")
	if err != nil {
		t.Fatal(err)
	}

	// Execute test file
	_, err = starlark.ExecFile(thread, "testdata/test.star", nil, starlark.StringDict{
		"hello_world_zip": starlark.String(zipBytes),
	})
	if err != nil {
		t.Error(err)
	}
}
