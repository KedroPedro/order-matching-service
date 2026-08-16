package errs

import (
	"fmt"
	"runtime"
)

type RepositoryError struct {
	msg    string
	source string
	err    error
}

func NewRepositoryError(message string, err error) error {
	_, file, line, _ := runtime.Caller(1)
	return &RepositoryError{
		msg:    message,
		source: fmt.Sprintf("file: %s, line: %d", file, line),
		err:    err,
	}
}

func (this *RepositoryError) Error() string {
	return fmt.Sprintf("repository error: %s, source: %s, %v", this.msg, this.source, this.err)
}

func (this *RepositoryError) Unwrap() error {
	return this.err
}
