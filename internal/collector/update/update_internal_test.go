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

//go:build windows && (amd64 || arm64)

package update

import (
	"bytes"
	"fmt"
	"log/slog"
	"testing"
	"unicode/utf16"
	"unsafe"

	"github.com/prometheus-community/windows_exporter/internal/ole"
	"github.com/prometheus-community/windows_exporter/internal/ole/wuapi"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/stretchr/testify/require"
	"golang.org/x/sys/windows"
)

func TestFetchUpdatesPartialCollection(t *testing.T) {
	for _, tc := range []struct {
		name       string
		countError bool
	}{
		{name: "skip unreadable item"},
		{name: "count failure", countError: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var logs bytes.Buffer

			logger := slog.New(slog.NewTextHandler(&logs, nil))
			c := &Collector{
				logger:                     logger,
				pendingUpdate:              prometheus.NewDesc("test_pending", "", []string{"id", "revision", "category", "severity", "title", "deployment_action"}, nil),
				pendingUpdateLastPublished: prometheus.NewDesc("test_published", "", []string{"id", "revision"}, nil),
				queryDurationSeconds:       prometheus.NewDesc("test_duration", "", nil, nil),
				lastScrapeMetric:           prometheus.NewDesc("test_timestamp", "", nil, nil),
			}

			searcher, attempted, criteria := fakeUpdateSearch(tc.countError)

			metrics, err := c.fetchUpdates(logger, searcher)
			if tc.countError {
				require.ErrorIs(t, err, ole.HRESULT(0x80070005))
				require.Nil(t, metrics)
				require.Empty(t, *attempted)

				return
			}

			require.NoError(t, err)
			require.Equal(t, searchCriteria, *criteria)
			require.Equal(t, []int32{0, 1, 2}, *attempted)
			require.Len(t, metrics, 6, "two readable updates retain info and publication metrics plus scrape metrics")

			ids := []string{}
			actions := []string{}

			for _, metric := range metrics {
				if metric.Desc() != c.pendingUpdate {
					continue
				}

				var sample dto.Metric
				require.NoError(t, metric.Write(&sample))

				for _, label := range sample.GetLabel() {
					switch label.GetName() {
					case "id":
						ids = append(ids, label.GetValue())
					case "deployment_action":
						actions = append(actions, label.GetValue())
					}
				}
			}

			require.Equal(t, []string{"update-0", "update-2"}, ids)
			require.Equal(t, []string{"installation", "optional_installation"}, actions)
			require.Contains(t, logs.String(), "failed to fetch Windows Update history item")
			require.Contains(t, logs.String(), "get collection item 1")
		})
	}
}

// fakeUpdateSearch implements only the native slots exercised by a scrape.
// Item 1 cannot be read; both surrounding items remain available. Item 0 is a
// regular update and item 2 an optional one. The search criteria are recorded.
func fakeUpdateSearch(countError bool) (*wuapi.UpdateSearcher, *[]int32, *string) {
	release := windows.NewCallback(func(uintptr) uintptr { return 0 })
	attempted := []int32{}

	var criteria string

	var selected int32

	var categoryMethods [10]uintptr

	categoryMethods[2] = release
	categoryMethods[9] = windows.NewCallback(func(_ uintptr, out *int32) uintptr {
		*out = 0

		return 0
	})
	categories := &wuapi.CategoryCollection{VTable: &categoryMethods[0]}

	var identityMethods [9]uintptr

	identityMethods[2] = release
	identityMethods[7] = windows.NewCallback(func(_ uintptr, out *int32) uintptr {
		*out = 1

		return 0
	})
	identityMethods[8] = windows.NewCallback(func(_ uintptr, out **uint16) uintptr {
		return fakeBSTR(fmt.Sprintf("update-%d", selected), out)
	})
	identity := &wuapi.UpdateIdentity{VTable: &identityMethods[0]}

	var updateMethods [49]uintptr

	updateMethods[2] = release
	updateMethods[7] = windows.NewCallback(func(_ uintptr, out **uint16) uintptr {
		return fakeBSTR("readable update", out)
	})
	updateMethods[11] = windows.NewCallback(func(_ uintptr, out **wuapi.CategoryCollection) uintptr {
		*out = categories

		return 0
	})
	updateMethods[19] = windows.NewCallback(func(_ uintptr, out **wuapi.UpdateIdentity) uintptr {
		*out = identity

		return 0
	})
	updateMethods[30] = windows.NewCallback(func(_ uintptr, out *ole.DATE) uintptr {
		*out = 25569 // Unix epoch.

		return 0
	})
	updateMethods[34] = windows.NewCallback(func(_ uintptr, out **uint16) uintptr {
		return fakeBSTR("Critical", out)
	})
	updateMethods[48] = windows.NewCallback(func(_ uintptr, out *wuapi.DeploymentAction) uintptr {
		*out = wuapi.DeploymentActionInstallation
		if selected == 2 {
			*out = wuapi.DeploymentActionOptionalInstallation
		}

		return 0
	})
	update := &wuapi.Update{VTable: &updateMethods[0]}

	var collectionMethods [11]uintptr

	collectionMethods[2] = release
	collectionMethods[7] = windows.NewCallback(func(_ uintptr, index uintptr, out **wuapi.Update) uintptr {
		selected = int32(index)

		attempted = append(attempted, selected)
		if selected == 1 {
			return 0x80070005
		}

		*out = update

		return 0
	})
	collectionMethods[10] = windows.NewCallback(func(_ uintptr, out *int32) uintptr {
		if countError {
			return 0x80070005
		}

		*out = 3

		return 0
	})
	updates := &wuapi.UpdateCollection{VTable: &collectionMethods[0]}

	var resultMethods [10]uintptr

	resultMethods[2] = release
	resultMethods[9] = windows.NewCallback(func(_ uintptr, out **wuapi.UpdateCollection) uintptr {
		*out = updates

		return 0
	})
	result := &wuapi.SearchResult{VTable: &resultMethods[0]}

	var searcherMethods [20]uintptr

	searcherMethods[2] = release
	searcherMethods[19] = windows.NewCallback(func(_ uintptr, input *uint16, out **wuapi.SearchResult) uintptr {
		criteria = windows.UTF16PtrToString(input)
		*out = result

		return 0
	})

	return &wuapi.UpdateSearcher{VTable: &searcherMethods[0]}, &attempted, &criteria
}

func fakeBSTR(value string, out **uint16) uintptr {
	units := append(utf16.Encode([]rune(value)), 0)

	ptr, _, _ := windows.NewLazySystemDLL("oleaut32.dll").NewProc("SysAllocStringLen").Call(
		uintptr(unsafe.Pointer(&units[0])),
		uintptr(len(units)-1),
	)
	if ptr == 0 {
		return 0x8007000e
	}

	*out = (*uint16)(unsafe.Pointer(ptr))

	return 0
}
