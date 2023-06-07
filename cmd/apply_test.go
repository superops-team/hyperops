package cmd

import (
	"context"
	"fmt"
	"os"
	"sync"

	"github.com/superops-team/hyperops/pkg/ops"
	"github.com/superops-team/hyperops/pkg/version"
	"github.com/gammazero/workerpool"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/valyala/bytebufferpool"
)

var applyBenchCmd = &cobra.Command{
	Use:   "bench",
	Short: "hyperops bench [flags]",
	Long:  "hyperops bench -b <opsfile> -c <concurrent-num> -n <nums>",
	Run: func(cmd *cobra.Command, args []string) {
		// 模拟添加多个job并发执行,需提前读取文件，避免go并发读文件线程不安全
		scriptContent, err := os.ReadFile(viper.GetString("benchfile"))
		if err != nil {
			fmt.Println(err)
			os.Exit(-1)
		}
		wp := workerpool.New(viper.GetInt("concurrent"))
		jobs := []string{}
		for i := 1; i < viper.GetInt("nums"); i++ {
			jobs = append(jobs, fmt.Sprintf("job%d", i))
		}
		for _, r := range jobs {
			r := r
			wp.Submit(func() {
				executeapplyBench(args, r, scriptContent)
			})
		}
		wp.StopWait()
	},
}

func executeapplyBench(args []string, jobName string, jobContent []byte) {
	ctx := context.Background()
	var mu sync.RWMutex
	v := version.GetVersion()
	mu.Lock()
	cfg := map[string]interface{}{
		"job_id":             jobName,
		"job_name":           jobName,
		"job_tags":           "this is a tags var",
		"hyperops_version":   v.Version,
		"hyperops_buildtime": v.BuildTime,
	}
	secrets := map[string]string{
		"password": "1234",
	}
	mu.Unlock()

	// bytes.Buffer not thread safe, should replace with io.Pipe or others
	bb := bytebufferpool.Get()

	opts := []func(*ops.ExecOpts){
		ops.SetOutputWriter(bb),
		ops.SetLocals(cfg),
		ops.SetSecrets(secrets),
	}
	err := ops.ExecScript(ctx, &ops.Target{
		ScriptContent: jobContent,
	}, opts...)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(-1)
		return
	}
	fmt.Print(string(bb.Bytes()))
	bytebufferpool.Put(bb)
}

func init() {
	applyBenchCmd.PersistentFlags().StringP("benchfile", "b", "", "待压测的ops脚本")
	BindViper(applyBenchCmd.PersistentFlags(), "benchfile")

	applyBenchCmd.PersistentFlags().StringP("concurrent", "c", "10", "concurrent num, default 10")
	BindViper(applyBenchCmd.PersistentFlags(), "concurrent")

	applyBenchCmd.PersistentFlags().StringP("nums", "n", "1000", "nums to run, default 1000")
	BindViper(applyBenchCmd.PersistentFlags(), "nums")

	RootCmd.AddCommand(applyBenchCmd)
}
