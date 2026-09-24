package werf

import (
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"
)

func setupLocksDir(t *testing.T) string {
	t.Helper()
	home := t.TempDir()
	serviceDir = filepath.Join(home, "service")
	locks := getHostLockerDir()
	if err := os.MkdirAll(locks, 0o755); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { serviceDir = "" })
	return locks
}

func makeLock(t *testing.T, dir, name string, age time.Duration) {
	t.Helper()
	p := filepath.Join(dir, name)
	if err := os.WriteFile(p, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	mt := time.Now().Add(-age)
	if err := os.Chtimes(p, mt, mt); err != nil {
		t.Fatal(err)
	}
}

func TestGCHostLockerDir_RemovesStale(t *testing.T) {
	locks := setupLocksDir(t)
	makeLock(t, locks, "stale", HostLocksGCMinAge+time.Hour)
	makeLock(t, locks, "fresh", 0)

	if err := GCHostLockerDir(); err != nil {
		t.Fatalf("GC: %v", err)
	}
	if _, err := os.Stat(filepath.Join(locks, "stale")); !os.IsNotExist(err) {
		t.Fatalf("stale lock must be removed, stat err=%v", err)
	}
	if _, err := os.Stat(filepath.Join(locks, "fresh")); err != nil {
		t.Fatalf("fresh lock must be kept: %v", err)
	}
}

func TestShouldRunHostLocksGC_Threshold(t *testing.T) {
	locks := setupLocksDir(t)

	should, err := ShouldRunHostLocksGC()
	if err != nil {
		t.Fatal(err)
	}
	if should {
		t.Fatal("empty locks dir must not trigger GC")
	}

	for i := 0; i < hostLocksAutoGCThreshold; i++ {
		makeLock(t, locks, "stale"+strconv.Itoa(i), HostLocksGCMinAge+time.Hour)
	}

	should, err = ShouldRunHostLocksGC()
	if err != nil {
		t.Fatal(err)
	}
	if !should {
		t.Fatal("stale files at threshold must trigger GC")
	}
}

func TestShouldRunHostLocksGC_IgnoresFresh(t *testing.T) {
	locks := setupLocksDir(t)
	for i := 0; i < hostLocksAutoGCThreshold; i++ {
		makeLock(t, locks, "fresh"+strconv.Itoa(i), 0)
	}

	should, err := ShouldRunHostLocksGC()
	if err != nil {
		t.Fatal(err)
	}
	if should {
		t.Fatal("fresh files must not trigger GC")
	}
}

func TestShouldRunHostLocksGC_MissingDir(t *testing.T) {
	serviceDir = filepath.Join(t.TempDir(), "service")
	t.Cleanup(func() { serviceDir = "" })

	should, err := ShouldRunHostLocksGC()
	if err != nil {
		t.Fatalf("missing dir must not error: %v", err)
	}
	if should {
		t.Fatal("missing dir must not trigger GC")
	}
}
