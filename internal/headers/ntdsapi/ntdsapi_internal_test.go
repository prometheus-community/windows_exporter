// SPDX-License-Identifier: MIT
//
// Copyright The Prometheus Authors

//go:build windows

package ntdsapi

import (
	"testing"
	"time"
	"unsafe"

	"github.com/stretchr/testify/require"
	"golang.org/x/sys/windows"
)

func TestReplNeighborLayout(t *testing.T) {
	t.Parallel()

	// Offsets of DS_REPL_NEIGHBORW (ntdsapi.h).
	var n replNeighbor

	if unsafe.Sizeof(uintptr(0)) == 4 {
		require.Equal(t, uintptr(88), unsafe.Offsetof(n.usnLastObjChangeSynced))
		require.Equal(t, uintptr(128), unsafe.Sizeof(n))
	} else {
		require.Equal(t, uintptr(32), unsafe.Offsetof(n.dwReplicaFlags))
		require.Equal(t, uintptr(40), unsafe.Offsetof(n.uuidNamingContextObjGUID))
		require.Equal(t, uintptr(104), unsafe.Offsetof(n.usnLastObjChangeSynced))
		require.Equal(t, uintptr(120), unsafe.Offsetof(n.ftimeLastSyncSuccess))
		require.Equal(t, uintptr(128), unsafe.Offsetof(n.ftimeLastSyncAttempt))
		require.Equal(t, uintptr(136), unsafe.Offsetof(n.dwLastSyncResult))
		require.Equal(t, uintptr(140), unsafe.Offsetof(n.cNumConsecutiveSyncFailures))
		require.Equal(t, uintptr(144), unsafe.Sizeof(n))
	}
}

func TestParseNeighbors(t *testing.T) {
	t.Parallel()

	nc, err := windows.UTF16PtrFromString("DC=example,DC=com")
	require.NoError(t, err)

	dsa, err := windows.UTF16PtrFromString("CN=NTDS Settings,CN=DC02,CN=Servers,CN=HQ,CN=Sites,CN=Configuration,DC=example,DC=com")
	require.NoError(t, err)

	success := time.Date(2026, 11, 22, 21, 0, 0, 0, time.UTC)

	// DS_REPL_NEIGHBORSW with two entries, as one buffer.
	buffer := struct {
		header  replNeighbors
		entries [2]replNeighbor
	}{
		header: replNeighbors{cNumNeighbors: 2},
		entries: [2]replNeighbor{
			{
				pszNamingContext:            nc,
				pszSourceDsaDN:              dsa,
				dwReplicaFlags:              0x10,
				ftimeLastSyncSuccess:        windows.NsecToFiletime(success.UnixNano()),
				ftimeLastSyncAttempt:        windows.NsecToFiletime(success.Add(time.Hour).UnixNano()),
				dwLastSyncResult:            8524,
				cNumConsecutiveSyncFailures: 2,
			},
			{pszNamingContext: nc},
		},
	}

	require.Equal(t, uintptr(replNeighborsArrayOffset), unsafe.Offsetof(buffer.entries))

	neighbors, err := parseNeighbors(&buffer.header)
	require.NoError(t, err)
	require.Equal(t, []Neighbor{
		{
			NamingContext:           "DC=example,DC=com",
			SourceDsaDN:             "CN=NTDS Settings,CN=DC02,CN=Servers,CN=HQ,CN=Sites,CN=Configuration,DC=example,DC=com",
			ReplicaFlags:            0x10,
			LastSyncSuccess:         success,
			LastSyncAttempt:         success.Add(time.Hour),
			LastSyncResult:          8524,
			ConsecutiveSyncFailures: 2,
		},
		{NamingContext: "DC=example,DC=com"},
	}, neighbors)

	empty, err := parseNeighbors(&replNeighbors{})
	require.NoError(t, err)
	require.Empty(t, empty)
}
