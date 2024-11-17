//go:build windows

package mi_test

import (
	"testing"
	"time"

	"github.com/prometheus-community/windows_exporter/internal/mi"
	"github.com/stretchr/testify/require"
)

type win32Process struct {
	Name string `mi:"Name"`
}

func Test_MI_Application_Initialize(t *testing.T) {
	application, err := mi.Application_Initialize()
	require.NoError(t, err)
	require.NotEmpty(t, application)

	err = application.Close()
	require.NoError(t, err)
}

func Test_MI_Application_TestConnection(t *testing.T) {
	application, err := mi.Application_Initialize()
	require.NoError(t, err)
	require.NotEmpty(t, application)

	destinationOptions, err := application.NewDestinationOptions()
	require.NoError(t, err)
	require.NotEmpty(t, destinationOptions)

	err = destinationOptions.SetTimeout(1 * time.Second)
	require.NoError(t, err)

	err = destinationOptions.SetLocale(mi.LocaleEnglish)
	require.NoError(t, err)

	session, err := application.NewSession(destinationOptions)
	require.NoError(t, err)
	require.NotEmpty(t, session)

	err = session.TestConnection()
	require.NoError(t, err)
	require.NotEmpty(t, session)

	err = session.Close()
	require.NoError(t, err)

	err = application.Close()
	require.NoError(t, err)
}

func Test_MI_Query(t *testing.T) {
	application, err := mi.Application_Initialize()
	require.NoError(t, err)
	require.NotEmpty(t, application)

	destinationOptions, err := application.NewDestinationOptions()
	require.NoError(t, err)
	require.NotEmpty(t, destinationOptions)

	err = destinationOptions.SetTimeout(1 * time.Second)
	require.NoError(t, err)

	err = destinationOptions.SetLocale(mi.LocaleEnglish)
	require.NoError(t, err)

	session, err := application.NewSession(destinationOptions)
	require.NoError(t, err)
	require.NotEmpty(t, session)

	operation, err := session.QueryInstances(mi.OperationFlagsStandardRTTI, nil, mi.NamespaceRootCIMv2, mi.QueryDialectWQL, "select Name from win32_process where handle = 0")

	require.NoError(t, err)
	require.NotEmpty(t, operation)

	instance, moreResults, err := operation.GetInstance()
	require.NoError(t, err)
	require.NotEmpty(t, instance)

	count, err := instance.GetElementCount()
	require.NoError(t, err)
	require.NotZero(t, count)

	element, err := instance.GetElement("Name")
	require.NoError(t, err)
	require.NotEmpty(t, element)

	value, err := element.GetValue()
	require.NoError(t, err)
	require.Equal(t, "System Idle Process", value)
	require.NotEmpty(t, value)

	require.False(t, moreResults)

	err = operation.Close()
	require.NoError(t, err)

	err = session.Close()
	require.NoError(t, err)

	err = application.Close()
	require.NoError(t, err)
}

func Test_MI_QueryUnmarshal(t *testing.T) {
	application, err := mi.Application_Initialize()
	require.NoError(t, err)
	require.NotEmpty(t, application)

	destinationOptions, err := application.NewDestinationOptions()
	require.NoError(t, err)
	require.NotEmpty(t, destinationOptions)

	err = destinationOptions.SetTimeout(1 * time.Second)
	require.NoError(t, err)

	err = destinationOptions.SetLocale(mi.LocaleEnglish)
	require.NoError(t, err)

	session, err := application.NewSession(destinationOptions)
	require.NoError(t, err)
	require.NotEmpty(t, session)

	var processes []win32Process

	queryProcess, err := mi.NewQuery("select Name from win32_process where handle = 0")
	require.NoError(t, err)

	err = session.QueryUnmarshal(&processes, mi.OperationFlagsStandardRTTI, nil, mi.NamespaceRootCIMv2, mi.QueryDialectWQL, queryProcess)
	require.NoError(t, err)
	require.Equal(t, []win32Process{{Name: "System Idle Process"}}, processes)

	err = session.Close()
	require.NoError(t, err)

	err = application.Close()
	require.NoError(t, err)
}

func Test_MI_EmptyQuery(t *testing.T) {
	application, err := mi.Application_Initialize()
	require.NoError(t, err)
	require.NotEmpty(t, application)

	destinationOptions, err := application.NewDestinationOptions()
	require.NoError(t, err)
	require.NotEmpty(t, destinationOptions)

	err = destinationOptions.SetTimeout(1 * time.Second)
	require.NoError(t, err)

	err = destinationOptions.SetLocale(mi.LocaleEnglish)
	require.NoError(t, err)

	session, err := application.NewSession(destinationOptions)
	require.NoError(t, err)
	require.NotEmpty(t, session)

	operation, err := session.QueryInstances(mi.OperationFlagsStandardRTTI, nil, mi.NamespaceRootCIMv2, mi.QueryDialectWQL, "SELECT Name, Status FROM win32_PrintJob")

	require.NoError(t, err)
	require.NotEmpty(t, operation)

	instance, moreResults, err := operation.GetInstance()
	require.NoError(t, err)
	require.Empty(t, instance)
	require.False(t, moreResults)

	err = operation.Close()
	require.NoError(t, err)

	err = session.Close()
	require.NoError(t, err)

	err = application.Close()
	require.NoError(t, err)
}

func Test_MI_Query_Unmarshal(t *testing.T) {
	application, err := mi.Application_Initialize()
	require.NoError(t, err)
	require.NotEmpty(t, application)

	destinationOptions, err := application.NewDestinationOptions()
	require.NoError(t, err)
	require.NotEmpty(t, destinationOptions)

	err = destinationOptions.SetTimeout(1 * time.Second)
	require.NoError(t, err)

	err = destinationOptions.SetLocale(mi.LocaleEnglish)
	require.NoError(t, err)

	session, err := application.NewSession(destinationOptions)
	require.NoError(t, err)
	require.NotEmpty(t, session)

	operation, err := session.QueryInstances(mi.OperationFlagsStandardRTTI, nil, mi.NamespaceRootCIMv2, mi.QueryDialectWQL, "SELECT Name FROM Win32_Process WHERE Handle = 0 OR Handle = 4")

	require.NoError(t, err)
	require.NotEmpty(t, operation)

	var processes []win32Process

	err = operation.Unmarshal(&processes)
	require.NoError(t, err)
	require.Equal(t, []win32Process{{Name: "System Idle Process"}, {Name: "System"}}, processes)

	err = operation.Close()
	require.NoError(t, err)

	err = session.Close()
	require.NoError(t, err)

	err = application.Close()
	require.NoError(t, err)
}
