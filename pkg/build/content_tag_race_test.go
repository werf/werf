package build

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func newTestConveyor() *Conveyor {
	return &Conveyor{
		stageDigestMutex: map[string]*sync.Mutex{},
	}
}

func TestContentDigestMutex_SameKeyReturnsSameMutex(t *testing.T) {
	c := newTestConveyor()

	a := c.GetStageDigestMutex("k1")
	b := c.GetStageDigestMutex("k1")
	require.Same(t, a, b, "same digest key must return the same mutex instance")

	other := c.GetStageDigestMutex("k2")
	require.NotSame(t, a, other, "distinct digest keys must return distinct mutex instances")
}

func TestContentDigestMutex_SerializesSameDigest(t *testing.T) {
	c := newTestConveyor()
	const digest = "content-digest-x"

	mu := c.GetStageDigestMutex(digest)
	mu.Lock()

	acquired := make(chan struct{})
	go func() {
		c.GetStageDigestMutex(digest).Lock()
		close(acquired)
		c.GetStageDigestMutex(digest).Unlock()
	}()

	mu.Unlock()
	select {
	case <-acquired:
	case <-time.After(time.Second):
		t.Fatal("second goroutine did not acquire the digest mutex after release")
	}
}

func TestContentDigestMutex_DifferentDigestsDoNotContend(t *testing.T) {
	c := newTestConveyor()

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

func TestDeferFnComposition_UnlocksAfterExistingCleanup(t *testing.T) {
	c := newTestConveyor()
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
