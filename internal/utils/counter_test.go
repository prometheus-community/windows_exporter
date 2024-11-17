//go:build windows

package utils_test

import (
	"math"
	"testing"

	"github.com/prometheus-community/windows_exporter/internal/utils"
	"github.com/stretchr/testify/assert"
)

func TestCounter(t *testing.T) {
	t.Parallel()

	c := utils.NewCounter(0)
	assert.Equal(t, 0.0, c.Value()) //nolint:testifylint

	c.AddValue(1)

	assert.Equal(t, 1.0, c.Value()) //nolint:testifylint

	c.AddValue(math.MaxUint32)

	assert.Equal(t, float64(math.MaxUint32)+1.0, c.Value()) //nolint:testifylint
}
