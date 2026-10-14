// Package watcher 提供設定檔變更的監聽與重載協調。
package watcher

import (
	"context"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"alert-webhooks/config"
	"alert-webhooks/pkg/logger"
	"alert-webhooks/pkg/service"

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
	debounceCancel context.CancelFunc
	debounceTimer  *time.Timer
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

	// 監控配置目錄和主要配置檔案
	configPaths := []string{
		"configs",
		"configs/config.development.yaml",
		"configs/alert_config.yaml",
		"configs/alert_config.minimal.yaml",
	}

	for _, path := range configPaths {
		if err := cw.watcher.Add(path); err != nil {
			logger.Warn("Failed to watch config path", "config_watcher",
				logger.String("path", path),
				logger.Err(err))
		} else {
			logger.Info("Watching config path", "config_watcher",
				logger.String("path", path))
		}
	}

	cw.isRunning = true
	cw.wg.Add(1)
	go cw.watchLoop(runCtx, watcher)

	logger.Info("Config watcher started", "config_watcher")
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
	debounceCancel := cw.debounceCancel
	debounceTimer := cw.debounceTimer
	cw.watcher = nil
	cw.cancel = nil
	cw.debounceCancel = nil
	cw.debounceTimer = nil
	cw.isRunning = false
	cw.mu.Unlock()

	if debounceCancel != nil {
		debounceCancel()
	}
	if debounceTimer != nil {
		debounceTimer.Stop()
	}
	if cancel != nil {
		cancel()
	}
	if watcher != nil {
		_ = watcher.Close()
	}
	cw.wg.Wait()

	logger.Info("Config watcher stopped", "config_watcher")
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
					logger.Info("Config file changed", "config_watcher",
						logger.String("file", event.Name),
						logger.String("operation", event.Op.String()))

					cw.scheduleConfigChange(event.Name)
				}
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			logger.Error("Config watcher error", "config_watcher", logger.Err(err))
		}
	}
}

func (cw *ConfigWatcher) scheduleConfigChange(filename string) {
	cw.mu.Lock()
	defer cw.mu.Unlock()
	if !cw.isRunning || cw.runCtx == nil {
		return
	}
	if cw.debounceCancel != nil {
		cw.debounceCancel()
	}
	if cw.debounceTimer != nil {
		cw.debounceTimer.Stop()
	}

	callbackCtx, cancel := context.WithCancel(cw.runCtx)
	timer := time.NewTimer(cw.debounce)
	cw.debounceCancel = cancel
	cw.debounceTimer = timer
	cw.wg.Add(1)
	go func() {
		defer cw.wg.Done()
		defer timer.Stop()
		select {
		case <-callbackCtx.Done():
			return
		case <-timer.C:
		}
		if callbackCtx.Err() != nil {
			return
		}
		cw.handleConfigChange(callbackCtx, filename)
	}()
}

// handleConfigChange 處理配置變更
func (cw *ConfigWatcher) handleConfigChange(ctx context.Context, filename string) {
	if ctx.Err() != nil {
		return
	}
	logger.Info("Handling config change", "config_watcher",
		logger.String("filename", filename))

	baseName := filepath.Base(filename)

	switch {
	case strings.HasPrefix(baseName, "config."):
		// 主配置檔案變更
		cw.reloadMain(ctx)
	case strings.HasPrefix(baseName, "alert_config"):
		// 警報模板配置變更
		cw.reloadTemplate(ctx)
	default:
		logger.Info("Unknown config file changed, skipping", "config_watcher",
			logger.String("filename", baseName))
	}
}

// reloadMainConfig 重新載入主配置
func (cw *ConfigWatcher) reloadMainConfig(ctx context.Context) {
	if ctx.Err() != nil {
		return
	}
	logger.Info("Reloading main config", "config_watcher")

	// 強制重新載入配置（繞過 sync.Once）
	logger.Info("Calling config.ForceReload()", "config_watcher")
	if ctx.Err() != nil {
		return
	}
	if err := config.ForceReload(); err != nil {
		logger.Error("Failed to reload main config", "config_watcher", logger.Err(err))
		return
	}
	logger.Info("Main config force reloaded", "config_watcher")

	if ctx.Err() != nil {
		return
	}
	if err := service.GetServiceManager().Reload(ctx); err != nil {
		logger.Error("Failed to reload services", "config_watcher", logger.Err(err))
		return
	}
	logger.Info("Main config and services reloaded successfully", "config_watcher")
}

// reloadTemplateConfig 重新載入模板配置
func (cw *ConfigWatcher) reloadTemplateConfig(ctx context.Context) {
	if ctx.Err() != nil {
		return
	}
	logger.Info("Reloading template config", "config_watcher")

	serviceManager := service.GetServiceManager()

	// 重新載入模板引擎
	if ctx.Err() != nil {
		return
	}
	serviceManager.ReloadTemplateEngine()

	logger.Info("Template config reloaded successfully", "config_watcher")
}

// IsRunning 檢查是否正在運行
func (cw *ConfigWatcher) IsRunning() bool {
	cw.mu.RLock()
	defer cw.mu.RUnlock()
	return cw.isRunning
}
