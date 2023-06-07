package version

import (
	"html/template"
	"io"
	"runtime"
)

// set by build LD_FLAGS
var (
	version   string
	buildTime string
)

// Version 版本信息定义
type Version struct {
	Version   string `json:"version"`
	BuildTime string `json:"build_time"`
	GoVersion string `json:"go_version"`
	Os        string `json:"os"`
	Arch      string `json:"arch"`
}

// GetVersion 获取版本
func GetVersion() Version {
	return Version{
		Version:   version,
		BuildTime: buildTime,
		GoVersion: runtime.Version(),
		Os:        runtime.GOOS,
		Arch:      runtime.GOARCH,
	}
}

var versionTemplate = ` Version:      {{.Version}}
 Go version:   {{.GoVersion}}
 Built:        {{.BuildTime}}
 OS/Arch:      {{.Os}}/{{.Arch}}
`

// TextFormatTo 格式化
func TextFormatTo(w io.Writer) error {
	tmpl, _ := template.New("version").Parse(versionTemplate)
	return tmpl.Execute(w, GetVersion())
}
