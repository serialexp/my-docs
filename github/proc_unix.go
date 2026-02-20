// ABOUTME: Unix-specific process attributes for detaching background clones.

//go:build !windows

package github

import "syscall"

func detachedProcAttr() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{Setsid: true}
}
