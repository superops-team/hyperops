package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/superops-team/hyperops/pkg/environment"
	"github.com/superops-team/hyperops/pkg/ops"
	"github.com/superops-team/hyperops/pkg/ops/event"
	"github.com/superops-team/hyperops/pkg/version"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"
)

var applyCmd = &cobra.Command{
	Use:   "apply",
	Short: "hyperops apply [flags]",
	Long:  "hyperops apply -f <opsfile> -n <jobname> --id=<jobid> --tags=<job tags>",
	Run: func(cmd *cobra.Command, args []string) {
		// create target to run
		target, err := ops.NewTarget(viper.GetString("file"))
		if err != nil {
			fmt.Println(err)
			os.Exit(-1)
		}

		env := environment.NewEnvStorage()
		err = environment.InitEnvironmentVariables(env)
		if err != nil {
			fmt.Println(err)
			os.Exit(-1)
		}

		for _, envVar := range viper.GetStringSlice("env") {
			pair := strings.SplitN(envVar, "=", 2)
			if len(pair) != 2 {
				continue
			}
			env.Set(pair[0], pair[1])
		}

		jobId := viper.GetString("id")
		jobName := viper.GetString("name")
		if jobId == "" {
			u, _ := uuid.NewRandom()
			jobId = u.String()
		}
		if jobName == "" {
			jobName = jobId
		}

		logdir := viper.GetString("logdir")

		if logdir != "" {
		}

		ctxConfigFile := viper.GetString("ctxconfig")
		debugFlag := viper.GetBool("debug")
		ctxMap := map[string]interface{}{}
		if debugFlag {
			ctxMap["HYPEROPS_WORKSPACE_KEEP"] = true
		}
		if ctxConfigFile != "" {
			buf, err := ioutil.ReadFile(ctxConfigFile)
			if err != nil {
				return
			}

			err = yaml.Unmarshal(buf, &ctxMap)
			if err != nil {
				return
			}
		}

		ExecuteApply(
			target,
			viper.GetString("file"),
			jobName,
			jobId,
			viper.GetString("tags"),
			viper.GetInt("timeout"),
			ctxMap,
		)
	},
}


func ExecuteApply(target *ops.Target, jobFile string, jobName string, jobId string, jobTags string, timeout int, ctxMap map[string]interface{}) {
	ctx := context.Background()
	eventCh := make(chan event.Event)
	done := make(chan struct{})
	var wg sync.WaitGroup
	// async log receive
	// 执行print(msg) 函数调用的所有日志会走到这个地方
	wg.Add(1)
	go func() {
		defer wg.Done()
		debug := viper.GetBool("debug")
		for {
			select {
			case ev := <-eventCh:
				if ev.Type == event.ETPrint {
					payload := ev.Payload.(event.PrintEvent)
					fmt.Printf("%s\n", payload.Msg)
				}
				if ev.Type == event.ETTask {
					payload := ev.Payload.(event.TaskEvent)
					if debug {
						fmt.Printf("job status change (%s -> %s)\n", payload.From, payload.To)
					}
				}
				if ev.Type == event.ETData {
					payload := ev.Payload.(event.DataEvent)
					s, _ := json.MarshalIndent(payload.Data, "", "\t")
					if debug {
						fmt.Printf("job data: \n%s\n", s)
					}
				}
			case <-done:
				return
			}
		}
	}()

	defer wg.Wait()

	v := version.GetVersion()
	cfg := map[string]interface{}{
		"job_id":             jobId,
		"job_name":           jobName,
		"job_tags":           jobTags,
		"version":   v.Version,
		"buildtime": v.BuildTime,
	}

	for k, v := range ctxMap {
		if _, ok := cfg[k]; !ok {
			cfg[k] = v
		}
	}

	opts := []func(*ops.ExecOpts){
		ops.AddEventsChannel(eventCh),
		ops.SetLocals(cfg),
		ops.SetTimeout(time.Duration(timeout) * time.Second),
	}

	err := ops.ExecScript(ctx, target, opts...)
	if err != nil {
		fmt.Println(err.Error())
	}

	done <- struct{}{}
	// time.Sleep(time.Second)
	close(eventCh)
}

func init() {
	applyCmd.PersistentFlags().StringP("file", "f", "", "ops file path, --file=/path/to/ops.star")
	BindViper(applyCmd.PersistentFlags(), "file")


	applyCmd.PersistentFlags().Bool("debug", false, "exec script in debug mode")
	BindViper(applyCmd.PersistentFlags(), "debug")

	applyCmd.PersistentFlags().StringP("user", "u", "", "user name, --user=hyperops")
	BindViper(applyCmd.PersistentFlags(), "user")

	applyCmd.PersistentFlags().StringP("name", "n", "", "job name, eg --name=jobname_uniq")
	BindViper(applyCmd.PersistentFlags(), "name")

	applyCmd.PersistentFlags().StringP("ctxconfig", "c", "", "ctx config,  --ctxconfig=ctx_config.yaml")
	BindViper(applyCmd.PersistentFlags(), "ctxconfig")

	applyCmd.PersistentFlags().StringP("where", "w", "", "where to exec script, machine or container")
	BindViper(applyCmd.PersistentFlags(), "where")

	applyCmd.PersistentFlags().StringP("id", "i", "", "job id, default auto generate by uuid, eg --id=123")
	BindViper(applyCmd.PersistentFlags(), "id")

	applyCmd.PersistentFlags().StringP("logdir", "", "", "enable log to specy log dir, default is no log support")
	BindViper(applyCmd.PersistentFlags(), "logdir")

	applyCmd.PersistentFlags().StringP("timeout", "", "1000", "set the max exec time seconds, eg --timeout=100")
	BindViper(applyCmd.PersistentFlags(), "timeout")

	applyCmd.PersistentFlags().String("tags", "", "job tags, multi tags split by comma eg --tags=a,b,c")
	BindViper(applyCmd.PersistentFlags(), "tags")

	applyCmd.PersistentFlags().StringArrayP("env", "e", []string{}, "Environment variables.")
	BindViper(applyCmd.PersistentFlags(), "env")

	RootCmd.AddCommand(applyCmd)
}
