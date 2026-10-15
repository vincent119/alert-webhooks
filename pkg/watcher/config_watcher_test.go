package watcher

import (
	"context"
	"sync"
	"testing"
	"time"
)

func TestConfigWatcherStopCancelsPendingDebounce(t *testing.T) {
	watcher := newTestConfigWatcherWithDebounce(t, time.Second)
	var calls int
	watcher.reloadMain = func(context.Context) { calls++ }

	watcher.scheduleConfigChange("configs/config.development.yaml")
	watcher.Stop()

	if calls != 0 {
		t.Fatalf("executed after stop %d main configuration reloads", calls)
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
		t.Fatal("debounce callback did not start")
	}

	watcher.Stop()
	select {
	case <-finished:
	case <-time.After(time.Second):
		t.Fatal("Stop did not wait for the running callback to finish")
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
		t.Fatalf("handled %q, want last event template", calls[0])
	}
	watcher.Stop()
}

func TestConfigWatcherSerializesReloadCallbacks(t *testing.T) {
	watcher := newTestConfigWatcher(t)
	started := make(chan string, 2)
	releaseFirst := make(chan struct{})
	firstFinished := make(chan struct{})

	watcher.reloadMain = func(context.Context) {
		started <- "main"
		<-releaseFirst
		close(firstFinished)
	}
	watcher.reloadTemplate = func(context.Context) {
		started <- "template"
	}

	watcher.scheduleConfigChange("configs/config.development.yaml")
	select {
	case got := <-started:
		if got != "main" {
			t.Fatalf("first callback = %q, want main", got)
		}
	case <-time.After(time.Second):
		t.Fatal("first callback did not start")
	}

	watcher.scheduleConfigChange("configs/alert_config.yaml")
	select {
	case got := <-started:
		t.Fatalf("second callback started before first finished: %q", got)
	case <-time.After(20 * time.Millisecond):
	}

	close(releaseFirst)
	select {
	case <-firstFinished:
	case <-time.After(time.Second):
		t.Fatal("first callback did not finish")
	}
	select {
	case got := <-started:
		if got != "template" {
			t.Fatalf("second callback = %q, want template", got)
		}
	case <-time.After(time.Second):
		t.Fatal("second callback did not start")
	}
}

func TestConfigWatcherStopIsIdempotent(t *testing.T) {
	watcher := newTestConfigWatcher(t)
	watcher.Stop()
	watcher.Stop()
}

func newTestConfigWatcher(t *testing.T) *ConfigWatcher {
	return newTestConfigWatcherWithDebounce(t, time.Millisecond)
}

func newTestConfigWatcherWithDebounce(t *testing.T, debounce time.Duration) *ConfigWatcher {
	t.Helper()
	watcher := NewConfigWatcher()
	ctx, cancel := context.WithCancel(context.Background())
	watcher.mu.Lock()
	watcher.isRunning = true
	watcher.runCtx = ctx
	watcher.cancel = cancel
	watcher.changes = make(chan string, 1)
	watcher.debounce = debounce
	watcher.wg.Add(1)
	watcher.mu.Unlock()
	go watcher.reloadLoop(ctx, watcher.changes, debounce)
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
	t.Fatal("condition was not met before the deadline")
}
