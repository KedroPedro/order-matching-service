package errs

import "fmt"

type AppError struct {
	msg string
	err error
}

func NewAppError(msg string, err error) error {
	return &AppError{
		msg: msg,
		err: err,
	}
}

func (this *AppError) Error() string {
	return fmt.Sprintf("app error: %s, %v", this.msg, this.err)
}
