package lock_manager

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/werf/common-go/pkg/locker_with_retry"
	"github.com/werf/lockgate"
	"github.com/werf/lockgate/pkg/distributed_locker"
)

const synchronizationRequestTimeout = 10 * time.Second

func NewHttp(ctx context.Context, address, clientID string) (Interface, error) {
	return newHTTPManager(ctx, address, clientID, nil), nil
}

func newHTTPManager(ctx context.Context, address, clientID string, fallback *httpFallbackState) Interface {
	backend := &httpBackend{ctx: ctx, endpoint: fmt.Sprintf("%s/%s/locker", address, clientID), client: &http.Client{Timeout: synchronizationRequestTimeout}}
	var locker lockgate.Locker = distributed_locker.NewDistributedLocker(backend)
	if fallback == nil {
		locker = locker_with_retry.NewLockerWithRetry(ctx, locker, locker_with_retry.LockerWithRetryOptions{MaxAcquireAttempts: maxAcquireAttempts, MaxReleaseAttempts: maxReleaseAttempts})
	}
	return &httpLockManager{manager: NewGeneric(locker), fallback: fallback}
}

type httpLockManager struct {
	manager  Interface
	fallback *httpFallbackState
}

var _ Interface = (*httpLockManager)(nil)

func (m *httpLockManager) LockStage(ctx context.Context, projectName, digest string) (LockHandle, error) {
	if err := ctx.Err(); err != nil {
		return LockHandle{}, err
	}
	if m.fallback != nil && m.fallback.disabled.Load() {
		return LockHandle{skipped: true}, nil
	}
	handle, err := m.manager.LockStage(ctx, projectName, digest)
	if err != nil && m.fallback != nil && m.fallback.disable(ctx, err) {
		return LockHandle{skipped: true}, nil
	}
	return handle, err
}

func (m *httpLockManager) Unlock(ctx context.Context, handle LockHandle) error {
	if handle.skipped {
		return nil
	}
	return m.manager.Unlock(ctx, handle)
}

type httpBackend struct {
	ctx      context.Context
	endpoint string
	client   *http.Client
}

var _ distributed_locker.DistributedLockerBackend = (*httpBackend)(nil)

func (b *httpBackend) Acquire(name string, opts distributed_locker.AcquireOptions) (lockgate.LockHandle, error) {
	request := distributed_locker.AcquireRequest{LockName: name, Opts: opts}
	var response distributed_locker.AcquireResponse
	if err := performPost(b.ctx, b.client, b.endpoint+"/acquire", request, &response); err != nil {
		return lockgate.LockHandle{}, err
	}
	if response.Err.Error != nil {
		return lockgate.LockHandle{}, response.Err.Error
	}
	if response.LockHandle.UUID == "" || response.LockHandle.LockName != name {
		return lockgate.LockHandle{}, fmt.Errorf("synchronization server returned an invalid lock handle")
	}
	return response.LockHandle, nil
}

func (b *httpBackend) RenewLease(handle lockgate.LockHandle) error {
	request := distributed_locker.RenewLeaseRequest{LockHandle: handle}
	var response distributed_locker.RenewLeaseResponse
	if err := performPost(b.ctx, b.client, b.endpoint+"/renew-lease", request, &response); err != nil {
		return err
	}
	return response.Err.Error
}

func (b *httpBackend) Release(handle lockgate.LockHandle) error {
	request := distributed_locker.ReleaseRequest{LockHandle: handle}
	var response distributed_locker.ReleaseResponse
	if err := performPost(context.WithoutCancel(b.ctx), b.client, b.endpoint+"/release", request, &response); err != nil {
		return err
	}
	return response.Err.Error
}
