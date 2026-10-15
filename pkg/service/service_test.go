package service

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestServiceManagerRetiresBundleAfterLastLease(t *testing.T) {
	closed := make(chan struct{})
	old := testBundle(func() error {
		close(closed)
		return nil
	})
	manager := &ServiceManager{current: old}
	lease := manager.Acquire()
	if lease == nil {
		t.Fatal("expected old service lease to be available")
	}

	manager.swap(testBundle(nil))
	select {
	case <-closed:
		t.Fatal("old service was closed before the lease was released")
	case <-time.After(50 * time.Millisecond):
	}

	lease.Release()
	select {
	case <-closed:
	case <-time.After(time.Second):
		t.Fatal("old service was not closed after the lease was released")
	}

	if current := manager.Acquire(); current == nil {
		t.Fatal("new service lease should be available after swap")
	} else {
		current.Release()
	}
	if err := manager.Close(testContext(t)); err != nil {
		t.Fatalf("close service manager failed: %v", err)
	}
}

func TestServiceManagerFailedReloadKeepsCurrentBundle(t *testing.T) {
	old := testBundle(nil)
	manager := &ServiceManager{current: old}
	buildErr := errors.New("候選服務建立失敗")
	if err := manager.reload(func() (*serviceBundle, error) {
		return nil, buildErr
	}); !errors.Is(err, buildErr) {
		t.Fatalf("expected candidate creation error to be preserved, got: %v", err)
	}

	manager.mu.RLock()
	current := manager.current
	manager.mu.RUnlock()
	if current != old {
		t.Fatal("candidate creation failure must not replace the current service")
	}
	if err := manager.Close(testContext(t)); err != nil {
		t.Fatalf("close service manager failed: %v", err)
	}
}

func TestServiceBundleCloseOnlyRunsOnce(t *testing.T) {
	var calls atomic.Int32
	bundle := testBundle(func() error {
		calls.Add(1)
		return nil
	})
	if err := bundle.close(); err != nil {
		t.Fatalf("first close failed: %v", err)
	}
	if err := bundle.close(); err != nil {
		t.Fatalf("second close failed: %v", err)
	}
	if got := calls.Load(); got != 1 {
		t.Fatalf("close function ran %d times, want 1 times", got)
	}
}

func TestServiceManagerConcurrentAcquireAndSwap(t *testing.T) {
	manager := &ServiceManager{current: testBundle(nil)}
	var workers sync.WaitGroup
	stop := make(chan struct{})
	for range 8 {
		workers.Add(1)
		go func() {
			defer workers.Done()
			for {
				select {
				case <-stop:
					return
				default:
				}
				lease := manager.Acquire()
				if lease != nil {
					lease.Release()
				}
			}
		}()
	}

	for range 100 {
		manager.swap(testBundle(nil))
	}
	close(stop)
	workers.Wait()
	if err := manager.Close(testContext(t)); err != nil {
		t.Fatalf("close service manager after concurrent swap failed: %v", err)
	}
}

func testBundle(closeFn func() error) *serviceBundle {
	return &serviceBundle{
		drained: make(chan struct{}),
		closeFn: closeFn,
	}
}

func testContext(t *testing.T) context.Context {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	t.Cleanup(cancel)
	return ctx
}
