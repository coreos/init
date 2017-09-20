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

	"github.com/coreos/init/tests/coreos-install/register"
	"github.com/coreos/init/tests/coreos-install/util"
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
		UseLocalServer: true,
	})
	register.Register(register.Test{
		Name: "CloudConfig Test",
		Func: baseTest,
		CloudConfig: util.StringToPtr(`#cloud-config

hostname: "coreos1"`),
		UseLocalServer: true,
	})
	register.Register(register.Test{
		Name:    "Alpha 1520.0",
		Func:    baseTest,
		Channel: util.StringToPtr("alpha"),
		Version: util.StringToPtr("1520.0.0"),
	})
	register.Register(register.Test{
		Name:    "Channel Only",
		Func:    baseTest,
		Channel: util.StringToPtr("beta"),
	})
	register.Register(register.Test{
		Name:    "arm64-usr alpha 1367.5.0",
		Func:    baseTest,
		Channel: util.StringToPtr("alpha"),
		Version: util.StringToPtr("1367.5.0"),
		Board:   util.StringToPtr("arm64-usr"),
	})
	register.Register(register.Test{
		Name: "Version Only",
		Func: pickVersion,
	})
	register.Register(register.Test{
		Name: "OEM - ami",
		Func: baseTest,
		OEM:  util.StringToPtr("ami"),
	})
	register.Register(register.Test{
		Name: "OEM - cloudstack",
		Func: baseTest,
		OEM:  util.StringToPtr("cloudstack"),
	})
	register.Register(register.Test{
		Name: "OEM - digitalocean",
		Func: baseTest,
		OEM:  util.StringToPtr("digitalocean"),
	})
	register.Register(register.Test{
		Name: "OEM - packet",
		Func: baseTest,
		OEM:  util.StringToPtr("packet"),
	})
	register.Register(register.Test{
		Name: "OEM - rackspace",
		Func: baseTest,
		OEM:  util.StringToPtr("rackspace"),
	})
	register.Register(register.Test{
		Name: "OEM - vmware_raw",
		Func: baseTest,
		OEM:  util.StringToPtr("vmware_raw"),
	})
	register.Register(register.Test{
		Name:           "Network Units Test",
		Func:           baseTest,
		UseLocalServer: true,
		NetworkUnits:   true,
	})
}

// used by tests which want to test versions without pinning
// a specific channel as on Container Linux machines the
// channel will default to the host machines channel.
//
// Will update the test.Version flag with the selected Version
func pickVersion(t *testing.T, test register.Test) {
	pinnedVersions := map[string]string{
		"alpha":  "1535.0.0",
		"beta":   "1520.3.0",
		"stable": "1465.7.0",
	}

	channel, _, _, err := util.GetDefaultChannelBoardVersion()
	if err != nil {
		t.Fatal(err)
	}

	if version, ok := pinnedVersions[channel]; ok {
		test.Version = util.StringToPtr(version)
	} else {
		t.Fatalf("unknown channel %s", channel)
	}

	baseTest(t, test)
}

func baseTest(t *testing.T, test register.Test) {
	diskFile, loopDevice := test.CreateDevice(t)
	defer test.CleanupDisk(t, diskFile, loopDevice)

	test.RunCoreOSInstall(t, loopDevice)

	rootDir := test.MountPartitions(t, loopDevice)
	defer test.UnmountPartitions(t, loopDevice)

	test.DefaultChecks(t, rootDir)
}
