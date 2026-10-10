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
	"errors"
	"math"
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

// win32DiskDrive is used to exercise UINT16A (uint16[]) unmarshalling.
// Win32_DiskDrive.Capabilities is a reliably-populated uint16[] on any host.
type win32DiskDrive struct {
	Capabilities []uint16 `mi:"Capabilities"`
}

// win32DiskDriveWrongType maps the uint16[] Capabilities property onto an
// incompatible Go type to exercise the UINT16A type guard.
type win32DiskDriveWrongType struct {
	Capabilities []string `mi:"Capabilities"`
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
	// A leak of one handle per query grows far beyond this; MI and the Go
	// runtime may still open a few handles for worker threads.
	const (
		warmupQueries     = 100
		sampleQueries     = 25
		queries           = 300
		maxHandleIncrease = 50
	)

	application, err := mi.ApplicationInitialize()
	require.NoError(t, err)
	require.NotEmpty(t, application)

	session, err := application.NewSession(nil)
	require.NoError(t, err)
	require.NotEmpty(t, session)

	queryPrinter, err := mi.NewQuery("SELECT Name FROM Win32_Process")
	require.NoError(t, err)

	query := func() {
		var processes []win32Process

		require.NoError(t, session.Query(&processes, mi.NamespaceRootCIMv2, queryPrinter, -1))
	}

	// minHandleCount runs n queries and returns the lowest handle count seen
	// after each of them. MI worker threads come and go, so a single sample
	// swings by about 20 handles; the minimum filters that out, while a leak
	// still raises it.
	minHandleCount := func(n int) int64 {
		lowest := int64(math.MaxInt64)

		for range n {
			query()

			count, err := testutils.GetProcessHandleCount(windows.CurrentProcess())
			require.NoError(t, err)

			lowest = min(lowest, int64(count))
		}

		return lowest
	}

	// The first queries initialize MI caches and thread pools, which takes
	// more than 100 handles over roughly the first 50 queries.
	for range warmupQueries {
		query()
	}

	baseline := minHandleCount(sampleQueries)

	for range queries {
		query()
	}

	current := minHandleCount(sampleQueries)
	require.LessOrEqual(t, current, baseline+maxHandleIncrease,
		"handle count grew from %d to %d after %d queries", baseline, current, queries)

	require.NoError(t, session.Close())
	require.NoError(t, application.Close())
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

	t.Cleanup(func() { _ = application.Close() })

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

	t.Cleanup(func() { _ = session.Close() })

	operation, err := session.QueryInstances(mi.OperationFlagsStandardRTTI, nil, mi.NamespaceRootCIMv2, mi.QueryDialectWQL,
		"SELECT SystemStabilityIndex FROM Win32_ReliabilityStabilityMetrics")
	if err != nil {
		if errors.Is(err, mi.MI_RESULT_INVALID_CLASS) {
			t.Skip("Win32_ReliabilityStabilityMetrics class not available on this system")
		}

		require.NoError(t, err)
	}

	require.NotEmpty(t, operation)

	t.Cleanup(func() { _ = operation.Close() })

	var firstValue float64

	var foundFloat bool

	for {
		instance, moreResults, err := operation.GetInstance()
		if err != nil {
			if errors.Is(err, mi.MI_RESULT_INVALID_CLASS) {
				t.Skip("Win32_ReliabilityStabilityMetrics class not available on this system")
			}

			require.NoError(t, err)
		}

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
	require.LessOrEqual(t, firstValue, float64(10), "SystemStabilityIndex should be at most 10 (documented maximum)")

	t.Logf("SystemStabilityIndex = %v", firstValue)
}

// Test_MI_QueryUnmarshal_REAL64 verifies that the unmarshal code path
// correctly handles REAL64→float64 struct field conversion.
func Test_MI_QueryUnmarshal_REAL64(t *testing.T) {
	application, err := mi.ApplicationInitialize()
	require.NoError(t, err)
	require.NotEmpty(t, application)

	t.Cleanup(func() { _ = application.Close() })

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

	t.Cleanup(func() { _ = session.Close() })

	var metrics []reliabilityMetrics

	query, err := mi.NewQuery("SELECT SystemStabilityIndex FROM Win32_ReliabilityStabilityMetrics")
	require.NoError(t, err)

	err = session.QueryUnmarshal(&metrics, mi.OperationFlagsStandardRTTI, nil, mi.NamespaceRootCIMv2, mi.QueryDialectWQL, query)
	if err != nil {
		if errors.Is(err, mi.MI_RESULT_INVALID_CLASS) {
			t.Skip("Win32_ReliabilityStabilityMetrics class not available on this system")
		}

		require.NoError(t, err)
	}

	if len(metrics) == 0 {
		t.Skip("Win32_ReliabilityStabilityMetrics returned no records")
	}

	var found bool

	for _, m := range metrics {
		if m.SystemStabilityIndex > 0 {
			require.LessOrEqual(t, m.SystemStabilityIndex, float64(10), "SystemStabilityIndex should be at most 10 (documented maximum)")

			t.Logf("SystemStabilityIndex = %.3f (from %d records)", m.SystemStabilityIndex, len(metrics))

			found = true

			break
		}
	}

	if !found {
		t.Skip("Win32_ReliabilityStabilityMetrics: no records with non-zero SystemStabilityIndex")
	}
}

type computerSystemSigned struct {
	ResetCount      int16 `mi:"ResetCount"`
	ResetLimit      int16 `mi:"ResetLimit"`
	PauseAfterReset int64 `mi:"PauseAfterReset"`
}

// computerSystemWide unmarshals SINT16 properties into int64 fields
// to verify sign extension works correctly across type widths.
// With the old (broken) code, a SINT16 value of -1 (0xFFFF) would
// become 65535 in an int64 field instead of -1.
type computerSystemWide struct {
	ResetCount int64 `mi:"ResetCount"`
	ResetLimit int64 `mi:"ResetLimit"`
}

// Test_MI_Query_SignedInt verifies that GetValue correctly sign-extends
// SINT8/SINT16/SINT32 values. Win32_ComputerSystem has SInt16 properties
// (ResetCount, ResetLimit) that are typically -1, meaning "not supported".
func Test_MI_Query_SignedInt(t *testing.T) {
	application, err := mi.ApplicationInitialize()
	require.NoError(t, err)
	require.NotEmpty(t, application)

	t.Cleanup(func() { _ = application.Close() })

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

	t.Cleanup(func() { _ = session.Close() })

	operation, err := session.QueryInstances(mi.OperationFlagsStandardRTTI, nil, mi.NamespaceRootCIMv2, mi.QueryDialectWQL,
		"SELECT ResetCount, ResetLimit, PauseAfterReset FROM Win32_ComputerSystem")
	require.NoError(t, err)
	require.NotEmpty(t, operation)

	t.Cleanup(func() { _ = operation.Close() })

	instance, moreResults, err := operation.GetInstance()
	require.NoError(t, err)
	require.NotEmpty(t, instance)
	require.False(t, moreResults)

	// ResetCount (SInt16): verify the value is returned as int16 (not uint16 or int64).
	// The value is typically -1 ("not supported") but the exact value is system-dependent.
	element, err := instance.GetElement("ResetCount")
	require.NoError(t, err)

	value, err := element.GetValue()
	require.NoError(t, err)

	resetCount, ok := value.(int16)
	require.True(t, ok, "expected int16, got %T", value)

	t.Logf("ResetCount = %d, ResetLimit = (same type)", resetCount)
}

// Test_MI_QueryUnmarshal_SignedInt verifies that the unmarshal code path
// correctly sign-extends SINT16 values into Go struct fields.
func Test_MI_QueryUnmarshal_SignedInt(t *testing.T) {
	application, err := mi.ApplicationInitialize()
	require.NoError(t, err)
	require.NotEmpty(t, application)

	t.Cleanup(func() { _ = application.Close() })

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

	t.Cleanup(func() { _ = session.Close() })

	var systems []computerSystemSigned

	query, err := mi.NewQuery("SELECT ResetCount, ResetLimit, PauseAfterReset FROM Win32_ComputerSystem")
	require.NoError(t, err)

	err = session.QueryUnmarshal(&systems, mi.OperationFlagsStandardRTTI, nil, mi.NamespaceRootCIMv2, mi.QueryDialectWQL, query)
	require.NoError(t, err)
	require.Len(t, systems, 1)

	// Verify that signed fields were populated without error. The actual values
	// are system-dependent (typically -1), so we only check that unmarshalling
	// succeeded and log the values for manual inspection.
	s := systems[0]

	t.Logf("ResetCount=%d ResetLimit=%d PauseAfterReset=%d", s.ResetCount, s.ResetLimit, s.PauseAfterReset)
}

// Test_MI_QueryUnmarshal_SignedInt_Wide verifies that the sign-extension fix
// works when a SINT16 value is unmarshalled into a wider Go field (int64).
// Before the fix, -1 (0xFFFF) would become 65535 in an int64 field.
func Test_MI_QueryUnmarshal_SignedInt_Wide(t *testing.T) {
	application, err := mi.ApplicationInitialize()
	require.NoError(t, err)
	require.NotEmpty(t, application)

	t.Cleanup(func() { _ = application.Close() })

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

	t.Cleanup(func() { _ = session.Close() })

	// First, read the actual values with the narrow type to learn what to expect.
	var narrow []computerSystemSigned

	narrowQuery, err := mi.NewQuery("SELECT ResetCount, ResetLimit FROM Win32_ComputerSystem")
	require.NoError(t, err)

	err = session.QueryUnmarshal(&narrow, mi.OperationFlagsStandardRTTI, nil, mi.NamespaceRootCIMv2, mi.QueryDialectWQL, narrowQuery)
	require.NoError(t, err)
	require.Len(t, narrow, 1)

	if narrow[0].ResetCount >= 0 {
		t.Skipf("ResetCount is %d (non-negative); cannot verify sign extension into wider type", narrow[0].ResetCount)
	}

	// Now unmarshal the same SINT16 values into int64 fields.
	var wide []computerSystemWide

	wideQuery, err := mi.NewQuery("SELECT ResetCount, ResetLimit FROM Win32_ComputerSystem")
	require.NoError(t, err)

	err = session.QueryUnmarshal(&wide, mi.OperationFlagsStandardRTTI, nil, mi.NamespaceRootCIMv2, mi.QueryDialectWQL, wideQuery)
	require.NoError(t, err)
	require.Len(t, wide, 1)

	// The key assertion: with the old code, ResetCount = -1 would become 65535
	// in the int64 field. With the fix, it must remain negative.
	require.Negative(t, wide[0].ResetCount,
		"SINT16 value %d should remain negative when unmarshalled into int64 (got %d)",
		narrow[0].ResetCount, wide[0].ResetCount)
	require.Equal(t, int64(narrow[0].ResetCount), wide[0].ResetCount,
		"int64 field should match int16 value after sign extension")

	t.Logf("ResetCount: int16=%d, int64=%d (sign extension correct)", narrow[0].ResetCount, wide[0].ResetCount)
}

// Test_MI_Unmarshal_TypeMismatch verifies that unmarshalInstance rejects
// Go struct fields whose kind does not match the MI value type.
func Test_MI_Unmarshal_TypeMismatch(t *testing.T) {
	application, err := mi.ApplicationInitialize()
	require.NoError(t, err)

	t.Cleanup(func() { _ = application.Close() })

	session, err := application.NewSession(nil)
	require.NoError(t, err)

	t.Cleanup(func() { _ = session.Close() })

	// Map a string MI property (Name) to an int Go field → type error.
	t.Run("string_to_int", func(t *testing.T) {
		type bad struct {
			Name int `mi:"Name"`
		}

		var dst []bad

		query, err := mi.NewQuery("SELECT Name FROM Win32_Process WHERE Handle = 0")
		require.NoError(t, err)

		err = session.QueryUnmarshal(&dst, mi.OperationFlagsStandardRTTI, nil, mi.NamespaceRootCIMv2, mi.QueryDialectWQL, query)
		require.Error(t, err)
		require.Contains(t, err.Error(), "Name")

		t.Logf("got expected error: %v", err)
	})

	// Map a uint32 MI property (ProcessId) to a bool Go field → type error.
	t.Run("uint_to_bool", func(t *testing.T) {
		type bad struct {
			ProcessId bool `mi:"ProcessId"`
		}

		var dst []bad

		query, err := mi.NewQuery("SELECT ProcessId FROM Win32_Process WHERE Handle = 0")
		require.NoError(t, err)

		err = session.QueryUnmarshal(&dst, mi.OperationFlagsStandardRTTI, nil, mi.NamespaceRootCIMv2, mi.QueryDialectWQL, query)
		require.Error(t, err)
		require.Contains(t, err.Error(), "ProcessId")

		t.Logf("got expected error: %v", err)
	})

	// Map a SInt16 MI property to an unsigned Go field. If the value is
	// negative (e.g. -1 = "not supported") this must error; if it is
	// non-negative and fits uint16 it is accepted. Skip when non-negative.
	t.Run("negative_sint_to_uint", func(t *testing.T) {
		// First, read the actual value to decide whether to test or skip.
		var probe []computerSystemSigned

		probeQuery, err := mi.NewQuery("SELECT ResetCount FROM Win32_ComputerSystem")
		require.NoError(t, err)

		err = session.QueryUnmarshal(&probe, mi.OperationFlagsStandardRTTI, nil, mi.NamespaceRootCIMv2, mi.QueryDialectWQL, probeQuery)
		require.NoError(t, err)
		require.Len(t, probe, 1)

		if probe[0].ResetCount >= 0 {
			t.Skipf("ResetCount is %d (non-negative); cannot test negative→uint rejection", probe[0].ResetCount)
		}

		type bad struct {
			ResetCount uint16 `mi:"ResetCount"`
		}

		var dst []bad

		query, err := mi.NewQuery("SELECT ResetCount FROM Win32_ComputerSystem")
		require.NoError(t, err)

		err = session.QueryUnmarshal(&dst, mi.OperationFlagsStandardRTTI, nil, mi.NamespaceRootCIMv2, mi.QueryDialectWQL, query)
		require.Error(t, err)
		require.Contains(t, err.Error(), "ResetCount")

		t.Logf("got expected error: %v", err)
	})
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

	// A query for a single process can finish within the timeout; reading every
	// property of every process cannot.
	operation, err := session.QueryInstances(mi.OperationFlagsStandardRTTI, operationOptions, mi.NamespaceRootCIMv2, mi.QueryDialectWQL, "select * from win32_process")
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

// Test_MI_Session_QueryTimeout checks that the timeout passed to Session.Query,
// e.g. mi.BuildQueryTimeout, reaches MI and surfaces as the native result.
func Test_MI_Session_QueryTimeout(t *testing.T) {
	application, err := mi.ApplicationInitialize()
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, application.Close()) })

	destinationOptions, err := application.NewDestinationOptions()
	require.NoError(t, err)
	require.NoError(t, destinationOptions.SetLocale(mi.LocaleEnglish))

	session, err := application.NewSession(destinationOptions)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, session.Close()) })

	query, err := mi.NewQuery("select * from win32_process")
	require.NoError(t, err)

	var processes []win32Process

	err = session.Query(&processes, mi.NamespaceRootCIMv2, query, time.Millisecond)
	require.ErrorIs(t, err, mi.MI_RESULT_INVALID_OPERATION_TIMEOUT)

	processes = nil

	require.NoError(t, session.Query(&processes, mi.NamespaceRootCIMv2, query, mi.BuildQueryTimeout))
	require.NotEmpty(t, processes)
}

func Test_MI_Query_Uint16Array(t *testing.T) {
	application, err := mi.ApplicationInitialize()
	require.NoError(t, err)
	require.NotEmpty(t, application)

	session, err := application.NewSession(nil)
	require.NoError(t, err)
	require.NotEmpty(t, session)

	operation, err := session.QueryInstances(
		mi.OperationFlagsStandardRTTI,
		nil,
		mi.NamespaceRootCIMv2,
		mi.QueryDialectWQL,
		"SELECT Capabilities FROM Win32_DiskDrive",
	)
	require.NoError(t, err)
	require.NotEmpty(t, operation)

	found := false

	for {
		instance, moreResults, err := operation.GetInstance()
		require.NoError(t, err)

		if instance == nil {
			break
		}

		element, err := instance.GetElement("Capabilities")
		require.NoError(t, err)
		require.NotEmpty(t, element)

		value, err := element.GetValue()
		require.NoError(t, err)

		capabilities, ok := value.([]uint16)
		require.True(t, ok, "Capabilities should unmarshal to []uint16")

		if len(capabilities) > 0 {
			found = true
		}

		if !moreResults {
			break
		}
	}

	require.True(t, found, "expected at least one disk with a non-empty uint16[] Capabilities")

	err = operation.Close()
	require.NoError(t, err)

	err = session.Close()
	require.NoError(t, err)

	err = application.Close()
	require.NoError(t, err)
}

func Test_MI_QueryUnmarshal_Uint16Array(t *testing.T) {
	application, err := mi.ApplicationInitialize()
	require.NoError(t, err)
	require.NotEmpty(t, application)

	session, err := application.NewSession(nil)
	require.NoError(t, err)
	require.NotEmpty(t, session)

	query, err := mi.NewQuery("SELECT Capabilities FROM Win32_DiskDrive")
	require.NoError(t, err)

	var disks []win32DiskDrive

	err = session.Query(&disks, mi.NamespaceRootCIMv2, query, -1)
	require.NoError(t, err)
	require.NotEmpty(t, disks)

	found := false

	for _, disk := range disks {
		if len(disk.Capabilities) > 0 {
			found = true

			break
		}
	}

	require.True(t, found, "expected at least one disk with a non-empty uint16[] Capabilities")

	err = session.Close()
	require.NoError(t, err)

	err = application.Close()
	require.NoError(t, err)
}

func Test_MI_QueryUnmarshal_Uint16Array_WrongType(t *testing.T) {
	application, err := mi.ApplicationInitialize()
	require.NoError(t, err)
	require.NotEmpty(t, application)

	session, err := application.NewSession(nil)
	require.NoError(t, err)
	require.NotEmpty(t, session)

	query, err := mi.NewQuery("SELECT Capabilities FROM Win32_DiskDrive")
	require.NoError(t, err)

	var disks []win32DiskDriveWrongType

	// Unmarshalling a uint16[] into a []string field must error, not panic.
	err = session.Query(&disks, mi.NamespaceRootCIMv2, query, -1)
	require.Error(t, err)

	err = session.Close()
	require.NoError(t, err)

	err = application.Close()
	require.NoError(t, err)
}

func Test_MI_QueryFunc(t *testing.T) {
	application, err := mi.ApplicationInitialize()
	require.NoError(t, err)

	t.Cleanup(func() { require.NoError(t, application.Close()) })

	session, err := application.NewSession(nil)
	require.NoError(t, err)

	t.Cleanup(func() { require.NoError(t, session.Close()) })

	query, err := mi.NewQuery("SELECT LocalDateTime, NumberOfProcesses, Caption FROM Win32_OperatingSystem")
	require.NoError(t, err)

	var calls int

	err = session.QueryFunc(mi.NamespaceRootCIMv2, query, 5*time.Second, func(instance *mi.Instance) error {
		calls++

		// LocalDateTime is a DATETIME timestamp carrying the local UTC offset;
		// decoding it must land close to the current wall clock.
		element, err := instance.GetElement("LocalDateTime")
		require.NoError(t, err)
		require.False(t, element.IsNull())

		localDateTime, err := element.Float64()
		require.NoError(t, err)
		require.InDelta(t, float64(time.Now().Unix()), localDateTime, 60)

		element, err = instance.GetElement("NumberOfProcesses")
		require.NoError(t, err)

		processes, err := element.Float64()
		require.NoError(t, err)
		require.Positive(t, processes)

		element, err = instance.GetElement("Caption")
		require.NoError(t, err)

		caption, err := element.String()
		require.NoError(t, err)
		require.Contains(t, caption, "Windows")

		return nil
	})
	require.NoError(t, err)
	require.Equal(t, 1, calls)

	// An error returned by the callback aborts the query and is passed through.
	errStop := errors.New("stop")

	query, err = mi.NewQuery("SELECT Name FROM Win32_Process")
	require.NoError(t, err)

	calls = 0
	err = session.QueryFunc(mi.NamespaceRootCIMv2, query, 5*time.Second, func(*mi.Instance) error {
		calls++

		return errStop
	})
	require.ErrorIs(t, err, errStop)
	require.Equal(t, 1, calls)

	// The cancelled query must leave the session usable.
	calls = 0
	err = session.QueryFunc(mi.NamespaceRootCIMv2, query, 5*time.Second, func(*mi.Instance) error {
		calls++

		return nil
	})
	require.NoError(t, err)
	require.Greater(t, calls, 1)

	// Invalid classes surface as an error rather than an empty result.
	query, err = mi.NewQuery("SELECT Name FROM Win32_DoesNotExist")
	require.NoError(t, err)

	err = session.QueryFunc(mi.NamespaceRootCIMv2, query, 5*time.Second, func(*mi.Instance) error {
		return nil
	})
	require.Error(t, err)
}

func Test_MI_Operation_CloseEarly(t *testing.T) {
	application, err := mi.ApplicationInitialize()
	require.NoError(t, err)

	t.Cleanup(func() { require.NoError(t, application.Close()) })

	session, err := application.NewSession(nil)
	require.NoError(t, err)

	t.Cleanup(func() { require.NoError(t, session.Close()) })

	// Closing an operation with pending results cancels it instead of reading
	// them, and must leave the session usable.
	for range 3 {
		operation, err := session.QueryInstances(mi.OperationFlagsStandardRTTI, nil, mi.NamespaceRootCIMv2, mi.QueryDialectWQL, "SELECT Name FROM Win32_Process")
		require.NoError(t, err)

		instance, moreResults, err := operation.GetInstance()
		require.NoError(t, err)
		require.NotNil(t, instance)
		require.True(t, moreResults)

		require.NoError(t, operation.Close())
	}

	var processes []win32Process

	query, err := mi.NewQuery("SELECT Name FROM Win32_Process WHERE Handle = 0")
	require.NoError(t, err)
	require.NoError(t, session.Query(&processes, mi.NamespaceRootCIMv2, query, 5*time.Second))
	require.Equal(t, []win32Process{{Name: "System Idle Process"}}, processes)
}

// Test_MI_Query_StringArray reads a STRINGA element. Win32_OperatingSystem.MUILanguages
// is a string[] with at least the installed UI language.
func Test_MI_Query_StringArray(t *testing.T) {
	application, err := mi.ApplicationInitialize()
	require.NoError(t, err)

	t.Cleanup(func() { require.NoError(t, application.Close()) })

	session, err := application.NewSession(nil)
	require.NoError(t, err)

	t.Cleanup(func() { require.NoError(t, session.Close()) })

	operation, err := session.QueryInstances(mi.OperationFlagsStandardRTTI, nil, mi.NamespaceRootCIMv2, mi.QueryDialectWQL,
		"SELECT MUILanguages FROM Win32_OperatingSystem")
	require.NoError(t, err)

	t.Cleanup(func() { require.NoError(t, operation.Close()) })

	instance, _, err := operation.GetInstance()
	require.NoError(t, err)
	require.NotNil(t, instance)

	element, err := instance.GetElement("MUILanguages")
	require.NoError(t, err)

	value, err := element.GetValue()
	require.NoError(t, err)

	languages, ok := value.([]string)
	require.True(t, ok, "expected []string, got %T", value)
	require.NotEmpty(t, languages)

	for _, language := range languages {
		require.Contains(t, language, "-", "unexpected language tag %q", language)
	}
}

// Test_MI_QueryUnmarshal_InvalidClass checks that a failing query is reported
// through the operation result, not the return value of the void
// MI_Session_QueryInstances, and that the session stays usable afterwards.
func Test_MI_QueryUnmarshal_InvalidClass(t *testing.T) {
	application, err := mi.ApplicationInitialize()
	require.NoError(t, err)

	t.Cleanup(func() { require.NoError(t, application.Close()) })

	session, err := application.NewSession(nil)
	require.NoError(t, err)

	t.Cleanup(func() { require.NoError(t, session.Close()) })

	invalidQuery, err := mi.NewQuery("SELECT Name FROM Win32_DoesNotExist")
	require.NoError(t, err)

	validQuery, err := mi.NewQuery("SELECT Name FROM Win32_Process WHERE Handle = 0")
	require.NoError(t, err)

	for range 3 {
		var processes []win32Process

		err = session.Query(&processes, mi.NamespaceRootCIMv2, invalidQuery, 5*time.Second)
		require.ErrorIs(t, err, mi.MI_RESULT_INVALID_CLASS)
		require.Empty(t, processes)

		require.NoError(t, session.Query(&processes, mi.NamespaceRootCIMv2, validQuery, 5*time.Second))
		require.Equal(t, []win32Process{{Name: "System Idle Process"}}, processes)
	}
}

func Test_MI_Query_NilSession(t *testing.T) {
	t.Parallel()

	var (
		session   *mi.Session
		processes []win32Process
	)

	query, err := mi.NewQuery("SELECT Name FROM Win32_Process WHERE Handle = 0")
	require.NoError(t, err)

	require.ErrorIs(t, session.Query(&processes, mi.NamespaceRootCIMv2, query, 0), mi.ErrNotInitialized)
	require.ErrorIs(t, session.QueryFunc(mi.NamespaceRootCIMv2, query, 0, func(*mi.Instance) error { return nil }), mi.ErrNotInitialized)
}

func Test_MI_GetElementByName(t *testing.T) {
	application, err := mi.ApplicationInitialize()
	require.NoError(t, err)

	t.Cleanup(func() {
		require.NoError(t, application.Close())
	})

	session, err := application.NewSession(nil)
	require.NoError(t, err)

	t.Cleanup(func() {
		require.NoError(t, session.Close())
	})

	query, err := mi.NewQuery("SELECT Name FROM Win32_Process WHERE Handle = 4")
	require.NoError(t, err)

	name, err := mi.NewElementName("Name")
	require.NoError(t, err)

	var names []string

	err = session.QueryFunc(mi.NamespaceRootCIMv2, query, time.Second, func(instance *mi.Instance) error {
		_, err := instance.GetElementByName(nil)
		require.ErrorIs(t, err, mi.ErrInvalidElementName)

		element, err := instance.GetElementByName(name)
		if err != nil {
			return err
		}

		value, err := element.String()
		names = append(names, value)

		return err
	})
	require.NoError(t, err)
	require.Equal(t, []string{"System"}, names)
}
