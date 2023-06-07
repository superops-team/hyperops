package util

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"

	pkgerr "github.com/pkg/errors"

	"go.starlark.net/starlark"
)

var (
	// ErrMalformattedKwarg occurs when a keyword argument passed to a starlark function is malformatted.
	ErrMalformattedKwarg = errors.New("malformatted keyword argument for method call")

	// ErrMissingKwarg occurs when a keyword argument is missing from a starlark function call.
	ErrMissingKwarg = errors.New("missing keyword argument for method call")

	// ErrMissingArg occurs when an argument is missing from a starlark function call.
	ErrMissingArg = errors.New("missing argument for method call")

	// ErrInvalidArgType occurs when an argument provided to a starlark function call has the wrong type.
	ErrInvalidArgType = errors.New("invalid argument type provided to method")

	// ErrMissingLibrary occurs when attempting to load a library that doesn't exist.
	ErrMissingLibrary = errors.New("could not find library to load")

	// ErrInvalidTypeConversion occurs when converting a golang type to a starlark.Value fails.
	ErrInvalidTypeConversion = errors.New("could not convert golang value to starlark.Value")

	// ErrInvalidKwarg occurs when a user incorrectly passes a kwarg to a starlark function
	ErrInvalidKwarg = errors.New("invalid kwarg was passed to the method")
)

type Retval interface{}

type errTuple struct {
	val Retval
	err error
}

func WithError(val Retval, err error) Retval {
	return errTuple{
		val,
		err,
	}
}

type ErrIncorrectType struct {
	shouldBe string
	is       string
}

func (eit ErrIncorrectType) Error() string {
	return fmt.Sprintf("incorrect type of %q, should be %q", eit.is, eit.shouldBe)
}

type ErrUnhashable string

func (err ErrUnhashable) Error() string {
	return fmt.Sprintf("%s is unhashable", string(err))
}

// AnnotateError error with detail info
func AnnotateError(err error) string {
	sb := new(strings.Builder)
	if err, ok := pkgerr.Cause(err).(*starlark.EvalError); ok {
		if len(err.CallStack) > 0 && err.CallStack.At(0).Pos.Filename() == "assert.star" {
			err.CallStack.Pop()
		}
		fmt.Fprintln(sb)
		fmt.Fprintf(sb, "error: %s\n", err.Msg)

		fmt.Fprint(sb, callStackString(err.CallStack))
		return sb.String()
	}
	fmt.Fprintf(sb, "%+v\n", err)
	return sb.String()
}

func callStackString(stack starlark.CallStack) string {
	out := new(strings.Builder)
	fmt.Fprintf(out, "traceback (most recent call last):\n")

	for _, fr := range stack {
		fmt.Fprintf(out, "  %s: in %s\n", fr.Pos, fr.Name)
		line := sourceLine(fr.Pos.Filename(), fr.Pos.Line)
		fmt.Fprintf(out, "    %s\n", strings.TrimSpace(line))
	}
	return out.String()
}

func sourceLine(path string, lineNumber int32) string {
	f, err := os.Open(path)
	if err != nil {
		return ""
	}
	defer func() { _ = f.Close() }()
	var index int32 = 1
	reader := bufio.NewReader(f)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			return ""
		}
		if index == lineNumber {
			return line
		}
		index++
	}
}

// DecorateError decorate all error with starlark detail
func DecorateError(thread *starlark.Thread, err error) error {
	if err == nil {
		return nil
	}
	pos := thread.CallFrame(1).Pos
	if pos.Col > 0 {
		return fmt.Errorf("%s:%d:%d: %v", pos.Filename(), pos.Line, pos.Col, err)
	}
	return fmt.Errorf("%s:%d: %v", pos.Filename(), pos.Line, err)
}
