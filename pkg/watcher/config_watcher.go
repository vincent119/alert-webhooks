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
	watcher   *fsnotify.Watcher
	mu        sync.RWMutex
	stopCh    chan struct{}
	isRunning bool
	debounce  time.Duration
}

// NewConfigWatcher 創建新的配置監控器
func NewConfigWatcher() *ConfigWatcher {
	return &ConfigWatcher{
		stopCh:   make(chan struct{}),
		debounce: 1 * time.Second, // 防抖動延遲 1 秒
	}
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
	cw.watcher = watcher

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

	// 啟動監控協程
	go cw.watchLoop(ctx)

	logger.Info("Config watcher started", "config_watcher")
	return nil
}

// Stop 停止監控
func (cw *ConfigWatcher) Stop() {
	cw.mu.Lock()
	defer cw.mu.Unlock()

	if !cw.isRunning {
		return
	}

	close(cw.stopCh)
	if cw.watcher != nil {
		cw.watcher.Close()
	}
	cw.isRunning = false

	logger.Info("Config watcher stopped", "config_watcher")
}

// watchLoop 監控循環
func (cw *ConfigWatcher) watchLoop(ctx context.Context) {
	var debounceTimer *time.Timer
	var debounceMu sync.Mutex

	for {
		select {
		case <-ctx.Done():
			return
		case <-cw.stopCh:
			return
		case event, ok := <-cw.watcher.Events:
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

					// 使用防抖動機制避免頻繁重載
					debounceMu.Lock()
					if debounceTimer != nil {
						debounceTimer.Stop()
					}
					debounceTimer = time.AfterFunc(cw.debounce, func() {
						cw.handleConfigChange(event.Name)
					})
					debounceMu.Unlock()
				}
			}
		case err, ok := <-cw.watcher.Errors:
			if !ok {
				return
			}
			logger.Error("Config watcher error", "config_watcher", logger.Err(err))
		}
	}
}

// handleConfigChange 處理配置變更
func (cw *ConfigWatcher) handleConfigChange(filename string) {
	logger.Info("Handling config change", "config_watcher",
		logger.String("filename", filename))

	baseName := filepath.Base(filename)

	switch {
	case strings.HasPrefix(baseName, "config."):
		// 主配置檔案變更
		cw.reloadMainConfig()
	case strings.HasPrefix(baseName, "alert_config"):
		// 警報模板配置變更
		cw.reloadTemplateConfig()
	default:
		logger.Info("Unknown config file changed, skipping", "config_watcher",
			logger.String("filename", baseName))
	}
}

// reloadMainConfig 重新載入主配置
func (cw *ConfigWatcher) reloadMainConfig() {
	logger.Info("Reloading main config", "config_watcher")

	// 強制重新載入配置（繞過 sync.Once）
	logger.Info("Calling config.ForceReload()", "config_watcher")
	config.ForceReload()
	logger.Info("Main config force reloaded", "config_watcher")

	// 重新初始化服務
	serviceManager := service.GetServiceManager()
	if err := serviceManager.InitServices(); err != nil {
		logger.Error("Failed to reinitialize services", "config_watcher", logger.Err(err))
		return
	}

	logger.Info("Main config reloaded successfully", "config_watcher")
}

// reloadTemplateConfig 重新載入模板配置
func (cw *ConfigWatcher) reloadTemplateConfig() {
	logger.Info("Reloading template config", "config_watcher")

	serviceManager := service.GetServiceManager()
	
	// 重新載入模板引擎
	serviceManager.ReloadTemplateEngine()
	
	logger.Info("Template config reloaded successfully", "config_watcher")
}

// IsRunning 檢查是否正在運行
func (cw *ConfigWatcher) IsRunning() bool {
	cw.mu.RLock()
	defer cw.mu.RUnlock()
	return cw.isRunning
}
