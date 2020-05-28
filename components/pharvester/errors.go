package pharvester

import "fmt"

const (
	ErrCodeUntyped                 = 1
	ErrCodePacketLoss              = 2
	ErrCodeRequestFail             = 3
	ErrCodeBadResponse             = 4
	ErrCodeCannotCreateProxyDialer = 5
	ErrCodeInvalidProxyURL         = 6
)

func NewError(code int, parent error, message ...interface{}) *Error {
	return &Error{Code: code, Message: message, Parent: parent}
}

type Error struct {
	Code    int
	Message []interface{}
	Parent  error
}

func (e Error) Error() string {
	return fmt.Sprintf("parent:[%v] code:[%d] msg:[%s]", e.Parent, e.Code, e.Message)
}
