// Package watcher 提供設定檔變更的監聽與重載協調。
package watcher

import (
	"context"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"alert-webhooks/config"
	"alert-webhooks/pkg/service"
	"github.com/vincent119/zlogger"

	"github.com/fsnotify/fsnotify"
)

// ConfigWatcher 配置檔案監控器
type ConfigWatcher struct {
	watcher        *fsnotify.Watcher
	mu             sync.RWMutex
	isRunning      bool
	debounce       time.Duration
	runCtx         context.Context
	cancel         context.CancelFunc
	changes        chan string
	wg             sync.WaitGroup
	reloadMain     func(context.Context)
	reloadTemplate func(context.Context)
}

// NewConfigWatcher 創建新的配置監控器
func NewConfigWatcher() *ConfigWatcher {
	cw := &ConfigWatcher{debounce: time.Second}
	cw.reloadMain = cw.reloadMainConfig
	cw.reloadTemplate = cw.reloadTemplateConfig
	return cw
}

// Start 開始監控配置檔案
func (cw *ConfigWatcher) Start(ctx context.Context) error {
	cw.mu.Lock()
	defer cw.mu.Unlock()

	if cw.isRunning {
		return nil
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	runCtx, cancel := context.WithCancel(ctx)
	cw.watcher = watcher
	cw.runCtx = runCtx
	cw.cancel = cancel
	cw.changes = make(chan string, 1)

	// 監控配置目錄和主要配置檔案
	configPaths := []string{
		"configs",
		"configs/config.development.yaml",
		"configs/alert_config.yaml",
		"configs/alert_config.minimal.yaml",
	}

	for _, path := range configPaths {
		if err := cw.watcher.Add(path); err != nil {
			zlogger.Warn("Failed to watch config path", zlogger.String("category", "config_watcher"), zlogger.String("path", path),
				zlogger.Err(err))
		} else {
			zlogger.Info("Watching config path", zlogger.String("category", "config_watcher"), zlogger.String("path", path))
		}
	}

	cw.isRunning = true
	cw.wg.Add(2)
	go cw.watchLoop(runCtx, watcher)
	go cw.reloadLoop(runCtx, cw.changes, cw.debounce)

	zlogger.Info("Config watcher started", zlogger.String("category", "config_watcher"))
	return nil
}

// Stop 停止監控
func (cw *ConfigWatcher) Stop() {
	cw.mu.Lock()
	if !cw.isRunning {
		cw.mu.Unlock()
		return
	}

	watcher := cw.watcher
	cancel := cw.cancel
	cw.watcher = nil
	cw.cancel = nil
	cw.changes = nil
	cw.isRunning = false
	cw.mu.Unlock()

	if cancel != nil {
		cancel()
	}
	if watcher != nil {
		_ = watcher.Close()
	}
	cw.wg.Wait()

	zlogger.Info("Config watcher stopped", zlogger.String("category", "config_watcher"))
}

// watchLoop 監控循環
func (cw *ConfigWatcher) watchLoop(ctx context.Context, watcher *fsnotify.Watcher) {
	defer cw.wg.Done()

	for {
		select {
		case <-ctx.Done():
			return
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}

			// 只處理寫入和創建事件
			if event.Op&fsnotify.Write == fsnotify.Write || event.Op&fsnotify.Create == fsnotify.Create {
				// 只監控 YAML 檔案
				if strings.HasSuffix(event.Name, ".yaml") || strings.HasSuffix(event.Name, ".yml") {
					zlogger.Info("Config file changed", zlogger.String("category", "config_watcher"), zlogger.String("file", event.Name),
						zlogger.String("operation", event.Op.String()))

					cw.scheduleConfigChange(event.Name)
				}
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			zlogger.Error("Config watcher error", zlogger.String("category", "config_watcher"), zlogger.Err(err))
		}
	}
}

func (cw *ConfigWatcher) scheduleConfigChange(filename string) {
	cw.mu.RLock()
	if !cw.isRunning || cw.runCtx == nil {
		cw.mu.RUnlock()
		return
	}
	changes := cw.changes
	ctx := cw.runCtx
	cw.mu.RUnlock()

	select {
	case <-ctx.Done():
		return
	default:
	}

	select {
	case changes <- filename:
		return
	default:
	}

	select {
	case <-changes:
	default:
	}

	select {
	case changes <- filename:
	case <-ctx.Done():
	}
}

func (cw *ConfigWatcher) reloadLoop(ctx context.Context, changes <-chan string, debounce time.Duration) {
	defer cw.wg.Done()

	var (
		filename string
		timer    *time.Timer
		timerC   <-chan time.Time
	)
	defer func() {
		if timer != nil {
			timer.Stop()
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case filename = <-changes:
			if ctx.Err() != nil {
				return
			}
			if timer == nil {
				timer = time.NewTimer(debounce)
				timerC = timer.C
				continue
			}
			if !timer.Stop() {
				select {
				case <-timer.C:
				default:
				}
			}
			timer.Reset(debounce)
		case <-timerC:
			timer = nil
			timerC = nil
			if ctx.Err() != nil {
				return
			}
			cw.handleConfigChange(ctx, filename)
		}
	}
}

// handleConfigChange 處理配置變更
func (cw *ConfigWatcher) handleConfigChange(ctx context.Context, filename string) {
	if ctx.Err() != nil {
		return
	}
	zlogger.Info("Handling config change", zlogger.String("category", "config_watcher"), zlogger.String("filename", filename))

	baseName := filepath.Base(filename)

	switch {
	case strings.HasPrefix(baseName, "config."):
		// 主配置檔案變更
		cw.reloadMain(ctx)
	case strings.HasPrefix(baseName, "alert_config"):
		// 警報模板配置變更
		cw.reloadTemplate(ctx)
	default:
		zlogger.Info("Unknown config file changed, skipping", zlogger.String("category", "config_watcher"), zlogger.String("filename", baseName))
	}
}

// reloadMainConfig 重新載入主配置
func (cw *ConfigWatcher) reloadMainConfig(ctx context.Context) {
	if ctx.Err() != nil {
		return
	}
	zlogger.Info("Reloading main config", zlogger.String("category", "config_watcher"))

	// 強制重新載入配置（繞過 sync.Once）
	zlogger.Info("Calling config.ForceReload()", zlogger.String("category", "config_watcher"))
	if ctx.Err() != nil {
		return
	}
	if err := config.ForceReload(); err != nil {
		zlogger.Error("Failed to reload main config", zlogger.String("category", "config_watcher"), zlogger.Err(err))
		return
	}
	zlogger.Info("Main config force reloaded", zlogger.String("category", "config_watcher"))

	if ctx.Err() != nil {
		return
	}
	if err := service.GetServiceManager().Reload(ctx); err != nil {
		zlogger.Error("Failed to reload services", zlogger.String("category", "config_watcher"), zlogger.Err(err))
		return
	}
	zlogger.Info("Main config and services reloaded successfully", zlogger.String("category", "config_watcher"))
}

// reloadTemplateConfig 重新載入模板配置
func (cw *ConfigWatcher) reloadTemplateConfig(ctx context.Context) {
	if ctx.Err() != nil {
		return
	}
	zlogger.Info("Reloading template config", zlogger.String("category", "config_watcher"))

	serviceManager := service.GetServiceManager()

	// 重新載入模板引擎
	if ctx.Err() != nil {
		return
	}
	serviceManager.ReloadTemplateEngine()

	zlogger.Info("Template config reloaded successfully", zlogger.String("category", "config_watcher"))
}

// IsRunning 檢查是否正在運行
func (cw *ConfigWatcher) IsRunning() bool {
	cw.mu.RLock()
	defer cw.mu.RUnlock()
	return cw.isRunning
}
