package context

import (
	"fmt"
	"sync"

	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"

	"github.com/superops-team/hyperops/pkg/ops/util"
)

// Context 当执行脚本时携带上下文
type Context struct {
	sync.Mutex
	results starlark.StringDict // results存储执行函数调用时的返回值
	values  starlark.StringDict
	config  map[string]interface{}
	secrets map[string]string
}

// NewContext 创建上下文
func NewContext(config map[string]interface{}, secrets map[string]string) *Context {
	if secrets != nil {
		sm := NewSecretsManager()
		for _, v := range secrets {
			sm.AddSecret(v)
		}
	}
	return &Context{
		results: starlark.StringDict{},
		values:  starlark.StringDict{},
		config:  config,
		secrets: secrets,
	}
}

// Struct 作为starlark.Struct方式传递
func (c *Context) Struct() *starlarkstruct.Struct {
	dict := starlark.StringDict{
		"set":        starlark.NewBuiltin("set", c.setValue),
		"get":        starlark.NewBuiltin("get", c.getValue),
		"values":     starlark.NewBuiltin("values", c.getValues),
		"get_config": starlark.NewBuiltin("get_config", c.getConfig),
		"get_secret": starlark.NewBuiltin("get_secret", c.getSecret),
		"set_secret": starlark.NewBuiltin("set_secret", c.setSecret),
	}

	for k, v := range c.results {
		dict[k] = v
	}

	return starlarkstruct.FromStringDict(starlark.String("context"), dict)
}

// SetResult 设置执行结果
func (c *Context) SetResult(name string, value starlark.Value) {
	c.Lock()
	defer c.Unlock()
	c.results[name] = value
}

func (c *Context) setValue(thread *starlark.Thread, _ *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	c.Lock()
	defer c.Unlock()

	var (
		key   starlark.String
		value starlark.Value
	)
	if err := starlark.UnpackArgs("set", args, kwargs, "key", &key, "value", &value); err != nil {
		return starlark.None, err
	}

	c.values[string(key)] = value
	return starlark.None, nil
}

func (c *Context) getValue(thread *starlark.Thread, _ *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	c.Lock()
	defer c.Unlock()

	var key starlark.String
	if err := starlark.UnpackArgs("get", args, kwargs, "key", &key); err != nil {
		return starlark.None, err
	}
	if v, ok := c.values[string(key)]; ok {
		return v, nil
	}
	return starlark.None, fmt.Errorf("value %s not set in context", string(key))
}

func (c *Context) getValues(thread *starlark.Thread, _ *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	c.Lock()
	defer c.Unlock()
	return starlarkstruct.FromStringDict(starlarkstruct.Default, c.values), nil
}

// getSecret 获取加密信息
func (c *Context) getSecret(thread *starlark.Thread, _ *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	if c.secrets == nil {
		return starlark.None, fmt.Errorf("no secrets provided")
	}

	var key starlark.String

	c.Lock()
	defer c.Unlock()
	if err := starlark.UnpackPositionalArgs("get_secret", args, kwargs, 1, &key); err != nil {
		return nil, err
	}

	return util.Marshal(c.secrets[string(key)])
}

// setSecret 添加加密信息，添加完加密信息后print指令会屏蔽掉该加密信息
func (c *Context) setSecret(thread *starlark.Thread, _ *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	c.Lock()
	defer c.Unlock()
	if c.secrets == nil {
		c.secrets = make(map[string]string)
	}

	params, err := util.GetParser(args, kwargs)
	if err != nil {
		return starlark.None, err
	}

	key, err := params.GetString(0)
	if err != nil {
		return starlark.None, err
	}
	val, err := params.GetString(1)
	if err != nil {
		return starlark.None, err
	}

	// add secret to secrets manager
	sm := NewSecretsManager()
	sm.AddSecret(val)

	c.secrets[key] = val
	return starlark.None, nil
}

// getConfig 获取指定配置项
func (c *Context) getConfig(thread *starlark.Thread, _ *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	c.Lock()
	defer c.Unlock()
	if c.config == nil {
		return starlark.None, fmt.Errorf("no config provided")
	}

	var key starlark.String
	if err := starlark.UnpackPositionalArgs("get_config", args, kwargs, 1, &key); err != nil {
		return nil, err
	}

	return util.Marshal(c.config[string(key)])
}
