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

package util

import (
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

func TryRegexpSearch(name, pattern string, data []byte) (string, error) {
	re := regexp.MustCompile(pattern)
	match := re.FindSubmatch(data)
	if len(match) < 2 {
		return "", fmt.Errorf("didn't find %s", name)
	}
	return string(match[1]), nil
}

func RegexpSearch(t *testing.T, itemName, pattern string, data []byte) string {
	result, err := TryRegexpSearch(itemName, pattern, data)
	if err != nil {
		t.Fatal(err)
	}
	return result
}

func RegexpContains(t *testing.T, pattern string, data []byte) bool {
	re := regexp.MustCompile(pattern)
	match := re.FindSubmatch(data)
	return len(match) > 0
}

func RegexpSearchAll(t *testing.T, itemName, pattern string, data []byte) (ret []string) {
	re := regexp.MustCompile(pattern)
	match := re.FindAllSubmatch(data, -1)
	if match == nil {
		t.Fatalf("couldn't find %s", itemName)
	}

	for _, m := range match {
		ret = append(ret, string(m[1]))
	}
	return
}

func MustRun(t *testing.T, command string, opts ...string) []byte {
	out, err := exec.Command(command, opts...).CombinedOutput()
	if err != nil {
		t.Log(string(out))
		t.Fatalf("%s %s failed: %v", command, strings.Join(opts, " "), err)
	}
	return out
}

func Run(t *testing.T, command string, opts ...string) error {
	_, err := exec.Command(command, opts...).CombinedOutput()
	return err
}

func StringToPtr(str string) *string {
	return &str
}

func FetchLocalImage(t *testing.T) string {
	tmpPath := os.Getenv("TMPDIR")
	if tmpPath == "" {
		tmpPath = "/var/tmp"
	}

	tmpDir, err := ioutil.TempDir(tmpPath, "")
	if err != nil {
		t.Fatalf("failed creating temp dir: %v", err)
	}

	err = DownloadFile(tmpDir, "coreos_production_image.bin.bz2")
	if err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("failed downloading image: %v", err)
	}

	err = DownloadFile(tmpDir, "coreos_production_image.bin.bz2.sig")
	if err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("failed downloading signature: %v", err)
	}

	err = DownloadFile(tmpDir, "version.txt")
	if err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("failed downloading version: %v", err)
	}

	return tmpDir
}

// Used to get defaults for channel, board, & version, first checks if the
// host machine is Container Linux and if so uses the data from the machine
// otherwise defaults to stable, amd64-usr, & current respectively
func GetDefaultChannelBoardVersion() (string, string, string, error) {
	data, err := ioutil.ReadFile("/usr/lib/os-release")
	if err != nil {
		return "stable", "amd64-usr", "current", nil
	}

	os, err := TryRegexpSearch("id", "ID=['\"]?([A-Za-z0-9 \\._\\-]*)['\"]?", data)
	if err != nil || os != "coreos" {
		return "stable", "amd64-usr", "current", nil
	}

	version, err := TryRegexpSearch("version", "VERSION_ID=['\"]?([A-Za-z0-9 \\._\\-]*)['\"]?", data)
	if err != nil {
		return "", "", "", err
	}

	board, err := TryRegexpSearch("board", "COREOS_BOARD=['\"]?([A-Za-z0-9 \\._\\-]*)['\"]?", data)
	if err != nil {
		return "", "", "", err
	}

	data, err = ioutil.ReadFile("/etc/coreos/update.conf")
	if err != nil {
		return "", "", "", fmt.Errorf("reading /etc/coreos/update.conf: %v", err)
	}

	channel, err := TryRegexpSearch("channel", "GROUP=['\"]?([A-Za-z0-9 \\._\\-]*)['\"]?", data)
	if err != nil {
		return "", "", "", err
	}

	return channel, board, version, nil
}

func DownloadFile(tmpDir, name string) error {
	file, err := os.Create(filepath.Join(tmpDir, name))
	if err != nil {
		return fmt.Errorf("failed to create file: %v", err)
	}
	defer file.Close()

	channel, board, version, err := GetDefaultChannelBoardVersion()
	if err != nil {
		return err
	}

	resp, err := http.Get(fmt.Sprintf("https://%s.release.core-os.net/%s/%s/%s", channel, board, version, name))
	if err != nil {
		return fmt.Errorf("failed to download file: %v", err)
	}
	defer resp.Body.Close()

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return fmt.Errorf("failed copying file data: %v", err)
	}

	return nil
}

// used to create a file in /run/systemd/network on the host machine if no
// files are present for use in network unit testing
func CreateNetworkUnit(t *testing.T) string {
	var ret string
	if _, err := os.Stat("/run/systemd/network"); os.IsNotExist(err) {
		err = os.MkdirAll("/run/systemd/network", 0777)
		if err != nil {
			t.Fatalf("creating /run/systemd/network: %v", err)
		}
		ret = "/run/systemd/network"
	} else {
		files, err := ioutil.ReadDir("/run/systemd/network")
		if err == nil && len(files) > 0 {
			// a network unit already exists
			return ""
		}
	}

	// no existing network files exist, write a valid .network file
	// which performs a no-op
	file, err := os.Create("/run/systemd/network/coreos-install-test.network")
	if err != nil {
		t.Fatalf("creating network unit: %v", err)
	}
	defer file.Close()

	_, err = file.WriteString(`# Created by coreos-install tests
[Match]
Architecture=coreos-install`)
	if err != nil {
		t.Fatalf("writing data to network unit: %v", err)
	}
	if ret == "" {
		ret = file.Name()
	}

	return ret
}

type HTTPServer struct {
	FileDir string
}

func (server *HTTPServer) Version(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, filepath.Join(server.FileDir, "version.txt"))
}

func (server *HTTPServer) Image(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, filepath.Join(server.FileDir, "coreos_production_image.bin.bz2"))
}

func (server *HTTPServer) Signature(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, filepath.Join(server.FileDir, "coreos_production_image.bin.bz2.sig"))
}

func (server *HTTPServer) Start(t *testing.T) string {
	http.HandleFunc("/current/version.txt", server.Version)

	data, err := ioutil.ReadFile(filepath.Join(server.FileDir, "version.txt"))
	if err != nil {
		t.Fatalf("Couldn't read version.txt")
	}
	version := RegexpSearch(t, "version", "COREOS_VERSION=(.*)", data)

	http.HandleFunc(fmt.Sprintf("/%s/coreos_production_image.bin.bz2", version), server.Image)
	http.HandleFunc(fmt.Sprintf("/%s/coreos_production_image.bin.bz2.sig", version), server.Signature)

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("creating listener: %v", err)
	}

	go http.Serve(listener, nil)

	return listener.Addr().String()
}
