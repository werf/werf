//go:build !windows

package tmp_manager

import (
	"os"
	"syscall"
)

func sameDevice(a, b os.FileInfo) bool {
	statA, okA := a.Sys().(*syscall.Stat_t)
	statB, okB := b.Sys().(*syscall.Stat_t)
	if !okA || !okB {
		return true
	}
	return statA.Dev == statB.Dev
}
