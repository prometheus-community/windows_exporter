package hcn

import (
	"fmt"

	"github.com/prometheus-community/windows_exporter/internal/utils"
	"github.com/prometheus-community/windows_exporter/internal/win32/guid"
	"golang.org/x/sys/windows"
)

//nolint:gochecknoglobals
var (
	defaultQuery = utils.Must(windows.UTF16PtrFromString(`{"SchemaVersion":{"Major": 2,"Minor": 0},"Flags":"None"}`))
)

func GetEndpointProperties(endpointID guid.GUID) (EndpointProperties, error) {
	endpoint, err := OpenEndpoint(endpointID)
	if err != nil {
		return EndpointProperties{}, fmt.Errorf("failed to open endpoint: %w", err)
	}

	defer CloseEndpoint(endpoint)

	result, err := QueryEndpointProperties(endpoint, defaultQuery)
	if err != nil {
		return EndpointProperties{}, fmt.Errorf("failed to query endpoint properties: %w", err)
	}

	return result, nil
}
