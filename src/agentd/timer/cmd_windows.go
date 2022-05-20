//go:build windows
// +build windows

package timer

func CmdStart(cmd *exec.Cmd) error {
	return cmd.Start()
}
