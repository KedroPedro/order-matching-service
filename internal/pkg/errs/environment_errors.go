package errs

import (
	"fmt"
	"runtime"
)

type MissedEnvironmentVariableError struct {
	variableName string
	source       string
}

func NewMissedEnvironmentVariableError(varName string) error {
	_, file, line, _ := runtime.Caller(1)
	return &MissedEnvironmentVariableError{
		variableName: varName,
		source:       fmt.Sprintf("file: %s, line: %d", file, line),
	}
}

func (this *MissedEnvironmentVariableError) Error() string {
	return fmt.Sprintf("envrionment variable %s is missed or not set, source: %s", this.variableName, this.source)
}
