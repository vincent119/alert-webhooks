package service

import (
	"strings"
	"sync"

	"alert-webhooks/config"
	"alert-webhooks/pkg/logger"
	"alert-webhooks/pkg/notification"
	"alert-webhooks/pkg/template"
)

// ServiceManager 服務管理器
type ServiceManager struct {
	telegramService *TelegramService
	slackService    *SlackService
	discordService  *DiscordService
	templateEngine  *template.TemplateEngine
	mu              sync.RWMutex
}

var (
	instance *ServiceManager
	once     sync.Once
)

// GetServiceManager 獲取服務管理器單例
func GetServiceManager() *ServiceManager {
	once.Do(func() {
		instance = &ServiceManager{}
	})
	return instance
}

// InitServices 初始化所有服務
func (sm *ServiceManager) InitServices() error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	// 初始化模板引擎
	sm.initTemplateEngine()

	// 初始化 Telegram 服務（可選）
	if config.Telegram.Enable && config.Telegram.Token != "" {
		telegramService, err := NewTelegramService(config.Telegram.Token)
		if err != nil {
			logger.Error("Failed to initialize Telegram service", "service_manager", logger.Err(err))
			// 不返回錯誤，讓其他服務繼續運行
		} else {
			sm.telegramService = telegramService
			logger.Info("Telegram service initialized successfully", "service_manager")
		}
	} else {
		logger.Info("Telegram not enabled or token not configured, skipping Telegram service", "service_manager")
	}

	// 初始化 Slack 服務（可選）
	if config.Slack.Enable && config.Slack.Token != "" {
		slackService, err := NewSlackService(config.Slack.Token)
		if err != nil {
			logger.Error("Failed to initialize Slack service", "service_manager", logger.Err(err))
			// 不返回錯誤，讓其他服務繼續運行
		} else {
			sm.slackService = slackService
			logger.Info("Slack service initialized successfully", "service_manager")
		}
	} else {
		logger.Info("Slack not enabled or token not configured, skipping Slack service", "service_manager")
	}

	// 初始化 Discord 服務（可選）
	if config.Conf.Discord.Enable && config.Conf.Discord.Token != "" {
		discordService, err := NewDiscordService(config.Conf.Discord)
		if err != nil {
			logger.Error("Failed to initialize Discord service", "service_manager", logger.Err(err))
			// 不返回錯誤，讓其他服務繼續運行
		} else {
			sm.discordService = discordService
			logger.Info("Discord service initialized successfully", "service_manager")
		}
	} else {
		logger.Info("Discord not enabled or token not configured, skipping Discord service", "service_manager")
	}

	// 初始化通知管理器（在所有服務初始化完成後）
	notificationManager := notification.GetNotificationManager()
	if err := notificationManager.Initialize(sm.templateEngine, sm.telegramService, sm.slackService, sm.discordService); err != nil {
		logger.Error("Failed to initialize notification manager", "service_manager", logger.Err(err))
	} else {
		logger.Info("Notification manager initialized successfully", "service_manager")
	}

	logger.Info("Services initialization completed", "service_manager")
	return nil
}

// GetTelegramService 獲取 Telegram 服務
func (sm *ServiceManager) GetTelegramService() *TelegramService {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.telegramService
}

// IsTelegramServiceReady 檢查 Telegram 服務是否就緒
func (sm *ServiceManager) IsTelegramServiceReady() bool {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.telegramService != nil
}

// GetSlackService 獲取 Slack 服務
func (sm *ServiceManager) GetSlackService() *SlackService {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.slackService
}

// GetDiscordService 獲取 Discord 服務
func (sm *ServiceManager) GetDiscordService() *DiscordService {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.discordService
}

// IsSlackServiceReady 檢查 Slack 服務是否就緒
func (sm *ServiceManager) IsSlackServiceReady() bool {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.slackService != nil
}

// IsDiscordServiceReady 檢查 Discord 服務是否就緒
func (sm *ServiceManager) IsDiscordServiceReady() bool {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.discordService != nil
}

// GetTemplateEngine 獲取模板引擎
func (sm *ServiceManager) GetTemplateEngine() *template.TemplateEngine {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.templateEngine
}

// initTemplateEngine 初始化模板引擎
func (sm *ServiceManager) initTemplateEngine() {
	logger.Info("Initializing template engine", "service_manager")
	
	templateEngine := template.NewTemplateEngine()
	
	// 不再在初始化時選擇特定配置，而是加載默認的 full 配置
	// 每個平台的 handler 會根據自己的 template_mode 動態決定 FormatOptions
	logger.Info("Loading default template config (full mode)", "service_manager")
	
	err := templateEngine.LoadConfigFromConfigs()
	if err != nil {
		logger.Warn("Failed to load template config, using defaults", "service_manager", logger.Err(err))
	}
	
	// 載入模板檔案
	templatePaths := []string{
		"templates/alerts",
		"./templates/alerts",
		"../templates/alerts",
	}
	
	var loaded bool
	for _, path := range templatePaths {
		if err := templateEngine.LoadTemplates(path); err == nil {
			logger.Info("Templates loaded successfully", "service_manager",
				logger.String("template_path", path),
				logger.String("available_languages", strings.Join(templateEngine.GetAvailableLanguages(), ", ")))
			loaded = true
			break
		}
	}
	
	if !loaded {
		logger.Warn("Failed to load templates from all paths", "service_manager")
	}
	
	sm.templateEngine = templateEngine
	logger.Info("Template engine initialized successfully", "service_manager")
}

// ReloadTemplateEngine 重新載入模板引擎
func (sm *ServiceManager) ReloadTemplateEngine() {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	
	logger.Info("Reloading template engine", "service_manager")
	sm.initTemplateEngine()
	logger.Info("Template engine reloaded successfully", "service_manager")
}
