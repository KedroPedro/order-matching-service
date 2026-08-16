package errs

import (
	"fmt"
	"runtime"
)

type EngineError struct {
	msg    string
	source string
}

func NewEngineError(message string) error {
	_, file, line, _ := runtime.Caller(1)
	return &EngineError{
		msg:    message,
		source: fmt.Sprintf("file: %s, line: %d", file, line),
	}
}

func (this *EngineError) Error() string {
	return fmt.Sprintf("engine error: %s, source: %s", this.msg, this.source)
}
