// Copyright 2024 The Prometheus Authors
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

package schedule_service

import (
	"errors"
	"fmt"
	"runtime"

	"github.com/go-ole/go-ole"
	"github.com/go-ole/go-ole/oleutil"
)

type ScheduleService struct {
	taskSchedulerObj *ole.IUnknown
	taskServiceObj   *ole.IDispatch
	taskService      *ole.VARIANT
}

func New() *ScheduleService {
	return &ScheduleService{}
}

func (s *ScheduleService) Connect() error {
	// The only way to run WMI queries in parallel while being thread-safe is to
	// ensure the CoInitialize[Ex]() call is bound to its current OS thread.
	// Otherwise, attempting to initialize and run parallel queries across
	// goroutines will result in protected memory errors.
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	if err := ole.CoInitializeEx(0, ole.COINIT_MULTITHREADED|ole.COINIT_DISABLE_OLE1DDE); err != nil {
		var oleCode *ole.OleError
		if errors.As(err, &oleCode) && oleCode.Code() != ole.S_OK && oleCode.Code() != 0x00000001 {
			return err
		}
	}

	scheduleClassID, err := ole.ClassIDFrom("Schedule.Service")
	if err != nil {
		return err
	}

	s.taskSchedulerObj, err = ole.CreateInstance(scheduleClassID, nil)
	if err != nil || s.taskSchedulerObj == nil {
		return err
	}

	s.taskService, err = oleutil.CallMethod(s.taskServiceObj, "Connect")
	if err != nil {
		return fmt.Errorf("failed to connect to task service: %w", err)
	}

	s.taskServiceObj = s.taskSchedulerObj.MustQueryInterface(ole.IID_IDispatch)

	return nil
}

func (s *ScheduleService) GetOLETaskServiceObj() *ole.IDispatch {
	return s.taskServiceObj
}

func (s *ScheduleService) Close() {
	if s.taskService != nil {
		_ = s.taskService.Clear()
	}

	if s.taskServiceObj != nil {
		s.taskServiceObj.Release()
	}

	if s.taskSchedulerObj != nil {
		s.taskSchedulerObj.Release()
	}

	ole.CoUninitialize()
}
