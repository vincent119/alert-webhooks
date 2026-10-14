package watcher

import (
	"context"
	"sync"
	"testing"
	"time"
)

func TestConfigWatcherStopCancelsPendingDebounce(t *testing.T) {
	watcher := newTestConfigWatcher(t)
	var calls int
	watcher.reloadMain = func(context.Context) { calls++ }

	watcher.scheduleConfigChange("configs/config.development.yaml")
	watcher.Stop()

	if calls != 0 {
		t.Fatalf("停止後仍執行 %d 次主設定重載", calls)
	}
}

func TestConfigWatcherStopWaitsForRunningCallback(t *testing.T) {
	watcher := newTestConfigWatcher(t)
	started := make(chan struct{})
	finished := make(chan struct{})
	watcher.reloadMain = func(ctx context.Context) {
		close(started)
		<-ctx.Done()
		close(finished)
	}

	watcher.scheduleConfigChange("configs/config.development.yaml")
	select {
	case <-started:
	case <-time.After(time.Second):
		t.Fatal("debounce callback 未開始")
	}

	watcher.Stop()
	select {
	case <-finished:
	case <-time.After(time.Second):
		t.Fatal("Stop 未等待執行中的 callback 結束")
	}
}

func TestConfigWatcherDebouncesLatestEvent(t *testing.T) {
	watcher := newTestConfigWatcher(t)
	var mu sync.Mutex
	var calls []string
	watcher.reloadMain = func(context.Context) {
		mu.Lock()
		calls = append(calls, "main")
		mu.Unlock()
	}
	watcher.reloadTemplate = func(context.Context) {
		mu.Lock()
		calls = append(calls, "template")
		mu.Unlock()
	}

	watcher.scheduleConfigChange("configs/config.development.yaml")
	watcher.scheduleConfigChange("configs/alert_config.yaml")

	eventually(t, func() bool {
		mu.Lock()
		defer mu.Unlock()
		return len(calls) == 1
	})
	mu.Lock()
	defer mu.Unlock()
	if calls[0] != "template" {
		t.Fatalf("處理 %q，預期最後事件 template", calls[0])
	}
	watcher.Stop()
}

func TestConfigWatcherStopIsIdempotent(t *testing.T) {
	watcher := newTestConfigWatcher(t)
	watcher.Stop()
	watcher.Stop()
}

func newTestConfigWatcher(t *testing.T) *ConfigWatcher {
	t.Helper()
	watcher := NewConfigWatcher()
	watcher.debounce = time.Millisecond
	ctx, cancel := context.WithCancel(context.Background())
	watcher.mu.Lock()
	watcher.isRunning = true
	watcher.runCtx = ctx
	watcher.cancel = cancel
	watcher.mu.Unlock()
	t.Cleanup(watcher.Stop)
	return watcher
}

func eventually(t *testing.T, condition func() bool) {
	t.Helper()
	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		if condition() {
			return
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatal("條件未在期限內成立")
}
