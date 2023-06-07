package localexec

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"os/exec"
	"path"
	"strings"
	"syscall"
	"time"
)

var (
	// ErrIllegalCmdsFields Cmds经空格切分后，每一部分都必须不以/为前缀、且不存在..
	ErrIllegalCmdsFields = errors.New("the fields of cmds has illegal prefix `/` or contains illegal string `..`")
	// ErrIllegalRelativePath relativePath包含..或其为绝对路径
	ErrIllegalRelativePath = errors.New("relativePath contains illegal string `..` and must not be absolute path")
	// ErrIllegalAbsPath absPath包含..、不为绝对路径或不包含指定前缀
	ErrIllegalAbsPath = errors.New("relativePath contains illegal string `..`, must not be absolute path and must have special prefix")
)

// Result cmd的返回结果
type Result struct {
	Code   int    `json:"code"`
	Stdout string `json:"stdout"`
	Stderr string `json:"stderr"`
}

// ExecCmd 执行linux命令
func ExecCmd(timeout time.Duration, dir string, cmdName string, args ...string) (*Result, error) {
	ctx, cancle := context.WithTimeout(context.Background(), timeout)
	defer cancle()
	cmd := exec.CommandContext(ctx, cmdName, args...)
	cmd.Dir = dir

	outBuf := bytes.NewBuffer(make([]byte, 0))
	outWriter := bufio.NewWriter(outBuf)
	cmd.Stdout = outWriter

	errBuf := bytes.NewBuffer(make([]byte, 0))
	errWriter := bufio.NewWriter(errBuf)
	cmd.Stderr = errWriter
	exitCode := 0
	err := cmd.Run()
	if exitError, ok := err.(*exec.ExitError); ok {
		ws := exitError.Sys().(syscall.WaitStatus)
		exitCode = ws.ExitStatus()
		err = nil
	}
	return &Result{
		Code:   exitCode,
		Stdout: outBuf.String(),
		Stderr: errBuf.String(),
	}, err
}

// ExecBatchCmdS 批量执行shell命令
func ExecBatchCmdS(timeout time.Duration, dir string, cmds string) (*Result, error) {
	ctx, cancle := context.WithTimeout(context.Background(), timeout)
	defer cancle()
	cmd := exec.CommandContext(ctx, "sh", "-c", cmds) // ByteSec: ignore RCE
	cmd.Dir = dir

	outBuf := bytes.NewBuffer(make([]byte, 0))
	outWriter := bufio.NewWriter(outBuf)
	cmd.Stdout = outWriter

	errBuf := bytes.NewBuffer(make([]byte, 0))
	errWriter := bufio.NewWriter(errBuf)
	cmd.Stderr = errWriter

	exitCode := 0
	err := cmd.Run()
	if exitError, ok := err.(*exec.ExitError); ok {
		ws := exitError.Sys().(syscall.WaitStatus)
		exitCode = ws.ExitStatus()
		err = nil
	}
	return &Result{
		Code:   exitCode,
		Stdout: outBuf.String(),
		Stderr: errBuf.String(),
	}, err
}

// ExecRestrictedBatchCmdS 批量执行受限的shell命令
// relativePath中不得包含..
// cmds中不得包含/的前缀或..
func ExecRestrictedBatchCmdS(timeout time.Duration, basedir, relativePath, cmds string) (*Result, error) {
	err := ValidateRelativePath(relativePath)
	if err != nil {
		return nil, err
	}
	err = validateCmds(cmds)
	if err != nil {
		return nil, err
	}
	ctx, cancle := context.WithTimeout(context.Background(), timeout)
	defer cancle()

	cmd := exec.CommandContext(ctx, "sh", "-c", cmds) // ByteSec: ignore RCE
	cmd.Dir = basedir + "/" + relativePath

	outBuf := bytes.NewBuffer(make([]byte, 0))
	outWriter := bufio.NewWriter(outBuf)
	cmd.Stdout = outWriter

	errBuf := bytes.NewBuffer(make([]byte, 0))
	errWriter := bufio.NewWriter(errBuf)
	cmd.Stderr = errWriter

	exitCode := 0
	err = cmd.Run()
	if exitError, ok := err.(*exec.ExitError); ok {
		ws := exitError.Sys().(syscall.WaitStatus)
		exitCode = ws.ExitStatus()
		err = nil
	}
	return &Result{
		Code:   exitCode,
		Stdout: outBuf.String(),
		Stderr: errBuf.String(),
	}, err
}

// ValidateRelativePath 验证相对路径是否合法
func ValidateRelativePath(relativePath string) error {
	if strings.Contains(relativePath, "..") || path.IsAbs(relativePath) {
		return ErrIllegalRelativePath
	}
	return nil
}

// ValidateAbsPath 验证相对路径是否合法
func ValidateAbsPath(absPath string, basedir string) error {
	if strings.Contains(absPath, "..") || !path.IsAbs(absPath) || len(absPath) <= len(basedir) || !strings.HasPrefix(absPath, basedir) {
		return ErrIllegalAbsPath
	}
	return nil
}

func validateCmds(cmds string) error {
	fields := strings.Fields(cmds)
	for _, field := range fields {
		if strings.HasPrefix(field, "/") || strings.Contains(field, "..") {
			return ErrIllegalCmdsFields
		}
	}
	return nil
}
