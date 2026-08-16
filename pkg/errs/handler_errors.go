package errs

import (
	"fmt"
	"runtime"
)

type HandlerError struct {
	msg    string
	source string
	err    error
}

func NewHandlerError(message string, err error) error {
	_, file, line, _ := runtime.Caller(1)
	return &HandlerError{
		msg:    message,
		source: fmt.Sprintf("file: %s, line: %d", file, line),
		err:    err,
	}
}

func (this *HandlerError) Error() string {
	return fmt.Sprintf("engine error: %s, source: %s, %v", this.msg, this.source, this.err)
}
