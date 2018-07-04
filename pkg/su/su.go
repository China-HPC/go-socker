package su

import (
	"bytes"
	"fmt"
	"os/exec"
	"syscall"

	"github.com/China-HPC/go-socker/pkg/user"
)

// Command creates a new exec.Cmd that will run with user privilege.
func Command(uid, command string, args ...string) (*exec.Cmd, error) {
	ucred, err := user.GetUserCredByUID(uid)
	if err != nil {
		return nil, err
	}
	cmd := exec.Command(command, args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{}
	cmd.SysProcAttr.Credential = ucred.Cred
	return cmd, nil
}

// Run creates and runs command with user privilege.
func Run(uid, command string, args ...string) error {
	cmd, err := Command(uid, command, args...)
	if err != nil {
		return err
	}
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("command(su %d) %s: %v: %s",
			uid, cmd.Path, err, stderr.String())
	}
	return nil
}

// Output creates and runs command with user privilege and returns the output.
func Output(uid, command string, args ...string) ([]byte, error) {
	cmd, err := Command(uid, command, args...)
	if err != nil {
		return nil, err
	}
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("command(su %d) %s: %v: %s (output: %s)",
			uid, cmd.Path, err, stderr.String(), string(out))
	}
	return out, nil
}

// CombinedOutput creates and runs command with user privilege and returns the
// combined output.
func CombinedOutput(uid, command string, args ...string) ([]byte, error) {
	cmd, err := Command(uid, command, args...)
	if err != nil {
		return nil, err
	}
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("command(su %d) %s: %v: %s",
			uid, cmd.Path, err, string(out))
	}
	return out, nil
}
