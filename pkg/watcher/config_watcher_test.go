package watcher

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/fsnotify/fsnotify"
)

func TestConfigWatcherStartStopAndIsRunning(t *testing.T) {
	watcher := NewConfigWatcher()
	if watcher.IsRunning() {
		t.Fatal("new watcher must not be running")
	}

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	if err := watcher.Start(ctx); err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	t.Cleanup(watcher.Stop)
	if !watcher.IsRunning() {
		t.Fatal("watcher must be running after Start")
	}
	if err := watcher.Start(ctx); err != nil {
		t.Fatalf("second Start() error = %v", err)
	}

	watcher.Stop()
	if watcher.IsRunning() {
		t.Fatal("watcher must not be running after Stop")
	}
}

func TestConfigWatcherStartWatchesExistingConfigPaths(t *testing.T) {
	workingDirectory, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd() error = %v", err)
	}
	repositoryRoot := filepath.Clean(filepath.Join(workingDirectory, "../.."))
	if err := os.Chdir(repositoryRoot); err != nil {
		t.Fatalf("Chdir() error = %v", err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(workingDirectory); err != nil {
			t.Errorf("restore working directory error = %v", err)
		}
	})

	watcher := NewConfigWatcher()
	if err := watcher.Start(context.Background()); err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	watcher.Stop()
}

func TestConfigWatcherProcessesYAMLFilesystemEvent(t *testing.T) {
	directory := t.TempDir()
	rawWatcher, err := fsnotify.NewWatcher()
	if err != nil {
		t.Fatalf("NewWatcher() error = %v", err)
	}
	if err := rawWatcher.Add(directory); err != nil {
		t.Fatalf("Add() error = %v", err)
	}

	watcher := NewConfigWatcher()
	called := make(chan struct{}, 1)
	watcher.reloadTemplate = func(context.Context) { called <- struct{}{} }
	ctx, cancel := context.WithCancel(context.Background())
	watcher.mu.Lock()
	watcher.watcher = rawWatcher
	watcher.isRunning = true
	watcher.runCtx = ctx
	watcher.cancel = cancel
	watcher.changes = make(chan string, 1)
	watcher.debounce = time.Millisecond
	watcher.wg.Add(2)
	watcher.mu.Unlock()
	go watcher.watchLoop(ctx, rawWatcher)
	go watcher.reloadLoop(ctx, watcher.changes, watcher.debounce)
	t.Cleanup(watcher.Stop)

	if err := os.WriteFile(filepath.Join(directory, "ignored.txt"), []byte("ignored"), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	if err := os.WriteFile(filepath.Join(directory, "alert_config.yaml"), []byte("enabled: true"), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	select {
	case <-called:
	case <-time.After(time.Second):
		t.Fatal("YAML change did not invoke template reload")
	}
}

func TestConfigWatcherWatchLoopHandlesClosedChannelsAndErrors(t *testing.T) {
	t.Run("events channel closed", func(t *testing.T) {
		rawWatcher := &fsnotify.Watcher{Events: make(chan fsnotify.Event), Errors: make(chan error)}
		watcher := NewConfigWatcher()
		watcher.wg.Add(1)
		close(rawWatcher.Events)
		go watcher.watchLoop(context.Background(), rawWatcher)
		watcher.wg.Wait()
	})

	t.Run("error then errors channel closed", func(t *testing.T) {
		rawWatcher := &fsnotify.Watcher{Events: make(chan fsnotify.Event), Errors: make(chan error, 1)}
		watcher := NewConfigWatcher()
		watcher.wg.Add(1)
		rawWatcher.Errors <- errors.New("watch failed")
		close(rawWatcher.Errors)
		go watcher.watchLoop(context.Background(), rawWatcher)
		watcher.wg.Wait()
	})
}

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

func TestConfigWatcherScheduleIgnoresInactiveOrCancelledWatcher(t *testing.T) {
	watcher := NewConfigWatcher()
	watcher.scheduleConfigChange("configs/config.development.yaml")

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	watcher.mu.Lock()
	watcher.isRunning = true
	watcher.runCtx = ctx
	watcher.changes = make(chan string, 1)
	watcher.mu.Unlock()
	watcher.scheduleConfigChange("configs/config.development.yaml")

	select {
	case filename := <-watcher.changes:
		t.Fatalf("scheduled %q after cancellation", filename)
	default:
	}
}

func TestConfigWatcherScheduleReplacesQueuedChange(t *testing.T) {
	watcher := NewConfigWatcher()
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	watcher.mu.Lock()
	watcher.isRunning = true
	watcher.runCtx = ctx
	watcher.changes = make(chan string, 1)
	watcher.changes <- "configs/config.development.yaml"
	watcher.mu.Unlock()

	watcher.scheduleConfigChange("configs/alert_config.yaml")

	select {
	case filename := <-watcher.changes:
		if filename != "configs/alert_config.yaml" {
			t.Fatalf("queued filename = %q", filename)
		}
	default:
		t.Fatal("latest change was not queued")
	}
}

func TestConfigWatcherHandlesKnownAndUnknownFiles(t *testing.T) {
	watcher := NewConfigWatcher()
	var mainCalls, templateCalls int
	watcher.reloadMain = func(context.Context) { mainCalls++ }
	watcher.reloadTemplate = func(context.Context) { templateCalls++ }

	ctx := context.Background()
	watcher.handleConfigChange(ctx, "/tmp/config.production.yaml")
	watcher.handleConfigChange(ctx, "/tmp/alert_config.minimal.yml")
	watcher.handleConfigChange(ctx, "/tmp/other.yaml")

	if mainCalls != 1 || templateCalls != 1 {
		t.Fatalf("reload calls = main:%d template:%d, want 1 each", mainCalls, templateCalls)
	}

	cancelled, cancel := context.WithCancel(context.Background())
	cancel()
	watcher.handleConfigChange(cancelled, "/tmp/config.production.yaml")
	if mainCalls != 1 {
		t.Fatalf("cancelled context triggered %d main reloads, want 1", mainCalls)
	}
}

func TestConfigWatcherReloadMainConfig(t *testing.T) {
	tests := []struct {
		name              string
		forceReloadErr    error
		reloadServicesErr error
		wantForceCalls    int
		wantServiceCalls  int
	}{
		{name: "成功", wantForceCalls: 1, wantServiceCalls: 1},
		{name: "設定重載失敗", forceReloadErr: errors.New("reload failed"), wantForceCalls: 1},
		{name: "服務重載失敗", reloadServicesErr: errors.New("service failed"), wantForceCalls: 1, wantServiceCalls: 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			watcher := NewConfigWatcher()
			var forceCalls, serviceCalls int
			watcher.forceReload = func() error {
				forceCalls++
				return tt.forceReloadErr
			}
			watcher.reloadServices = func(context.Context) error {
				serviceCalls++
				return tt.reloadServicesErr
			}

			watcher.reloadMainConfig(context.Background())

			if forceCalls != tt.wantForceCalls || serviceCalls != tt.wantServiceCalls {
				t.Fatalf("calls = force:%d service:%d, want force:%d service:%d", forceCalls, serviceCalls, tt.wantForceCalls, tt.wantServiceCalls)
			}
		})
	}
}

func TestConfigWatcherReloadFunctionsRespectCancellation(t *testing.T) {
	watcher := NewConfigWatcher()
	var forceCalls, serviceCalls, templateCalls int
	watcher.forceReload = func() error {
		forceCalls++
		return nil
	}
	watcher.reloadServices = func(context.Context) error {
		serviceCalls++
		return nil
	}
	watcher.reloadTemplates = func() { templateCalls++ }

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	watcher.reloadMainConfig(ctx)
	watcher.reloadTemplateConfig(ctx)

	if forceCalls != 0 || serviceCalls != 0 || templateCalls != 0 {
		t.Fatalf("cancelled reload calls = force:%d service:%d template:%d", forceCalls, serviceCalls, templateCalls)
	}
}

func TestConfigWatcherReloadMainStopsWhenCancelledAfterConfigReload(t *testing.T) {
	watcher := NewConfigWatcher()
	ctx, cancel := context.WithCancel(context.Background())
	watcher.forceReload = func() error {
		cancel()
		return nil
	}
	serviceCalls := 0
	watcher.reloadServices = func(context.Context) error {
		serviceCalls++
		return nil
	}

	watcher.reloadMainConfig(ctx)

	if serviceCalls != 0 {
		t.Fatalf("service reload calls = %d, want 0", serviceCalls)
	}
}

func TestConfigWatcherReloadTemplateConfig(t *testing.T) {
	watcher := NewConfigWatcher()
	calls := 0
	watcher.reloadTemplates = func() { calls++ }

	watcher.reloadTemplateConfig(context.Background())

	if calls != 1 {
		t.Fatalf("template reload calls = %d, want 1", calls)
	}
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
