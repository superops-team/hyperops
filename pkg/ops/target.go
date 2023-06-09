package ops

import (
	"io/ioutil"
)

type Type string

const (
	OpsYaml     = Type("ops:yaml")
	OpsStarlark = Type("ops:starlark")
)

// Target 要执行的目标对象
type Target struct {
	ScriptPath    string `json:"script_path,omitempty"` // 执行脚本路径
	ScriptContent []byte `json:"file"`                  // 执行的脚本路径或者脚本
	ScritType     Type   `json:"type,omitempty"`        // 执行的类型
}

func NewTarget(scriptPath string) (*Target, error) {
	t := &Target{
		ScriptPath: scriptPath,
		ScritType:  OpsStarlark,
	}
	scriptContent, err := ioutil.ReadFile(scriptPath) // ByteSec: ignore FILE_OPER
	if err != nil {
		return nil, err
	}
	t.ScriptContent = scriptContent
	return t, nil
}
