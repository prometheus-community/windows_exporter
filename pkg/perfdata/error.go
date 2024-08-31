package perfdata

// Error represents error returned from Performance Counters API.
type Error struct {
	ErrorCode uint32
	errorText string
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
