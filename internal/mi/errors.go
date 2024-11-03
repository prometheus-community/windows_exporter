//go:build windows

package mi

import "errors"

var (
	ErrNotInitialized    = errors.New("not initialized")
	ErrInvalidEntityType = errors.New("invalid entity type")
)
