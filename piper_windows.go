//go:build windows

package piper

import (
	"github.com/amitybell/piper-bin-windows"
	"syscall"
)

var (
	piperAsset = windows.Asset
	piperExe   = "piper.exe"

	sysProcAttr = &syscall.SysProcAttr{
		HideWindow:    true,
		CreationFlags: 0x08000000,
	}
)
