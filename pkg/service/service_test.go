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
		t.Fatal("預期可取得舊服務 lease")
	}

	manager.swap(testBundle(nil))
	select {
	case <-closed:
		t.Fatal("舊服務在 lease 釋放前即被關閉")
	case <-time.After(50 * time.Millisecond):
	}

	lease.Release()
	select {
	case <-closed:
	case <-time.After(time.Second):
		t.Fatal("lease 釋放後未關閉舊服務")
	}

	if current := manager.Acquire(); current == nil {
		t.Fatal("交換後應可取得新服務 lease")
	} else {
		current.Release()
	}
	if err := manager.Close(testContext(t)); err != nil {
		t.Fatalf("關閉服務管理器失敗: %v", err)
	}
}

func TestServiceManagerFailedReloadKeepsCurrentBundle(t *testing.T) {
	old := testBundle(nil)
	manager := &ServiceManager{current: old}
	buildErr := errors.New("候選服務建立失敗")
	if err := manager.reload(func() (*serviceBundle, error) {
		return nil, buildErr
	}); !errors.Is(err, buildErr) {
		t.Fatalf("預期保留候選建立錯誤，實際為: %v", err)
	}

	manager.mu.RLock()
	current := manager.current
	manager.mu.RUnlock()
	if current != old {
		t.Fatal("候選建立失敗後不應替換現有服務")
	}
	if err := manager.Close(testContext(t)); err != nil {
		t.Fatalf("關閉服務管理器失敗: %v", err)
	}
}

func TestServiceBundleCloseOnlyRunsOnce(t *testing.T) {
	var calls atomic.Int32
	bundle := testBundle(func() error {
		calls.Add(1)
		return nil
	})
	if err := bundle.close(); err != nil {
		t.Fatalf("第一次關閉失敗: %v", err)
	}
	if err := bundle.close(); err != nil {
		t.Fatalf("第二次關閉失敗: %v", err)
	}
	if got := calls.Load(); got != 1 {
		t.Fatalf("關閉函式執行 %d 次，預期 1 次", got)
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
		t.Fatalf("並行交換後關閉服務管理器失敗: %v", err)
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
