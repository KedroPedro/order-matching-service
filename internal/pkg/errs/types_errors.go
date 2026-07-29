package errs

import (
	"fmt"
	"runtime"
)

type TypeError struct {
	msg    string
	source string
}

func NewTypeError(message string) error {
	_, file, line, _ := runtime.Caller(1)
	return &TypeError{
		msg:    message,
		source: fmt.Sprintf("file: %s, line: %d", file, line),
	}
}

func (this *TypeError) Error() string {
	return fmt.Sprintf("type error: %s, source: %s", this.msg, this.source)
}
