//go:build !windows

package git

import "os/exec"

func hideConsole(*exec.Cmd) {}
