package public

import (
	"errors"

	"github.com/prometheus-community/windows_exporter/internal/mi"
	"github.com/prometheus-community/windows_exporter/internal/perfdata"
)

var (
	ErrCollectorNotInitialized = errors.New("collector not initialized")
	ErrNoData                  = errors.New("no data")
)

var (
	// ErrsBuildCanIgnored indicates errors that can be ignored during build.
	// This is used to allow end users to enable collectors that are available on all systems, e.g. mssql.
	ErrsBuildCanIgnored = []error{
		perfdata.ErrNoData,
		mi.MI_RESULT_INVALID_NAMESPACE,
	}
)
