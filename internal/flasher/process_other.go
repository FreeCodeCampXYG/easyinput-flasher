//go:build !windows

package flasher

import "syscall"

func hiddenWindowAttributes() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{Setpgid: true}
}
