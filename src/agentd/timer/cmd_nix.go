//go:build !windows
// +build !windows

package timer

import (
	"os/exec"
	"syscall"
)

func CmdStart(cmd *exec.Cmd) error {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	return cmd.Start()
}
