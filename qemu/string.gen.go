// Copyright 2016 The go-qemu Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Code generated by "stringer -type=Status -output=string.gen.go"; DO NOT EDIT.

package qemu

import "fmt"

const _Status_name = "StatusDebugStatusInMigrateStatusInternalErrorStatusIOErrorStatusPausedStatusPostMigrateStatusPreLaunchStatusFinishMigrateStatusRestoreVMStatusRunningStatusSaveVMStatusShutdownStatusSuspendedStatusWatchdogStatusGuestPanicked"

var _Status_index = [...]uint8{0, 11, 26, 45, 58, 70, 87, 102, 121, 136, 149, 161, 175, 190, 204, 223}

func (i Status) String() string {
	if i < 0 || i >= Status(len(_Status_index)-1) {
		return fmt.Sprintf("Status(%d)", i)
	}
	return _Status_name[_Status_index[i]:_Status_index[i+1]]
}
