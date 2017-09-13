// Copyright 2017 CoreOS, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package negative

import (
	"testing"

	"github.com/coreos/init/tests/register"
	"github.com/coreos/init/tests/util"
)

var (
	diskSize = "No space left on device"
)

func init() {
	register.Register(register.Test{
		Name:         "Disk Size too small - Local",
		Func:         installShouldFail,
		DiskSize:     2 * 1024 * 1024 * 1024,
		UseLocalFile: true,
		OutputRegexp: diskSize,
	})
	register.Register(register.Test{
		Name:           "Disk Size too small - Remote",
		Func:           installShouldFail,
		DiskSize:       2 * 1024 * 1024 * 1024,
		UseLocalServer: true,
		OutputRegexp:   diskSize,
	})
}

func installShouldFail(t *testing.T, test register.Test) {
	diskFile, loopDevice := test.CreateDevice(t)
	defer test.CleanupDisk(t, diskFile, loopDevice)

	out, err := test.RunCoreOSInstallNegative(t, loopDevice)
	if err == nil {
		t.Fatalf("install passed when it shouldn't have")
	}

	if !util.RegexpContains(t, test.OutputRegexp, out) {
		t.Fatalf("failed output validation: %s", out)
	}

	test.ValidatePartitionTableWiped(t, diskFile)
}
