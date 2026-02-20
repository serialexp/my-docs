// ABOUTME: Windows-specific process attributes for detaching background clones.

//go:build windows

package github

import "syscall"

func detachedProcAttr() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{CreationFlags: 0x00000008} // CREATE_NO_WINDOW
}
