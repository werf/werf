//go:build !windows

package tmp_manager

import (
	"os"
	"syscall"
)

type deviceFileInfo struct {
	os.FileInfo
	stat *syscall.Stat_t
}

func (i deviceFileInfo) Sys() any { return i.stat }
