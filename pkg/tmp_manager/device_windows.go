package tmp_manager

import "os"

func sameDevice(_, _ os.FileInfo) bool {
	return true
}
