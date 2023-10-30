//go:build linux

package piper

import (
	"github.com/amitybell/piper-bin-linux"
	"syscall"
)

var (
	piperAsset = linux.Asset
	piperExe   = "piper"

	sysProcAttr *syscall.SysProcAttr
)
