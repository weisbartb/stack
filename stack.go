package stack

import (
	"bytes"
	"fmt"
	"io"
	"path"
	"runtime"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

const maxDepth = 32

type IStacker interface {
	zerolog.LogObjectMarshaler
	Stack() MarshallableStack
}

type stackFrame interface {
	zerolog.LogObjectMarshaler
	File() string
	Line() int
	Name() string
	PC() uintptr
}

func (f *PCFrame) UnmarshalJSON(b []byte) error {
	// Required for codec.JSON
	return nil
}

func (f PCFrame) MarshalJSON() ([]byte, error) {
	buf := bytes.Buffer{}
	buf.WriteByte('{')
	buf.WriteString(`"name":"` + f.Name() + `",`)
	buf.WriteString(`"file":"` + f.File() + `",`)
	buf.WriteString(`"line":` + strconv.Itoa(f.Line()))
	buf.WriteByte('}')
	return buf.Bytes(), nil
}

type PCFrame uintptr

func (f PCFrame) PC() uintptr { return uintptr(f) - 1 }

func (f PCFrame) File() string {

	fn := runtime.FuncForPC(f.PC())
	if fn == nil {
		return "unknown"
	}
	file, _ := fn.FileLine(f.PC())
	return file
}
func (f PCFrame) Line() int {
	fn := runtime.FuncForPC(f.PC())
	if fn == nil {
		return 0
	}

	_, line := fn.FileLine(f.PC())
	return line
}
func (f PCFrame) Name() string {
	fn := runtime.FuncForPC(f.PC())
	if fn == nil {
		return "unknown"
	}
	return fn.Name()
}
func (f PCFrame) Format(s fmt.State, verb rune) {
	switch verb {
	case 's':
		switch {
		case s.Flag('+'):
			io.WriteString(s, f.Name())
			io.WriteString(s, "\n\t")
			io.WriteString(s, f.File())
		default:
			io.WriteString(s, path.Base(f.File()))
		}
	case 'd':
		io.WriteString(s, strconv.Itoa(f.Line()))
	case 'v':
		f.Format(s, 's')
		io.WriteString(s, ":")
		f.Format(s, 'd')
	}
}

func (f PCFrame) MarshalZerologObject(e *zerolog.Event) {
	e.Str("caller", f.Name()).Str("file", f.File()).Int("line", f.Line())
}

type MarshallableStack []stackFrame

func (m MarshallableStack) Format(st fmt.State, verb rune) {
	switch verb {
	case 'v':
		switch {
		case st.Flag('+'):
			for _, f := range m {
				fmt.Fprintf(st, "\n%+v", f)
			}
		}
	}
}

func (m MarshallableStack) StackTrace() errors.StackTrace {
	if len(m) == 0 {
		return nil
	}
	if m[0].PC() == 0 {
		return nil
	}
	f := make(errors.StackTrace, len(m))
	for i := 0; i < len(f); i++ {
		f[i] = errors.Frame(m[i].PC())
	}
	return f
}

func (m MarshallableStack) MarshalZerologArray(a *zerolog.Array) {
	for _, v := range m {
		a.Object(v)
	}
}

func (m MarshallableStack) MarshalZerologObject(e *zerolog.Event) {
	e.Array("frames", m)
}

func (m MarshallableStack) MarshalStringFrame(i int) string {
	return m[i].Name() + " in " + m[i].File() + ":" + strconv.Itoa(m[i].Line())
}

func (m MarshallableStack) MarshalString() string {
	out := strings.Builder{}
	for i := 0; i < len(m); i++ {
		if i > 0 {
			out.WriteByte('\n')
		}
		out.WriteString(m.MarshalStringFrame(i))
	}
	return out.String()
}

func GetStackProgramCounter(offset int) MarshallableStack {

	var pcs [maxDepth]uintptr
	// Shave the current call stack off this
	sizeOfStack := runtime.Callers(offset+2, pcs[:])
	s := make(MarshallableStack, sizeOfStack)
	for i := 0; i < sizeOfStack; i++ {
		s[i] = PCFrame(pcs[i])
	}
	return s
}

func MarshalStack() MarshallableStack {
	return GetStackProgramCounter(1)
}
