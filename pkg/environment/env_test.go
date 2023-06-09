package environment

import (
	"os"
	"path/filepath"
	"testing"

	homedir "github.com/mitchellh/go-homedir"
)

func TestInitEnvironmentVariables(t *testing.T) {
	f := NewFakeEnvStorage()

	originalEnvFiles := envFiles
	defer func() { envFiles = originalEnvFiles }()

	testEnvFile := filepath.Join(t.TempDir(), ".env.test")
	envFiles = []string{testEnvFile}

	if err := os.WriteFile(testEnvFile, []byte("FOO=bar\n"), os.ModePerm); err != nil {
		t.Fatal(err)
	}

    _ = InitEnvironmentVariables(f)

	homeDir, _ := homedir.Dir()

	if envHomeDir := f.Envs["HOME"]; envHomeDir != homeDir {
		t.Errorf("expecting $HOME value '%s', got '%s'", homeDir, envHomeDir)
	}

	UID := uid()

	if envUID := f.Envs["UID"]; envUID != UID {
		t.Errorf("expecting $UID value '%s', got '%s'", UID, envUID)
	}

	workDir, _ := os.Getwd()

	if envWorkDir := f.Envs["PWD"]; envWorkDir != workDir {
		t.Errorf("expecting $PWD value '%s', got '%s'", workDir, envWorkDir)
	}

	if !f.CalledLoad {
		t.Error("did not call Load on EnvSotrage")
	}

	if foo := f.Envs["FOO"]; foo != "bar" {
		t.Errorf("expected FOO to be bar: %v", foo)
	}
}
