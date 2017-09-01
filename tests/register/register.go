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

package register

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/coreos/init/tests/util"
)

type Test struct {
	Name           string
	Func           func(*testing.T, Test)
	BinaryPath     string
	DiskSize       int64
	IgnitionConfig *string
	CloudConfig    *string
}

func (test Test) Run(t *testing.T) {
	originalTmpDir := os.Getenv("TMPDIR")
	defer os.Setenv("TMPDIR", originalTmpDir)

	if originalTmpDir == "" {
		err := os.Setenv("TMPDIR", "/var/tmp")
		if err != nil {
			t.Fatalf("couldn't set TMPDIR env var: %v", err)
		}
	}

	tmpDir, err := ioutil.TempDir("", "coreos-install-test")
	if err != nil {
		t.Fatalf("failed to create temp working dir in %s: %v", os.TempDir(), err)
	}
	defer test.RemoveAll(t, tmpDir)

	err = os.Setenv("TMPDIR", tmpDir)
	if err != nil {
		t.Fatalf("couldn't set TMPDIR env var: %v", err)
	}

	test.Func(t, test)
}

func (test Test) CreateDevice(t *testing.T) (string, string) {
	diskFile, err := os.Create(filepath.Join(os.TempDir(), "coreos-install-disk"))
	if err != nil {
		t.Fatalf("failed to create disk file: %v", err)
	}
	defer diskFile.Close()

	// default DiskSize to 10GB
	if test.DiskSize == 0 {
		test.DiskSize = 10 * 1024 * 1024 * 1024
	}

	err = os.Truncate(diskFile.Name(), test.DiskSize)
	if err != nil {
		t.Fatalf("failed to truncate disk file: %v", err)
	}

	// back a loop device with the disk file
	device := string(util.MustRun(t, "losetup", "-P", "-f", diskFile.Name(), "--show"))
	return diskFile.Name(), strings.TrimSpace(device)
}

func (test Test) CleanupDisk(t *testing.T, diskFile, loopDevice string) {
	util.MustRun(t, "losetup", "-d", loopDevice)
}

func (test Test) MountPartitions(t *testing.T, loopDevice string) string {
	root := filepath.Join(os.TempDir(), "root-mount-point")
	err := os.Mkdir(root, 0777)
	if err != nil {
		t.Fatalf("couldn't create root mount dir: %v", err)
	}

	util.MustRun(t, "mount", fmt.Sprintf("%sp9", loopDevice), root)
	util.MustRun(t, "mount", fmt.Sprintf("%sp1", loopDevice), filepath.Join(root, "boot"))
	util.MustRun(t, "mount", fmt.Sprintf("%sp3", loopDevice), filepath.Join(root, "usr"), "-o", "ro")
	util.MustRun(t, "mount", fmt.Sprintf("%sp6", loopDevice), filepath.Join(root, "usr", "share", "oem"))

	return root
}

func (test Test) UnmountPartitions(t *testing.T, loopDevice string) {
	util.MustRun(t, "umount", fmt.Sprintf("%sp6", loopDevice))
	util.MustRun(t, "umount", fmt.Sprintf("%sp3", loopDevice))
	util.MustRun(t, "umount", fmt.Sprintf("%sp1", loopDevice))
	util.MustRun(t, "umount", fmt.Sprintf("%sp9", loopDevice))
}

func (test Test) RunCoreOSInstall(t *testing.T, loopDevice string, opts ...string) {
	opts = append(opts, "-d", loopDevice)

	if test.IgnitionConfig != nil {
		ignitionPath := test.WriteFile(t, "coreos-ignition-file", *test.IgnitionConfig)
		opts = append(opts, "-i", ignitionPath)
	}

	if test.CloudConfig != nil {
		cloudinitPath := test.WriteFile(t, "coreos-cloudconfig-file", *test.CloudConfig)
		opts = append(opts, "-c", cloudinitPath)
	}

	util.MustRun(t, test.BinaryPath, opts...)
}

func (test Test) RemoveAll(t *testing.T, path string) {
	err := os.RemoveAll(path)
	if err != nil {
		t.Errorf("couldn't remove %s: %v", path, err)
	}
}

func (test Test) WriteFile(t *testing.T, name, data string) string {
	tmpFile, err := os.Create(filepath.Join(os.TempDir(), name))
	if err != nil {
		t.Fatalf("failed creating %s: %v", name, err)
	}
	defer tmpFile.Close()

	_, err = tmpFile.WriteString(data)
	if err != nil {
		t.Fatalf("writing to %s failed: %v", name, err)
	}

	return tmpFile.Name()
}

func (test Test) ValidateIgnition(t *testing.T, rootDir, config string) {
	oemPath := filepath.Join(rootDir, "usr", "share", "oem")

	data, err := ioutil.ReadFile(filepath.Join(oemPath, "coreos-install.json"))
	if os.IsNotExist(err) {
		t.Fatalf("couldn't find coreos-install.json")
	} else if err != nil {
		t.Fatalf("reading coreos-install.json: %v", err)
	}

	if string(data) != config {
		t.Fatalf("coreos-install.json doesn't match: expected %s, received %s", config, data)
	}

	data, err = ioutil.ReadFile(filepath.Join(oemPath, "grub.cfg"))
	if os.IsNotExist(err) {
		t.Fatalf("couldn't find grub.cfg")
	} else if err != nil {
		t.Fatalf("reading grub.cfg: %v", err)
	}

	util.RegexpContains(t, "ignition grub.cfg info", "coreos.config.url=oem:///coreos-install.json", data)
}

func (test Test) ValidateCloudConfig(t *testing.T, rootDir, config string) {
	data, err := ioutil.ReadFile(filepath.Join(rootDir, "var", "lib", "coreos-install", "user_data"))
	if os.IsNotExist(err) {
		t.Fatalf("couldn't find coreos-install/user_data")
	} else if err != nil {
		t.Fatalf("reading coreos-install/user_data: %v", err)
	}

	if string(data) != config {
		t.Fatalf("coreos-install/user_data doesn't match: expected %s, received %s", config, data)
	}
}

func (test Test) ValidateOSRelease(t *testing.T, rootDir string) {
	if _, err := os.Stat(filepath.Join(rootDir, "usr", "lib", "os-release")); os.IsNotExist(err) {
		t.Fatalf("/usr/lib/os-release was not found")
	} else if err != nil {
		t.Fatalf("running stat on /usr/lib/os-release: %v", err)
	}
}

func (test Test) DefaultChecks(t *testing.T, rootDir string) {
	test.ValidateOSRelease(t, rootDir)

	if test.IgnitionConfig != nil {
		test.ValidateIgnition(t, rootDir, *test.IgnitionConfig)
	}

	if test.CloudConfig != nil {
		test.ValidateCloudConfig(t, rootDir, *test.CloudConfig)
	}
}

var Tests []Test

func Register(t Test) {
	Tests = append(Tests, t)
}
