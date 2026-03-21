package cli

import "errors"

type UsageError struct {
	Msg string
}

func (e *UsageError) Error() string { return e.Msg }

func ExitCode(err error) int {
	if err == nil {
		return 0
	}
	var ue *UsageError
	if errors.As(err, &ue) {
		return 2
	}
	return 1
}
