package stack_test

import (
	"bytes"
	"fmt"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
	"github.com/weisbartb/stack"
	"testing"
)

func TestHasTrace(t *testing.T) {
	var err error = &stack.Error{}
	require.False(t, stack.HasTrace(nil))
	require.False(t, stack.HasTrace(err))
	err = stack.Trace(errors.New("test"))
	require.True(t, stack.HasTrace(err))
	err = errors.New("test")
	require.False(t, stack.HasTrace(err))
}

func TestError_Cause(t *testing.T) {
	var err = &stack.Error{}
	require.Nil(t, err.Cause())
	e := errors.New("test")
	err = stack.Trace(e)
	require.Equal(t, e, err.Cause())
}

func TestError_Unwrap(t *testing.T) {
	var err = &stack.Error{}
	require.Nil(t, err.Unwrap())
	e := errors.New("test")
	err = stack.Trace(e)
	require.Equal(t, e, err.Unwrap())
}

func TestError_TopLine(t *testing.T) {
	var err = &stack.Error{}
	topLine := err.TopLine()
	require.Empty(t, topLine)
	err = stack.Trace(errors.New("test"))
	require.NotEmpty(t, err.TopLine())
}

func TestError_Stack(t *testing.T) {
	var err = &stack.Error{}
	require.Nil(t, err.Stack())
	err = stack.Trace(errors.New("test"))
	require.NotEmpty(t, err.Stack())
}

func TestError_StackTrace(t *testing.T) {
	var err = &stack.Error{}
	require.Nil(t, err.StackTrace())
	err = stack.Trace(errors.New("test"))
	require.NotEmpty(t, err.StackTrace())
}

func TestWrap(t *testing.T) {
	err := stack.Wrap(nil, "test", stack.ErrorKVP{
		Key:   "test",
		Value: "val",
	})
	require.Nil(t, err)
	err = stack.Wrap(errors.New("test error"), "test", stack.ErrorKVP{
		Key:   "test",
		Value: "val",
	})
	require.NotNil(t, err)
	require.Equal(t, "test: test error", err.Error())
	require.Len(t, err.Extra(), 1)
	err = stack.Wrap(err, "test", stack.ErrorKVP{
		Key:   "test",
		Value: "val",
	})
	require.NotNil(t, err)
	require.Equal(t, "test: test: test error", err.Error())
	require.Len(t, err.Extra(), 2)
}

func TestError_Is(t *testing.T) {
	tmpErr := errors.New("test error")
	err := stack.Wrap(tmpErr, "test")
	require.False(t, err.Is(nil))
	require.NotNil(t, err)
	require.True(t, err.Is(tmpErr))
	require.False(t, err.Is(errors.New("tmp")))
}
func TestTrace(t *testing.T) {
	err := stack.Trace(nil, stack.ErrorKVP{
		Key:   "test",
		Value: "val",
	})
	require.Nil(t, err)
	err = stack.Trace(errors.New("test error"), stack.ErrorKVP{
		Key:   "test",
		Value: "val",
	})
	require.NotNil(t, err)
	require.Equal(t, "test error", err.Error())
	require.Len(t, err.Extra(), 1)
	err2 := stack.Trace(err, stack.ErrorKVP{
		Key:   "test",
		Value: "val",
	})
	require.NotNil(t, err)
	require.Equal(t, "test error", err.Error())
	require.Len(t, err.Extra(), 2)
	require.Equal(t, err, err2)
}

func TestZeroLogIntegration(t *testing.T) {
	buf := bytes.Buffer{}
	log := zerolog.New(&buf)
	log.Error().Err(stack.Trace(errors.New("string"), stack.ErrorKVP{
		Key:   "test",
		Value: "val",
	})).Send()
	require.Contains(t, buf.String(), `"extra":[{"test":"val"}]}}`)
	buf2 := bytes.Buffer{}
	log2 := zerolog.New(&buf2)
	log2.Error().Object("embed", stack.Trace(errors.New("string"), stack.ErrorKVP{
		Key:   "test",
		Value: "val",
	})).Send()
	require.Contains(t, buf.String(), `"extra":[{"test":"val"}]}}`)
}

func TestError_Format(t *testing.T) {
	val := fmt.Sprintf("%+v", stack.Trace(errors.New("test"), stack.ErrorKVP{
		Key:   "test",
		Value: "val",
	}))
	require.Contains(t, val, "Extra:")
	require.Contains(t, val, "github.com/weisbartb/stack_test.TestError_Format")
	val2 := fmt.Sprintf("%v", stack.Trace(errors.New("test"), stack.ErrorKVP{
		Key:   "test",
		Value: "val",
	}))
	require.NotContains(t, val2, "Extra:")
	require.Contains(t, val, "github.com/weisbartb/stack_test.TestError_Format")
}

func TestTraceToString(t *testing.T) {
	msg := stack.TraceToString(errors.New("test"))
	require.Contains(t, msg, "stack_test.TestTraceToString")
	msg = stack.TraceToString(stack.Trace(errors.New("test")))
	require.Contains(t, msg, "stack_test.TestTraceToString")
}
