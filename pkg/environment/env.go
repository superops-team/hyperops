package environment

import (
	"os"

	homedir "github.com/mitchellh/go-homedir"
)

var envFiles = []string{".env.hyperops", ".env"}

// InitEnvironmentVariables init env for runtime
//   HOME: home path
//   PWD: binary path
//   UID: current user uid
func InitEnvironmentVariables(envStorage EnvStorage) error {
	var (
		homeDir, workDir string
		err              error
	)

	homeDir, err = homedir.Dir()
	if err != nil {
		return err
	}
	if envStorage.Get("HOME") == "" {
		envStorage.Set("HOME", homeDir)
	}

	initUid(envStorage)

	if envStorage.Get("PWD") == "" {
		workDir, err = os.Getwd()
		if err != nil {
			return err
		}
		envStorage.Set("PWD", workDir)
	}

	for _, envFile := range envFiles {
		if _, err = os.Stat(envFile); os.IsNotExist(err) {
			continue
		}

		err = envStorage.Load(envFile)
		if err != nil {
			return err
		}
	}
	return nil
}
