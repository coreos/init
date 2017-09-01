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
	"os/exec"
	"regexp"
	"strings"
	"testing"
)

func RegexpSearch(t *testing.T, itemName, pattern string, data []byte) string {
	re := regexp.MustCompile(pattern)
	match := re.FindSubmatch(data)
	if len(match) < 2 {
		t.Fatalf("couldn't find %s", itemName)
	}
	return string(match[1])
}

func RegexpContains(t *testing.T, itemName, pattern string, data []byte) bool {
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

func StringToPtr(str string) *string {
	return &str
}
