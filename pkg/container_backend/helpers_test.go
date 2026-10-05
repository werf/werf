package container_backend

import (
	"os"

	"github.com/onsi/gomega"
	"github.com/samber/lo"
)

func testChownableOwnership() (uint32, uint32) {
	uid, gid := os.Getuid(), os.Getgid()
	if uid == 0 {
		return 1001, 1001
	}

	groups, err := os.Getgroups()
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	if otherGid, found := lo.Find(groups, func(g int) bool { return g != gid }); found {
		gid = otherGid
	}

	return uint32(uid), uint32(gid)
}
