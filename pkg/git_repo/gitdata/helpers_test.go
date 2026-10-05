package gitdata

import (
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/onsi/gomega"

	"github.com/werf/werf/v3/pkg/werf"
)

func createOrphanWorktreeFixture(kind, id, source string, accessed time.Time) string {
	dir := filepath.Join(werf.GetLocalCacheDir(), "worktrees", kind, id)
	gomega.Expect(os.MkdirAll(filepath.Join(dir, "worktree"), 0o755)).To(gomega.Succeed())
	gomega.Expect(os.WriteFile(filepath.Join(dir, "worktree", "file"), []byte("checkout"), 0o644)).To(gomega.Succeed())
	gomega.Expect(os.WriteFile(filepath.Join(dir, "last_access_at"), []byte(strconv.FormatInt(accessed.Unix(), 10)), 0o644)).To(gomega.Succeed())
	if source != "" {
		gomega.Expect(os.WriteFile(filepath.Join(dir, "git_dir"), []byte(source), 0o644)).To(gomega.Succeed())
	}
	return dir
}
