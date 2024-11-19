//go:build windows

package perfdata

import "errors"

var ErrNoData = NewPdhError(PdhNoData)

// Error represents error returned from Performance Counters API.
type Error struct {
	ErrorCode uint32
	errorText string
}

func (m *Error) Is(err error) bool {
	if err == nil {
		return false
	}

	var e *Error
	if errors.As(err, &e) {
		return m.ErrorCode == e.ErrorCode
	}

	return false
}

func (m *Error) Error() string {
	return m.errorText
}

func NewPdhError(code uint32) error {
	return &Error{
		ErrorCode: code,
		errorText: PdhFormatError(code),
	}
}
