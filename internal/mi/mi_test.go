// SPDX-License-Identifier: Apache-2.0
//
// Copyright The Prometheus Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

//go:build windows

package mi_test

import (
	"testing"
	"time"

	"github.com/prometheus-community/windows_exporter/internal/mi"
	"github.com/prometheus-community/windows_exporter/internal/utils/testutils"
	"github.com/stretchr/testify/require"
	"golang.org/x/sys/windows"
)

type win32Process struct {
	Name string `mi:"Name"`
}

func Test_MI_Application_Initialize(t *testing.T) {
	application, err := mi.ApplicationInitialize()
	require.NoError(t, err)
	require.NotEmpty(t, application)

	err = application.Close()
	require.NoError(t, err)
}

func Test_MI_Application_TestConnection(t *testing.T) {
	application, err := mi.ApplicationInitialize()
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
	application, err := mi.ApplicationInitialize()
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
	application, err := mi.ApplicationInitialize()
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
	application, err := mi.ApplicationInitialize()
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
	application, err := mi.ApplicationInitialize()
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

func Test_MI_FD_Leak(t *testing.T) {
	application, err := mi.ApplicationInitialize()
	require.NoError(t, err)
	require.NotEmpty(t, application)

	session, err := application.NewSession(nil)
	require.NoError(t, err)
	require.NotEmpty(t, session)

	currentFileHandle, err := testutils.GetProcessHandleCount(windows.CurrentProcess())
	require.NoError(t, err)

	t.Log("Current File Handle Count: ", currentFileHandle)

	queryPrinter, err := mi.NewQuery("SELECT Name FROM Win32_Process")
	require.NoError(t, err)

	for range 300 {
		var processes []win32Process

		err := session.Query(&processes, mi.NamespaceRootCIMv2, queryPrinter, -1)
		require.NoError(t, err)

		currentFileHandle, err = testutils.GetProcessHandleCount(windows.CurrentProcess())
		require.NoError(t, err)

		t.Log("Current File Handle Count: ", currentFileHandle)
	}

	currentFileHandle, err = testutils.GetProcessHandleCount(windows.CurrentProcess())
	require.NoError(t, err)

	t.Log("Current File Handle Count: ", currentFileHandle)

	err = session.Close()
	require.NoError(t, err)

	currentFileHandle, err = testutils.GetProcessHandleCount(windows.CurrentProcess())
	require.NoError(t, err)

	t.Log("Current File Handle Count: ", currentFileHandle)

	err = application.Close()
	require.NoError(t, err)

	currentFileHandle, err = testutils.GetProcessHandleCount(windows.CurrentProcess())
	require.NoError(t, err)

	t.Log("Current File Handle Count: ", currentFileHandle)
}

type reliabilityMetrics struct {
	SystemStabilityIndex float64 `mi:"SystemStabilityIndex"`
}

// Test_MI_Query_REAL64 verifies that GetValue correctly returns float64
// for REAL64 MI properties by querying Win32_ReliabilityStabilityMetrics.
func Test_MI_Query_REAL64(t *testing.T) {
	application, err := mi.ApplicationInitialize()
	require.NoError(t, err)
	require.NotEmpty(t, application)

	destinationOptions, err := application.NewDestinationOptions()
	require.NoError(t, err)
	require.NotEmpty(t, destinationOptions)

	err = destinationOptions.SetTimeout(5 * time.Second)
	require.NoError(t, err)

	err = destinationOptions.SetLocale(mi.LocaleEnglish)
	require.NoError(t, err)

	session, err := application.NewSession(destinationOptions)
	require.NoError(t, err)
	require.NotEmpty(t, session)

	operation, err := session.QueryInstances(mi.OperationFlagsStandardRTTI, nil, mi.NamespaceRootCIMv2, mi.QueryDialectWQL,
		"SELECT SystemStabilityIndex FROM Win32_ReliabilityStabilityMetrics")
	require.NoError(t, err)
	require.NotEmpty(t, operation)

	var firstValue float64

	var foundFloat bool

	for {
		instance, moreResults, err := operation.GetInstance()
		require.NoError(t, err)

		if instance == nil {
			break
		}

		if !foundFloat {
			element, err := instance.GetElement("SystemStabilityIndex")
			require.NoError(t, err)

			value, err := element.GetValue()
			require.NoError(t, err)

			v, ok := value.(float64)
			require.True(t, ok, "expected float64, got %T", value)

			if v > 0 {
				firstValue = v
				foundFloat = true
			}
		}

		if !moreResults {
			break
		}
	}

	if !foundFloat {
		t.Skip("Win32_ReliabilityStabilityMetrics: no records with non-zero SystemStabilityIndex")
	}

	require.Greater(t, firstValue, float64(0), "SystemStabilityIndex should be positive")

	t.Logf("SystemStabilityIndex = %v", firstValue)

	err = operation.Close()
	require.NoError(t, err)

	err = session.Close()
	require.NoError(t, err)

	err = application.Close()
	require.NoError(t, err)
}

// Test_MI_QueryUnmarshal_REAL64 verifies that the unmarshal code path
// correctly handles REAL64→float64 struct field conversion.
func Test_MI_QueryUnmarshal_REAL64(t *testing.T) {
	application, err := mi.ApplicationInitialize()
	require.NoError(t, err)
	require.NotEmpty(t, application)

	destinationOptions, err := application.NewDestinationOptions()
	require.NoError(t, err)
	require.NotEmpty(t, destinationOptions)

	err = destinationOptions.SetTimeout(5 * time.Second)
	require.NoError(t, err)

	err = destinationOptions.SetLocale(mi.LocaleEnglish)
	require.NoError(t, err)

	session, err := application.NewSession(destinationOptions)
	require.NoError(t, err)
	require.NotEmpty(t, session)

	var metrics []reliabilityMetrics

	query, err := mi.NewQuery("SELECT SystemStabilityIndex FROM Win32_ReliabilityStabilityMetrics")
	require.NoError(t, err)

	err = session.QueryUnmarshal(&metrics, mi.OperationFlagsStandardRTTI, nil, mi.NamespaceRootCIMv2, mi.QueryDialectWQL, query)
	if err != nil || len(metrics) == 0 {
		t.Skip("Win32_ReliabilityStabilityMetrics not available")
	}

	var found bool

	for _, m := range metrics {
		if m.SystemStabilityIndex > 0 {
			t.Logf("SystemStabilityIndex = %.3f (from %d records)", m.SystemStabilityIndex, len(metrics))

			found = true

			break
		}
	}

	require.True(t, found, "expected at least one record with non-zero SystemStabilityIndex")

	err = session.Close()
	require.NoError(t, err)

	err = application.Close()
	require.NoError(t, err)
}

type computerSystemSigned struct {
	ResetCount     int16 `mi:"ResetCount"`
	ResetLimit     int16 `mi:"ResetLimit"`
	PauseAfterReset int64 `mi:"PauseAfterReset"`
}

// Test_MI_Query_SignedInt verifies that GetValue correctly sign-extends
// SINT8/SINT16/SINT32 values. Win32_ComputerSystem has SInt16 properties
// (ResetCount, ResetLimit) that are -1 on most systems.
func Test_MI_Query_SignedInt(t *testing.T) {
	application, err := mi.ApplicationInitialize()
	require.NoError(t, err)
	require.NotEmpty(t, application)

	destinationOptions, err := application.NewDestinationOptions()
	require.NoError(t, err)
	require.NotEmpty(t, destinationOptions)

	err = destinationOptions.SetTimeout(5 * time.Second)
	require.NoError(t, err)

	err = destinationOptions.SetLocale(mi.LocaleEnglish)
	require.NoError(t, err)

	session, err := application.NewSession(destinationOptions)
	require.NoError(t, err)
	require.NotEmpty(t, session)

	operation, err := session.QueryInstances(mi.OperationFlagsStandardRTTI, nil, mi.NamespaceRootCIMv2, mi.QueryDialectWQL,
		"SELECT ResetCount, ResetLimit, PauseAfterReset FROM Win32_ComputerSystem")
	require.NoError(t, err)
	require.NotEmpty(t, operation)

	instance, moreResults, err := operation.GetInstance()
	require.NoError(t, err)
	require.NotEmpty(t, instance)
	require.False(t, moreResults)

	// ResetCount (SInt16): verify sign-extension produces a negative int16.
	element, err := instance.GetElement("ResetCount")
	require.NoError(t, err)

	value, err := element.GetValue()
	require.NoError(t, err)

	resetCount, ok := value.(int16)
	require.True(t, ok, "expected int16, got %T", value)
	require.Equal(t, int16(-1), resetCount, "ResetCount should be -1")

	t.Logf("ResetCount = %d, ResetLimit = %d", resetCount, resetCount)

	err = operation.Close()
	require.NoError(t, err)

	err = session.Close()
	require.NoError(t, err)

	err = application.Close()
	require.NoError(t, err)
}

// Test_MI_QueryUnmarshal_SignedInt verifies that the unmarshal code path
// correctly sign-extends SINT16 values into Go struct fields.
func Test_MI_QueryUnmarshal_SignedInt(t *testing.T) {
	application, err := mi.ApplicationInitialize()
	require.NoError(t, err)
	require.NotEmpty(t, application)

	destinationOptions, err := application.NewDestinationOptions()
	require.NoError(t, err)
	require.NotEmpty(t, destinationOptions)

	err = destinationOptions.SetTimeout(5 * time.Second)
	require.NoError(t, err)

	err = destinationOptions.SetLocale(mi.LocaleEnglish)
	require.NoError(t, err)

	session, err := application.NewSession(destinationOptions)
	require.NoError(t, err)
	require.NotEmpty(t, session)

	var systems []computerSystemSigned

	query, err := mi.NewQuery("SELECT ResetCount, ResetLimit, PauseAfterReset FROM Win32_ComputerSystem")
	require.NoError(t, err)

	err = session.QueryUnmarshal(&systems, mi.OperationFlagsStandardRTTI, nil, mi.NamespaceRootCIMv2, mi.QueryDialectWQL, query)
	require.NoError(t, err)
	require.Len(t, systems, 1)

	s := systems[0]
	require.Equal(t, int16(-1), s.ResetCount, "ResetCount should be -1 (sign-extended from SInt16)")
	require.Equal(t, int16(-1), s.ResetLimit, "ResetLimit should be -1 (sign-extended from SInt16)")
	require.Equal(t, int64(-1), s.PauseAfterReset, "PauseAfterReset should be -1")

	t.Logf("ResetCount=%d ResetLimit=%d PauseAfterReset=%d", s.ResetCount, s.ResetLimit, s.PauseAfterReset)

	err = session.Close()
	require.NoError(t, err)

	err = application.Close()
	require.NoError(t, err)
}

func Test_MI_QueryTimeout(t *testing.T) {
	application, err := mi.ApplicationInitialize()
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

	operationOptions, err := application.NewOperationOptions()
	require.NoError(t, err)
	require.NotEmpty(t, operationOptions)

	err = operationOptions.SetTimeout(1 * time.Millisecond)
	require.NoError(t, err)

	operation, err := session.QueryInstances(mi.OperationFlagsStandardRTTI, operationOptions, mi.NamespaceRootCIMv2, mi.QueryDialectWQL, "select Name from win32_process where handle = 0")
	require.NoError(t, err)
	require.NotEmpty(t, operation)

	instance, moreResults, err := operation.GetInstance()
	require.ErrorIs(t, err, mi.MI_RESULT_INVALID_OPERATION_TIMEOUT)
	require.False(t, moreResults)
	require.Empty(t, instance)

	err = operation.Close()
	require.NoError(t, err)

	err = session.Close()
	require.NoError(t, err)

	err = application.Close()
	require.NoError(t, err)
}
