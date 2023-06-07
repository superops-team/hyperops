package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// RootCmd base command
var RootCmd = &cobra.Command{
	Use:   "hyperops",
	Short: "hyperops ops commandline tool",
	Long:  "hyperops ops commandline tool for dev and ops",
	Run: func(cmd *cobra.Command, args []string) {
	},
}

// Execute 执行入口
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
