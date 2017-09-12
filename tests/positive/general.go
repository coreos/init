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

package positive

import (
	"testing"

	"github.com/coreos/init/tests/register"
	"github.com/coreos/init/tests/util"
)

func init() {
	register.Register(register.Test{
		Name: "Base Test",
		Func: baseTest,
	})
	register.Register(register.Test{
		Name: "Ignition Test",
		Func: baseTest,
		IgnitionConfig: util.StringToPtr(`{
			"ignition": {"version": "2.1.0"}
		}`),
	})
	register.Register(register.Test{
		Name: "CloudConfig Test",
		Func: baseTest,
		CloudConfig: util.StringToPtr(`#cloud-config

		hostname: "coreos1"`),
	})
}

func baseTest(t *testing.T, test register.Test) {
	diskFile, loopDevice := test.CreateDevice(t)
	defer test.CleanupDisk(t, diskFile, loopDevice)

	test.RunCoreOSInstall(t, loopDevice)

	rootDir := test.MountPartitions(t, loopDevice)
	defer test.UnmountPartitions(t, loopDevice)

	test.DefaultChecks(t, rootDir)
}
