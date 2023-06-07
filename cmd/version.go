package cmd

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/superops-team/hyperops/pkg/version"

	metrics "github.com/superops-team/hyperops/pkg/metrics"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "version",
	Long: `eg：
        hyperops version
        `,
	Run: func(cmd *cobra.Command, args []string) {
		_ = version.TextFormatTo(os.Stdout)
	},
}

func init() {
	v := version.GetVersion()
	if v.Version != "" {
		metrics.InfoGauge.WithLabelValues(v.Version).Inc()
	}
	RootCmd.AddCommand(versionCmd)
}
