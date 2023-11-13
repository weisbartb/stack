package stack_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/weisbartb/stack"
	"strconv"
	"strings"
	"testing"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
)

func a() stack.MarshallableStack {
	return b()
}
func b() stack.MarshallableStack {
	return c()
}
func c() stack.MarshallableStack {
	return stack.MarshalStack()
}

func TestGetStack(t *testing.T) {
	s := a()
	require.NotNil(t, s)
	// This uses 3 for a,b,c and other 3 for runtime,test runner, this method
	require.Equal(t, 6, len(s))
}
func TestFormatStack(t *testing.T) {
	s := a()
	require.NotNil(t, s)
	for _, v := range s {
		require.NotEmpty(t, v.File())
		require.NotEmpty(t, v.Name())
	}
	require.Contains(t, s[0].Name(), "github.com/weisbartb/stack_test.c")
	require.Contains(t, s[1].Name(), "github.com/weisbartb/stack_test.b")
	require.Contains(t, s[2].Name(), "github.com/weisbartb/stack_test.a")
}
func TestGetStackSingleError(t *testing.T) {
	_, err := strconv.Atoi("banana")
	require.NotNil(t, stack.Trace(err))
}

func TestGetTrace(t *testing.T) {
	var err error = stack.Trace(nil)
	require.Nil(t, stack.GetTrace(err))
	_, err = strconv.Atoi("banana")
	err = stack.Trace(err)
	require.NotNil(t, err)
	wrapped := errors.Wrap(err, "something")
	require.Equal(t, stack.GetTrace(wrapped), err)
}

func TestError_Extra(t *testing.T) {
	_, err := strconv.Atoi("banana")
	err = stack.Trace(err, stack.ErrorKVP{
		Key:   "something",
		Value: "extra",
	})
	require.NotNil(t, err)
	wrapped := errors.Wrap(err, "something")
	require.Len(t, stack.GetTrace(wrapped).Extra(), 1)
}

func TestMarshableStack_MarshalZerologObject(t *testing.T) {
	buf := bytes.Buffer{}
	zl := zerolog.New(&buf)
	err := stack.Trace(errors.New("test"), stack.ErrorKVP{
		Key:   "test",
		Value: "val",
	})
	zl.Err(err).Send()
	require.True(t, strings.Contains(buf.String(), `"extra":[{"test":"val"}]`))
	buf.Reset()
	err = stack.Trace(err)
	zl.Err(err).Send()
	require.True(t, strings.Contains(buf.String(), `"extra":[{"test":"val"}]`))
	err = stack.Trace(err, stack.ErrorKVP{
		Key:   "test2",
		Value: "foo",
	})
	buf.Reset()
	zl.Err(err).Send()
	require.True(t, strings.Contains(buf.String(), `"extra":[{"test":"val"},{"test2":"foo"}]`))
	err = stack.Trace(err)
	buf.Reset()
	zl.Err(err).Send()
	require.True(t, strings.Contains(buf.String(), `[{"test":"val"},{"test2":"foo"}]`))
	err = stack.Trace(err, stack.ErrorKVP{
		Key:   "test2",
		Value: "foo",
	})
	buf.Reset()
	zl.Error().Object("stack", err.Stack()).Send()
	require.True(t, strings.Contains(buf.String(), `"stack":{"frames":[{"caller":`))
}

func TestMarshableStack_MarshalString(t *testing.T) {
	rawStack := stack.Trace(errors.New("test")).Stack()
	val := rawStack.MarshalString()
	require.Contains(t, val, "stack_test.TestMarshableStack_MarshalString")
}

func TestMisc(t *testing.T) {
	var frameStack stack.PCFrame
	require.NoError(t, json.Unmarshal([]byte("{}"), &frameStack))
	rawStack := stack.Trace(errors.New("test")).Stack()
	deadFrame := stack.PCFrame(uintptr(1))
	require.Equal(t, 0, deadFrame.Line())
	require.Equal(t, "unknown", deadFrame.File())
	require.Equal(t, "unknown", deadFrame.Name())
	require.Equal(t, "stack_test.go", fmt.Sprintf("%s", rawStack[0]))
}
