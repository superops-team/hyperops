package cmd

import (
	"github.com/spf13/cobra"
	"github.com/superops-team/hyperops/pkg/environment"
	"github.com/superops-team/hyperops/pkg/ops/starlib"
	"github.com/superops-team/hyperops/pkg/ops/starlib/sh"
	"go.starlark.net/repl"
	"go.starlark.net/starlark"
)

var replCmd = &cobra.Command{
	Use:   "repl",
	Short: "hyperops repl",
	Long:  "hyperops repl",
	Run: func(cmd *cobra.Command, args []string) {
		env := environment.NewEnvStorage()
		_ = environment.InitEnvironmentVariables(env)
		Repl()
	},
}

func Repl() {
	thread := &starlark.Thread{Load: starlib.Loader}
	locals := starlark.StringDict{
		"sh": starlark.NewBuiltin("sh", sh.Exec),
	}
	repl.REPL(thread, locals)
}

func init() {
	RootCmd.AddCommand(replCmd)
}
