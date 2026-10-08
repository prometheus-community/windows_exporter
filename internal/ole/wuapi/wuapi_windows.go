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

// Package wuapi provides the Windows Update COM interfaces needed for update searches and metadata.
package wuapi

import (
	"iter"
	"unsafe"

	"github.com/prometheus-community/windows_exporter/internal/ole"
	"golang.org/x/sys/windows"
)

// Native method slots from wuapi.h, including inherited IDispatch methods.
// https://learn.microsoft.com/en-us/windows/win32/api/wuapi/
const (
	updateSessionPutApplicationID = 8
	updateSessionCreateSearcher   = 12
	updateSessionPutLocale        = 16 // IUpdateSession2
	updateSearcherSearch          = 19
	updateSearcherPutOnline       = 21
	updateSearcherHistoryCount    = 22
	searchResultUpdates           = 9
	updateCollectionItem          = 7
	updateCollectionCount         = 10
	updateTitle                   = 7
	updateCategories              = 11
	updateIdentity                = 19
	updateDeploymentTime          = 30
	updateSeverity                = 34
	identityRevision              = 7
	identityUpdateID              = 8
	categoryCollectionCount       = 9
	categoryName                  = 7
	categoryOrder                 = 12
)

type (
	UpdateSession  struct{ ole.Object } // IUpdateSession2
	UpdateSearcher struct{ ole.Object }
	SearchResult   struct{ ole.Object }
	Update         struct{ ole.Object }
	UpdateIdentity struct{ ole.Object }
	Category       struct{ ole.Object }
)

func NewUpdateSession() (*UpdateSession, error) {
	return ole.Create[UpdateSession](
		windows.GUID{Data1: 0x4cb43d7f, Data2: 0x7eee, Data3: 0x4906, Data4: [8]byte{0x86, 0x98, 0x60, 0xda, 0x1c, 0x38, 0xf2, 0xfe}},
		windows.GUID{Data1: 0x91caf7b0, Data2: 0xeb23, Data3: 0x49ed, Data4: [8]byte{0x99, 0x37, 0xc5, 0x2d, 0x81, 0x7f, 0x46, 0xf7}},
	)
}

func (s *UpdateSession) SetUserLocale(locale uint32) error {
	return s.Put(updateSessionPutLocale, uintptr(locale))
}

func (s *UpdateSession) SetClientApplicationID(id string) error {
	return s.PutString(updateSessionPutApplicationID, id)
}

func (s *UpdateSession) CreateUpdateSearcher() (*UpdateSearcher, error) {
	return s.Get[*UpdateSearcher](updateSessionCreateSearcher)
}

func (s *UpdateSearcher) SetOnline(online bool) error {
	var value uintptr
	if online {
		value = 0xffff
	} // VARIANT_TRUE is a signed 16-bit -1.

	return s.Put(updateSearcherPutOnline, value)
}

func (s *UpdateSearcher) GetTotalHistoryCount() (int32, error) {
	return s.Get[int32](updateSearcherHistoryCount)
}

func (s *UpdateSearcher) Search(criteria string) (*SearchResult, error) {
	return s.GetStringArg[*SearchResult](updateSearcherSearch, criteria)
}

func (s *SearchResult) Updates() (*updateCollection[Update], error) {
	return s.Get[*updateCollection[Update]](searchResultUpdates)
}

func (u *Update) Title() (string, error)        { return u.String(updateTitle) }
func (u *Update) MsrcSeverity() (string, error) { return u.String(updateSeverity) }
func (u *Update) Categories() (*updateCollection[Category], error) {
	return u.Get[*updateCollection[Category]](updateCategories)
}

func (u *Update) Identity() (*UpdateIdentity, error) { return u.Get[*UpdateIdentity](updateIdentity) }

func (u *Update) LastDeploymentChangeTime() (ole.DATE, error) {
	return u.Get[ole.DATE](updateDeploymentTime)
}
func (i *UpdateIdentity) UpdateID() (string, error)      { return i.String(identityUpdateID) }
func (i *UpdateIdentity) RevisionNumber() (int32, error) { return i.Get[int32](identityRevision) }
func (c *Category) Name() (string, error)                { return c.String(categoryName) }
func (c *Category) Order() (int32, error)                { return c.Get[int32](categoryOrder) }

type updateCollection[T Update | Category] struct{ ole.Object }

func (c *updateCollection[T]) count() (int32, error) {
	slot := uintptr(updateCollectionCount)
	if _, ok := any((*T)(nil)).(*Category); ok {
		slot = categoryCollectionCount
	}

	return c.Get[int32](slot)
}

func (c *updateCollection[T]) item(index int32) (*T, error) {
	// Both WUA collections use a LONG zero-based index and get_Item at slot 7.
	return c.GetArg[*T](updateCollectionItem, uintptr(index))
}

// All yields borrowed interfaces, valid only in the loop body. It releases each
// item on advance, early exit, or panic; callers must not release the items.
// The collection remains owned by the caller and must be released separately.
func (c *updateCollection[T]) All() iter.Seq2[*T, error] {
	return ole.BorrowedItems(c.count, c.item, func(item *T) { (*ole.Object)(unsafe.Pointer(item)).Release() })
}
