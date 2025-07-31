package true_git

import (
	"context"
	"fmt"
	"os"
	"sync"
)

var workTreePoolLimit = os.Getenv("WERF_GIT_WORK_TREE_POOL_LIMIT")

type WorkTreePool struct {
	slots    chan int
	baseDir  string
	poolSize int
}

type poolCache struct {
	mu    sync.RWMutex
	cache map[string]*WorkTreePool
}

var globalPoolCache = poolCache{
	cache: make(map[string]*WorkTreePool),
}

func GetWorkTreePool(baseDir string, poolSize int) (*WorkTreePool, error) {
	globalPoolCache.mu.RLock()
	wp, ok := globalPoolCache.cache[baseDir]
	globalPoolCache.mu.RUnlock()

	if ok {
		return wp, nil
	}

	globalPoolCache.mu.Lock()
	defer globalPoolCache.mu.Unlock()

	if wp, ok := globalPoolCache.cache[baseDir]; ok {
		return wp, nil
	}

	wp, err := NewWorkTreePool(baseDir, poolSize)
	if err != nil {
		return nil, err
	}

	globalPoolCache.cache[baseDir] = wp
	return wp, nil
}

func NewWorkTreePool(baseDir string, poolSize int) (*WorkTreePool, error) {
	if poolSize < 1 {
		return nil, fmt.Errorf("poolSize must be >= 1")
	}

	for i := 1; i < poolSize; i++ {
		dir := poolMemberDir(baseDir, i)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return nil, fmt.Errorf("create pool member dir %s: %w", dir, err)
		}
	}

	slots := make(chan int, poolSize)
	for i := 0; i < poolSize; i++ {
		slots <- i
	}

	return &WorkTreePool{
		slots:    slots,
		baseDir:  baseDir,
		poolSize: poolSize,
	}, nil
}

func (p *WorkTreePool) Acquire(ctx context.Context) (int, string, error) {
	select {
	case slot := <-p.slots:
		var dir string
		if slot == 0 {
			dir = p.baseDir
		} else {
			dir = poolMemberDir(p.baseDir, slot)
		}
		return slot, dir, nil
	case <-ctx.Done():
		return 0, "", ctx.Err()
	}
}

func (p *WorkTreePool) Release(slot int) {
	p.slots <- slot
}

func poolMemberDir(baseDir string, slot int) string {
	if slot == 0 {
		return baseDir
	}
	return fmt.Sprintf("%s-wtpool-%d", baseDir, slot)
}
