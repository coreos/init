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

package tests

import (
	"flag"
	"os"
	"path/filepath"
	"testing"

	"github.com/coreos/init/tests/coreos-install/register"
	"github.com/coreos/init/tests/coreos-install/util"

	_ "github.com/coreos/init/tests/coreos-install/registry"
)

var flagBinaryPath string

func init() {
	flag.StringVar(&flagBinaryPath, "coreos-install", "coreos-install", "path to coreos-install binary")
}

func TestMain(m *testing.M) {
	flag.Parse()
	os.Exit(m.Run())
}

func TestCoreosInstall(t *testing.T) {
	// download an image to speed up most tests
	localImagePath := util.FetchLocalImage(t)
	defer os.RemoveAll(localImagePath)

	server := util.HTTPServer{
		FileDir: localImagePath,
	}
	addr := server.Start(t)

	ctx := register.Context{
		BinaryPath:     flagBinaryPath,
		LocalImagePath: filepath.Join(localImagePath, "coreos_production_image.bin.bz2"),
		LocalAddress:   addr,
	}

	networkUnit := util.CreateNetworkUnit(t)
	if networkUnit != "" {
		defer os.RemoveAll(networkUnit)
	}

	for _, test := range register.Tests {
		t.Run(test.Name, func(t *testing.T) {
			test.Ctx = ctx
			test.Run(t)
		})
	}
}
