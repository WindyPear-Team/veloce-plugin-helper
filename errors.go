package pluginhelper

import "fmt"

type Error struct {
	Code    string
	Message string
}

func (e *Error) Error() string {
	if e.Code == "" {
		return e.Message
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func ErrorWithCode(code, message string) error {
	return &Error{Code: code, Message: message}
}
