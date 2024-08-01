//go:build macos

package piper

import (
	"syscall"
)

var (
	piperAsset = macos.Asset
	piperExe   = "piper"

	sysProcAttr *syscall.SysProcAttr
)
