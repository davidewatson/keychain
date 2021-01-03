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

package command

import (
	"context"
	"log"
	"os"
	"os/exec"
	"strings"
	"text/template"
	"time"
)

// Command encapsulates a command to run.
type Command struct {
	Command string   // Name of command (relative or absolute)
	Args    []string // Slice of arguments for command
	Timeout int      // Number of seconds before process times out
}

// RunCommand runs command with arguments and a timeout. If the timeout expires
// context.DeadlineExceeded is returned. If there is no err, then stdout is
// return.
func RunCommand(ctx context.Context, command Command) ([]byte, error) {
	absPath, err := exec.LookPath(command.Command)
	if err != nil {
		log.Printf("didn't find %s executable", command.Command)
		return nil, err
	}

	// We use the context for the timeout and kill process functionality...
	newCtx, cancel := context.WithTimeout(ctx, time.Duration(command.Timeout)*time.Second)
	defer cancel()

	// TODO: Create the command with our context
	cmd := exec.CommandContext(newCtx, absPath, command.Args...)
	output, err := cmd.Output()

	// Check the context error to see if a timeout occurred. The error returned
	// by cmd.Output() will be OS specific based on what happens when a process
	// is killed.
	if newCtx.Err() == context.DeadlineExceeded {
		log.Printf("Command %s timed out\n", command.Command)
		return nil, newCtx.Err()
	}

	// If there's no context error, we know the command completed (or errored).
	if err != nil {
		log.Printf("Command %s returned non-zero, err %v, output %v\n", command.Command, err, output)
		return nil, err
	}

	return output, nil
}

// ProvisionServiceIdentityParams is used when templating ProvisionServiceIdentity commands
type ProvisionServiceIdentityParams struct {
	Algorithm string
	Days      int
	Subject   string
}

// ProvisionServiceIdentity shells out to create a certificate and returns it
func ProvisionServiceIdentity(ctx context.Context, params ProvisionServiceIdentityParams) ([]byte, error) {
	var buf strings.Builder

	tmpl := os.Getenv("GENERATE_CERT_COMMAND")
	t := template.Must(template.New("GetServiceIdentity").Parse(tmpl))
	t.Execute(&buf, params)

	cert, err := RunCommand(ctx, Command{Command: buf.String()})
	if err != nil {
		return nil, err
	}
	return cert, nil
}

// GetKeychainSecretParams is used when templating GetKeychainSecret commands
type GetKeychainSecretParams struct {
	Name  string
	Group string
}

// GetKeychainSecret shells out to get a Keychain secret and returns it
func GetKeychainSecret(ctx context.Context, params GetKeychainSecretParams) ([]byte, error) {
	var buf strings.Builder

	tmpl := os.Getenv("GET_SECRET_COMMAND")
	t := template.Must(template.New("GetServiceIdentity").Parse(tmpl))
	t.Execute(&buf, params)

	cert, err := RunCommand(ctx, Command{Command: buf.String()})
	if err != nil {
		return nil, err
	}
	return cert, nil
}

/*
// GetCertificateData reads filename from disk and returns a base64 encoded string
// https://kubernetes.io/docs/tasks/tls/managing-tls-in-a-cluster/
//env:
//- name: SSL_CERT_FILE
//  value: /etc/pki/tls/certs/thefacebook-ca-bundle.crt
func GetCertificateData(filename string) (map[string][]byte, error) {
	var secretData map[string][]byte

	envFilename := os.Getenv("KEYCHAIN_CERT")
	if envFilename != "" {
		filename = envFilename
	}

	_, err := ioutil.ReadFile(filename)
	if err != nil {
		return secretData, err
	}

	return secretData, nil
}
*/
