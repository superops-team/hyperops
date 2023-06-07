package ops

import (
	"io"
	"io/ioutil"
	"time"

	"github.com/superops-team/hyperops/pkg/ops/event"
)

// ExecOpts 设置运行时相关开关
type ExecOpts struct {
	// 是否启用浮点数
	AllowFloat bool
	// 是否启用集合类型
	AllowSet bool
	// 是否启用lamda表达式
	AllowLambda bool
	// 是否启用闭包
	AllowNestedDef bool
	// allow reassignment to top-level names; also, allow if/for/while at top-level
	AllowGlobalReassign bool
	// 密码相关
	Secrets map[string]string
	// 局部变量
	Locals map[string]interface{}
	// output接收
	OutputWriter io.Writer
	// 模块加载方法
	ModuleLoader ModuleLoader
	// 事件转发订阅
	EventsCh chan event.Event
	// 超时
	Timeout time.Duration
}

// DefaultExecOpts 默认执行配置
func DefaultExecOpts(o *ExecOpts) {
	o.AllowFloat = true
	o.AllowSet = true
	o.AllowLambda = true
	o.AllowGlobalReassign = true
	o.OutputWriter = ioutil.Discard
	o.ModuleLoader = DefaultModuleLoader
	o.Timeout = 100 * time.Second
}

// AddEventsChannel 设置事件接收器
func AddEventsChannel(eventsCh chan event.Event) func(o *ExecOpts) {
	return func(o *ExecOpts) {
		o.EventsCh = eventsCh
	}
}

// SetOutputWriter 设置输出重定向
func SetOutputWriter(w io.Writer) func(o *ExecOpts) {
	return func(o *ExecOpts) {
		if w != nil {
			o.OutputWriter = w
		}
	}
}

// SetSecrets 设置KEY/VAL对
func SetSecrets(secrets map[string]string) func(o *ExecOpts) {
	return func(o *ExecOpts) {
		if secrets != nil {
			if len(secrets) == 0 {
				return
			}
			// notice: this a copy of secrets for thread safety
			s := make(map[string]string, len(secrets))
			for key, val := range secrets {
				s[key] = val
			}
			o.Secrets = s
		}
	}
}

// SetLocals 设置KEY/VAL对
func SetLocals(locals map[string]interface{}) func(o *ExecOpts) {
	return func(o *ExecOpts) {
		if locals != nil {
			if len(locals) == 0 {
				return
			}
			// notice: a copy of local config for thread safe
			l := make(map[string]interface{}, len(locals))
			for key, val := range locals {
				l[key] = val
			}
			o.Locals = l
		}
	}
}

// SetTimeout 设置超时
func SetTimeout(duration time.Duration) func(o *ExecOpts) {
	return func(o *ExecOpts) {
		o.Timeout = duration
	}
}
