package build

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func newTestAIConveyor() *Conveyor {
	return &Conveyor{
		stageDigestMutex: map[string]*sync.Mutex{},
	}
}

func TestAI_ContentDigestMutex_SerializesSameDigest(t *testing.T) {
	c := newTestAIConveyor()
	const digest = "content-digest-x"

	mu := c.GetStageDigestMutex(digest)
	mu.Lock()

	var acquired atomic.Bool
	done := make(chan struct{})
	go func() {
		c.GetStageDigestMutex(digest).Lock()
		acquired.Store(true)
		c.GetStageDigestMutex(digest).Unlock()
		close(done)
	}()

	time.Sleep(20 * time.Millisecond)
	require.False(t, acquired.Load(), "second Lock on same digest must block while first is held")

	mu.Unlock()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("second goroutine did not acquire the digest mutex after release")
	}
	require.True(t, acquired.Load())
}

func TestAI_ContentDigestMutex_DifferentDigestsDoNotContend(t *testing.T) {
	c := newTestAIConveyor()

	c.GetStageDigestMutex("digest-a").Lock()
	defer c.GetStageDigestMutex("digest-a").Unlock()

	done := make(chan struct{})
	go func() {
		c.GetStageDigestMutex("digest-b").Lock()
		c.GetStageDigestMutex("digest-b").Unlock()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("Lock on a distinct digest must not block on another digest")
	}
}

func TestAI_DeferFnComposition_UnlocksAfterExistingCleanup(t *testing.T) {
	c := newTestAIConveyor()
	const digest = "content-digest-defer"

	var order []string
	prevDeferFn := func() {
		order = append(order, "cleanup")
	}

	c.GetStageDigestMutex(digest).Lock()
	deferFn := func() {
		if prevDeferFn != nil {
			prevDeferFn()
		}
		c.GetStageDigestMutex(digest).Unlock()
		order = append(order, "unlock")
	}

	deferFn()

	require.Equal(t, []string{"cleanup", "unlock"}, order,
		"deferFn must run existing cleanup before releasing the digest mutex")

	acquired := make(chan struct{})
	go func() {
		c.GetStageDigestMutex(digest).Lock()
		c.GetStageDigestMutex(digest).Unlock()
		close(acquired)
	}()

	select {
	case <-acquired:
	case <-time.After(time.Second):
		t.Fatal("deferFn did not release the digest mutex")
	}
}
