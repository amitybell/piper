//go:build darwin

package piper

import (
	"syscall"

	macos "github.com/nabbl/piper-bin-macos"
)

var (
	piperAsset = macos.Asset
	piperExe   = "piper"

	sysProcAttr *syscall.SysProcAttr
)
