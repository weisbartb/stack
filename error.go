package stack

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strconv"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/ugorji/go/codec"
)

type unwrapper interface {
	Unwrap() error
}
type stackTracer interface {
	StackTrace() errors.StackTrace
}

type ErrorKVP struct {
	Key   string
	Value any
}

func (e ErrorKVP) MarshalZerologObject(ze *zerolog.Event) {
	ze.Interface(e.Key, e.Value)
}

type kvpList []ErrorKVP

func (k kvpList) MarshalZerologArray(a *zerolog.Array) {
	for _, v := range k {
		a.Object(v)
	}
}

type Error struct {
	error
	extraData kvpList
	stack     MarshallableStack
}
type t interface {
	Is(error) bool
}

func (es *Error) Is(target error) bool {
	if es == nil || target == nil {
		return false
	}
	err := es.error
	for {
		if err == target {
			return true
		}
		if err = errors.Unwrap(err); err == nil {
			return false
		}
	}
}

func Trace(err error, extra ...ErrorKVP) *Error {
	if err == nil {
		return nil
	}
	// Prevent nesting tracing
	if stackTracedError := GetTrace(err); stackTracedError != nil {
		stackTracedError.extraData = append(stackTracedError.extraData, extra...)
		return stackTracedError
	}
	return &Error{
		error: err,
		// Some things may use this
		extraData: extra,
		// Pop this off the call stack
		stack: MarshalStack()[1:],
	}
}

func Wrap(err error, message string, extra ...ErrorKVP) *Error {
	if err == nil {
		return nil
	}
	// Prevent nested tracing
	var sTrace *Error
	if errors.As(err, &sTrace) {
		sTrace.error = errors.Wrap(sTrace.error, message)
		if len(extra) > 0 {
			sTrace.extraData = append(sTrace.extraData, extra...)
		}
		return sTrace
	}
	return &Error{
		error:     errors.Wrap(err, message),
		extraData: extra,
		// Pop this off the call stack
		stack: MarshalStack()[1:],
	}
}
func (es *Error) StackTrace() errors.StackTrace {
	return es.stack.StackTrace()
}

func (es *Error) Stack() MarshallableStack {
	return es.stack
}
func (es *Error) TopLine() string {
	if len(es.stack) == 0 {
		return ""
	}
	return es.stack[0].File() + ":" + strconv.Itoa(es.stack[0].Line())
}
func (es *Error) Extra() []ErrorKVP {
	return es.extraData
}
func (es *Error) Unwrap() error {
	return es.error
}
func (es *Error) MarshalZerologObject(e *zerolog.Event) {

	e.Str("error", es.error.Error()).Array("stack", es.stack)
	if es.extraData != nil {
		e.Array("extra", es.extraData)
	}
}

func (es *Error) Cause() error {
	return es.error
}

func (es *Error) Format(s fmt.State, verb rune) {
	switch verb {
	case 'v':
		if s.Flag('+') {
			fmt.Fprintf(s, "%+v", es.Cause())
			es.stack.Format(s, verb)
			if es.extraData != nil {
				s.Write([]byte("\nExtra:\n"))
				json.NewEncoder(s).Encode(es.extraData)
			}
			return
		}
		fallthrough
	case 's':
		io.WriteString(s, es.Error())
	case 'q':
		fmt.Fprintf(s, "%q", es.Error())
	}
}

// HasTrace determines if there is a tracable error
func HasTrace(err error) (isStackTraced bool) {
	if err == nil {
		return isStackTraced
	}
	var tmp *Error
	if errors.As(err, &tmp) {
		isStackTraced = tmp.stack != nil
	}

	return isStackTraced
}

// GetTrace gets a stack trace or an empty error body, this is to ensure it can be safely chained
func GetTrace(err error) *Error {
	for {
		if err == nil {
			break
		}
		if tErr, isStackTraced := err.(stackTracer); isStackTraced {
			var internalError *Error
			if errors.As(err, &internalError) {
				return internalError
			}
			var out MarshallableStack
			for _, frame := range tErr.StackTrace() {
				out = append(out, PCFrame(frame))
			}
			return &Error{
				error:     err,
				extraData: nil,
				stack:     out,
			}
		}
		if wrapper, ok := err.(unwrapper); ok {
			err = wrapper.Unwrap()
		} else {
			break
		}
	}
	return nil
}

// TraceToString converts a trace to a string that can be read, this doesn't capture the extra data just a stack trace
// This is mostly used for tests to make it more readable if something melts down
func TraceToString(err error) string {
	if err == nil {
		return "No error"
	}
	target := GetTrace(err)
	if target == nil || target.stack == nil {
		return err.Error()
	}
	buf := bytes.Buffer{}
	codec.NewEncoder(&buf, &codec.JsonHandle{Indent: 2}).Encode(target.stack)
	return buf.String()
}
