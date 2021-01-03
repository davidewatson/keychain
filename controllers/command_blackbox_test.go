/*
Copyright (c) 2020 Facebook, Inc. and its affiliates.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers_test

import (
	"context"
	"os/exec"
	"reflect"
	"testing"
	"time"

	. "github.com/davidewatson/keychain/controllers"
)

const (
	defaultTimeout = 10 * time.Second
)

func TestRunCommand(t *testing.T) {
	var testsTable = []struct {
		name    string
		command string
		args    []string
		timeout time.Duration
		err     error
	}{
		{name: "relative paths work", command: "ls", args: nil, timeout: time.Second, err: nil},
		{name: "absolute paths work", command: "/bin/ls", args: nil, timeout: time.Second, err: nil},
		{name: "errors are propagated", command: "false", args: nil, timeout: time.Second, err: &exec.ExitError{}},
		{name: "timeouts kill", command: "sleep", args: []string{"1"}, timeout: 0 * time.Second, err: context.DeadlineExceeded},
	}

	for _, tt := range testsTable {
		t.Run(tt.name, func(t *testing.T) {
			_, err := RunCommand(context.Background(),
				Command{
					Command: tt.command,
					Args:    tt.args,
					Timeout: tt.timeout,
				})

			if reflect.TypeOf(err) != reflect.TypeOf(tt.err) {
				t.Errorf("Error observed %v, expected %v", reflect.TypeOf(err), reflect.TypeOf(tt.err))
			}
		})
	}
}
