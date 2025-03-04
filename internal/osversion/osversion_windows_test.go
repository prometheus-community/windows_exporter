package osversion

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOSVersionString(t *testing.T) {
	v := OSVersion{
		Version:      809042555,
		MajorVersion: 123,
		MinorVersion: 2,
		Build:        12345,
	}

	assert.Equal(t, "the version is: 123.2.12345", fmt.Sprintf("the version is: %s", v))
}
