//go:build linux && !s390x && !arm && !386
// +build linux,!s390x,!arm,!386

// Copyright (c) 2015-2023 MinIO, Inc.
//
// This file is part of MinIO Object Storage stack
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package disk

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/prometheus/procfs/blockdevice"
	"golang.org/x/sys/unix"
)

// GetInfo returns total and free bytes available in a directory, e.g. `/`.
func GetInfo(path string, firstTime bool) (info Info, err error) {
	s := syscall.Statfs_t{}
	err = syscall.Statfs(path, &s)
	if err != nil {
		return Info{}, err
	}
	reservedBlocks := s.Bfree - s.Bavail
	info = Info{
		Total: uint64(s.Frsize) * (s.Blocks - reservedBlocks),
		Free:  uint64(s.Frsize) * s.Bavail,
		Files: s.Files,
		Ffree: s.Ffree,
		//nolint:unconvert
		FSType: getFSType(int64(s.Type)),
	}

	st := syscall.Stat_t{}
	err = syscall.Stat(path, &st)
	if err != nil {
		return Info{}, err
	}
	//nolint:unconvert
	devID := uint64(st.Dev) // Needed to support multiple GOARCHs
	info.Major = unix.Major(devID)
	info.Minor = unix.Minor(devID)

	// Check for overflows.
	// https://github.com/minio/minio/issues/8035
	// XFS can show wrong values at times error out
	// in such scenarios.
	if info.Free > info.Total {
		return info, fmt.Errorf("detected free space (%d) > total drive space (%d), fs corruption at (%s). please run 'fsck'", info.Free, info.Total, path)
	}
	info.Used = info.Total - info.Free

	if firstTime {
		bfs, err := blockdevice.NewDefaultFS()
		if err == nil {
			devName := ""
			diskstats, _ := bfs.ProcDiskstats()
			for _, dstat := range diskstats {
				// ignore all loop devices
				if strings.HasPrefix(dstat.DeviceName, "loop") {
					continue
				}
				if dstat.MajorNumber == info.Major && dstat.MinorNumber == info.Minor {
					devName = dstat.DeviceName
					break
				}
			}
			if devName != "" {
				info.Name = devName
				qst, err := bfs.SysBlockDeviceQueueStats(devName)
				if err != nil { // Mostly not found error
					// Check if there is a parent device:
					//   e.g. if the mount is based on /dev/nvme0n1p1, let's calculate the
					//        real device name (nvme0n1) to get its sysfs information
					parentDevPath, e := os.Readlink("/sys/class/block/" + devName)
					if e == nil {
						parentDev := filepath.Base(filepath.Dir(parentDevPath))
						qst, err = bfs.SysBlockDeviceQueueStats(parentDev)
					}
				}
				if err == nil {
					info.NRRequests = qst.NRRequests
					rot := qst.Rotational == 1 // Rotational is '1' if the device is HDD
					info.Rotational = &rot
				}
			}
		}
	}

	return info, nil
}
