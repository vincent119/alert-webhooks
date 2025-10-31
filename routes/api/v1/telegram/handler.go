// Package telegram 提供 Telegram 相關的路由處理功能，包括訊息發送和機器人資訊查詢
package telegram

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"alert-webhooks/config"
	"alert-webhooks/pkg/alertmodel"
	"alert-webhooks/pkg/logger"
	"alert-webhooks/pkg/service"
	"alert-webhooks/pkg/template"

	"github.com/gin-gonic/gin"
	"github.com/go-telegram/bot/models"
)

// Handler Telegram 路由處理器
type Handler struct {
	telegramService *service.TelegramService
	templateEngine  *template.TemplateEngine
}

// NewHandler 創建新的 Telegram 路由處理器
func NewHandler(telegramService *service.TelegramService) *Handler {
	logger.Info("Creating new Telegram handler", "telegram_handler")

	// 從 ServiceManager 獲取模板引擎
	serviceManager := service.GetServiceManager()
	templateEngine := serviceManager.GetTemplateEngine()

	if templateEngine == nil {
		logger.Error("Template engine not available from service manager", "telegram_handler")
		return &Handler{
			telegramService: telegramService,
			templateEngine:  template.NewTemplateEngine(), // 後備方案
		}
	}

	logger.Info("Template engine obtained from service manager", "telegram_handler")



	// 模板引擎現在由 ServiceManager 管理，不需要在 handler 中初始化
	handler := &Handler{
		telegramService: telegramService,
		templateEngine:  templateEngine,
	}

	logger.Info("Telegram handler created", "telegram_handler",
		logger.Bool("has_template_engine", templateEngine != nil))

	return handler
}

// SendMessageRequest send message request structure
type SendMessageRequest struct {
	Message           string `json:"message"` // 可選，當使用 AlertManager 格式時不需要
	TemplateLanguage  string `json:"template_language"` // 可選，優先使用配置檔案中的設定
	AlertManagerData  *AlertManagerWebhook `json:"alertmanager_data"` // AlertManager webhook 數據
}

// AlertManagerWebhook AlertManager webhook 結構（簡化版）
type AlertManagerWebhook struct {
	Receiver          string            `json:"receiver"`
	Status            string            `json:"status"`
	Alerts            []Alert           `json:"alerts"`
	GroupLabels       map[string]string `json:"groupLabels"`
	CommonLabels      map[string]string `json:"commonLabels"`
	CommonAnnotations map[string]string `json:"commonAnnotations"`
	ExternalURL       string            `json:"externalURL"`
	Version           string            `json:"version"`
	GroupKey          string            `json:"groupKey"`
	TruncatedAlerts   int               `json:"truncatedAlerts"`
}

// Alert 單個警報結構
type Alert struct {
	Status       string            `json:"status"`
	Labels       map[string]string `json:"labels"`
	Annotations  map[string]string `json:"annotations"`
	StartsAt     string            `json:"startsAt"`
	EndsAt       string            `json:"endsAt"`
	GeneratorURL string            `json:"generatorURL"`
	Fingerprint  string            `json:"fingerprint"`
}

// SendMessageResponse send message response structure
type SendMessageResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Level   int    `json:"level"`
}

// SendMessage send Telegram message
// @Summary Send Telegram message
// @Description Send message to specified Telegram chat level
// @Tags telegram
// @Accept json
// @Produce json
// @Param chatid path string true "聊天等級 (格式: L{0-4})"
// @Param request body SendMessageRequest true "訊息內容"
// @Security BasicAuth
// @Success 200 {object} SendMessageResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /telegram/chatid_{chatid} [post]
func (h *Handler) SendMessage(c *gin.Context) {
	// 從 URL 參數獲取 chatid
	chatIDParam := c.Param("chatid")

	// 驗證 chatid 格式 (L{0-4} 或 {0-4})
	var levelStr string
	if strings.HasPrefix(chatIDParam, "L") {
		levelStr = strings.TrimPrefix(chatIDParam, "L")
	} else {
		levelStr = chatIDParam
	}

	// 驗證等級範圍 (0-4)
	level, err := strconv.Atoi(levelStr)
	if err != nil || level < 0 || level > 4 {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid chatid format. Must be 0-4 or L0-L4",
		})
		return
	}

	// 從請求體獲取訊息內容
	var req SendMessageRequest

	// 先嘗試解析為標準格式
	body, err := c.GetRawData()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Failed to read request body: " + err.Error(),
		})
		return
	}

	// Debug 模式下記錄接收到的 JSON 數據
	// 檢查是否為開發環境或 debug 模式，並且日誌級別允許 debug 輸出
	isDev := config.IsDevelopment()
	isDebugMode := strings.ToLower(config.App.Mode) == "debug"
	isDebugLevel := strings.ToLower(config.Log.Level) == "debug"

	// 先記錄條件判斷的結果用於調試
	logger.Info("Debug condition check", "telegram_handler",
		logger.Bool("is_development", isDev),
		logger.Bool("is_debug_mode", isDebugMode),
		logger.Bool("is_debug_level", isDebugLevel),
		logger.String("app_mode", config.App.Mode),
		logger.String("log_level", config.Log.Level))

	// 如果滿足任一條件就輸出 debug 日誌
	if isDev || isDebugMode || isDebugLevel {
		logger.Debug("Received JSON request body", "telegram_handler",
			logger.String("chatid", chatIDParam),
			logger.String("level", levelStr),
			logger.String("raw_body", string(body)),
			logger.String("content_type", c.GetHeader("Content-Type")),
			logger.String("user_agent", c.GetHeader("User-Agent")),
			logger.String("remote_addr", c.ClientIP()))
	}

	// 嘗試解析為標準格式
	if parseErr := c.ShouldBindJSON(&req); parseErr != nil {
		// Debug 模式下記錄標準格式解析失敗的情況
		if isDev || isDebugMode || isDebugLevel {
			logger.Debug("Standard format parsing failed, trying AlertManager format", "telegram_handler",
				logger.String("parse_error", parseErr.Error()))
		}

		// 如果標準格式解析失敗，嘗試解析為直接的 AlertManager 格式
		var alertManagerData AlertManagerWebhook
		if unmarshalErr := json.Unmarshal(body, &alertManagerData); unmarshalErr != nil {
			// Debug 模式下記錄 AlertManager 格式解析失敗的情況
			if isDev || isDebugMode || isDebugLevel {
				logger.Debug("AlertManager format parsing also failed", "telegram_handler",
					logger.String("unmarshal_error", unmarshalErr.Error()),
					logger.String("body_preview", getBodyPreview(body)))
			}

			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "Invalid request body format: " + unmarshalErr.Error(),
			})
			return
		}

		// Debug 模式下記錄成功解析為 AlertManager 格式
		if isDev || isDebugMode || isDebugLevel {
			logger.Debug("Successfully parsed as AlertManager format", "telegram_handler",
				logger.String("status", alertManagerData.Status),
				logger.Int("alerts_count", len(alertManagerData.Alerts)),
				logger.String("receiver", alertManagerData.Receiver))
		}

		// 直接設置 AlertManager 數據
		req.AlertManagerData = &alertManagerData
	} else {
		// Debug 模式下記錄成功解析為標準格式
		if isDev || isDebugMode || isDebugLevel {
			logger.Debug("Successfully parsed as standard format", "telegram_handler",
				logger.String("message_length", fmt.Sprintf("%d", len(req.Message))),
				logger.String("template_language", req.TemplateLanguage),
				logger.Bool("has_alertmanager_data", req.AlertManagerData != nil))
		}
	}

	// 驗證訊息內容 - 允許 AlertManager 數據或普通訊息
	if req.Message == "" && req.AlertManagerData == nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Either message or alertmanager_data must be provided",
		})
		return
	}

	// 處理訊息內容
	if req.AlertManagerData != nil {
		// 使用請求中的模板語言，如果沒有則使用配置檔案中的預設語言
		templateLanguage := req.TemplateLanguage
		if templateLanguage == "" {
			templateLanguage = config.Telegram.TemplateLanguage
		}
		if templateLanguage == "" {
			templateLanguage = "eng"
		}

		// 使用共用 model 產生模板資料，並透過模板引擎渲染
		var formatOptions template.FormatOptions
		if h.templateEngine != nil {
			formatOptions = h.templateEngine.GetCurrentFormatOptions()
		} else {
			formatOptions = h.getFormatOptionsForTelegram()
		}

		data := alertmodel.BuildTemplateData(
			req.AlertManagerData.Status,
			convertAlertSliceToMap(req.AlertManagerData.Alerts),
			convertStringMapToInterface(req.AlertManagerData.GroupLabels),
			convertStringMapToInterface(req.AlertManagerData.CommonLabels),
			convertStringMapToInterface(req.AlertManagerData.CommonAnnotations),
			req.AlertManagerData.ExternalURL,
			formatOptions,
		)

		// 透過模板引擎渲染（含語言回退）
		actualLanguage := templateLanguage
		if h.templateEngine != nil {
			actualLanguage = h.templateEngine.GetDefaultLanguage(templateLanguage)

			logger.Debug("About to render template", "telegram_handler",
				logger.String("language", actualLanguage),
				logger.String("platform", "telegram"))

			msg, rerr := h.templateEngine.RenderTemplateForPlatform(actualLanguage, "telegram", data)

			// 立即記錄渲染結果
			logger.Debug("Template render returned", "telegram_handler",
				logger.Bool("success", rerr == nil),
				logger.String("error", func() string {
					if rerr != nil {
						return rerr.Error()
					}
					return "nil"
				}()),
				logger.Int("message_length", len(msg)))

			if rerr == nil {
				logger.Debug("Template rendered successfully, attempting to send", "telegram_handler",
					logger.Int("level", level),
					logger.String("language", actualLanguage),
					logger.Int("message_length", len(msg)))

				if sendErr := h.telegramService.SendMessage(level, msg); sendErr != nil {
					logger.Error("Failed to send message via template engine path", "telegram_handler",
						logger.Int("level", level),
						logger.String("error_detail", sendErr.Error()),
						logger.Int("message_length", len(msg)))
					c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Failed to send message: " + sendErr.Error()})
					return
				}

				logger.Info("Message sent successfully via template engine", "telegram_handler",
					logger.Int("level", level))
				c.JSON(http.StatusOK, SendMessageResponse{Success: true, Message: "Successfully sent message to Telegram", Level: level})
				return
			}

			logger.Warn("Template rendering failed, falling back to separate messages", "telegram_handler",
				logger.String("render_error", rerr.Error()))
		} else {
			logger.Error("Template engine is nil", "telegram_handler")
		}

		// 若模板引擎不可用或渲染失敗，使用既有的分離發送備援
		err = h.sendSeparateAlertMessages(req.AlertManagerData, templateLanguage, level)
		if err != nil {
			logger.Error("Failed to send alert messages", "telegram_handler",
				logger.Int("level", level),
				logger.Err(err))
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": "Failed to send alert messages: " + err.Error(),
			})
			return
		}
	} else {
		// 發送普通訊息
		err = h.telegramService.SendMessage(level, req.Message)
		if err != nil {
			logger.Error("Failed to send Telegram message", "telegram_handler",
				logger.Int("level", level),
				logger.String("message", req.Message),
				logger.Err(err))
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": "Failed to send message: " + err.Error(),
			})
			return
		}
	}

	// 返回成功回應
	c.JSON(http.StatusOK, SendMessageResponse{
		Success: true,
		Message: "Successfully sent message to Telegram",
		Level:   level,
	})
}
// convertAlertSliceToMap 將 Alert 結構切片轉為通用 map 切片
func convertAlertSliceToMap(alerts []Alert) []map[string]interface{} {
    res := make([]map[string]interface{}, 0, len(alerts))
    for _, a := range alerts {
        m := map[string]interface{}{
            "status":       a.Status,
            "labels":       convertStringMapToInterface(a.Labels),
            "annotations":  convertStringMapToInterface(a.Annotations),
            "startsAt":     a.StartsAt,
            "endsAt":       a.EndsAt,
            "generatorURL": a.GeneratorURL,
            "fingerprint":  a.Fingerprint,
        }
        res = append(res, m)
    }
    return res
}

// convertStringMapToInterface 將 map[string]string 轉為 map[string]interface{}
func convertStringMapToInterface(m map[string]string) map[string]interface{} {
    if m == nil {
        return nil
    }
    res := make(map[string]interface{}, len(m))
    for k, v := range m {
        res[k] = v
    }
    return res
}


// GetBotInfo 獲取機器人資訊
// @Summary 獲取機器人資訊
// @Description 獲取 Telegram 機器人的基本資訊
// @Tags telegram
// @Produce json
// @Security BasicAuth
// @Success 200 {object} map[string]interface{}
// @Router /telegram/info [get]
func (h *Handler) GetBotInfo(c *gin.Context) {
	// 獲取機器人資訊
	meInterface, err := h.telegramService.GetBotInfo()
	if err != nil {
		logger.Error("Failed to get bot info", "telegram_handler", logger.Err(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to get bot info",
		})
		return
	}

	// 類型斷言到具體的 User 類型
	me, ok := meInterface.(*models.User)
	if !ok {
		logger.Error("Failed to cast bot info to User type", "telegram_handler")
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Invalid bot info type",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"bot_info": gin.H{
			"id":        me.ID,
			"username":  me.Username,
			"first_name": me.FirstName,
			"can_join_groups": me.CanJoinGroups,
			"can_read_all_group_messages": me.CanReadAllGroupMessages,
		},
	})
}

// generateAlertManagerMessage 生成 AlertManager 模板訊息
func (h *Handler) generateAlertManagerMessage(webhook *AlertManagerWebhook, language string) string {
	// 統計警報
	firingCount := 0
	resolvedCount := 0
	for _, alert := range webhook.Alerts {
		if alert.Status == "firing" {
			firingCount++
		} else if alert.Status == "resolved" {
			resolvedCount++
		}
	}

	// 獲取第一個警報的基本信息
	var alertName, env, severity, namespace string
	if len(webhook.Alerts) > 0 {
		firstAlert := webhook.Alerts[0]
		alertName = firstAlert.Labels["alertname"]
		env = firstAlert.Labels["env"]
		severity = firstAlert.Labels["severity"]
		namespace = firstAlert.Labels["namespace"]
	}

	// 若第一筆 alert 的 labels 沒有 namespace，回退到 CommonLabels 取值
	if namespace == "" && webhook.CommonLabels != nil {
		if ns, ok := webhook.CommonLabels["namespace"]; ok && ns != "" {
			namespace = ns
		}
	}

	// 轉換警報數據為模板格式
	var alertData []template.AlertData
	for _, alert := range webhook.Alerts {
		alertData = append(alertData, template.AlertData{
			Status:       alert.Status,
			Labels:       alert.Labels,
			Annotations:  alert.Annotations,
			StartsAt:     alert.StartsAt,
			EndsAt:       alert.EndsAt,
			GeneratorURL: alert.GeneratorURL,
		})
	}

	// 準備模板數據
	templateData := template.TemplateData{
		Status:        webhook.Status,
		AlertName:     alertName,
		Env:           env,
		Severity:      severity,
		Namespace:     namespace,
		TotalAlerts:   len(webhook.Alerts),
		FiringCount:   firingCount,
		ResolvedCount: resolvedCount,
		Alerts:        alertData,
		ExternalURL:   webhook.ExternalURL,
	}

	// 使用模板引擎目前的 FormatOptions，確保與配置檔一致
	if h.templateEngine != nil {
		templateData.FormatOptions = h.templateEngine.GetCurrentFormatOptions()
	}

	// 嘗試使用模板引擎渲染
	logger.Debug("Attempting to use template engine", "telegram_handler",
		logger.Bool("has_template_engine", h.templateEngine != nil),
		logger.String("language", language))

	if h.templateEngine != nil {
		// 獲取合適的語言（包含回退邏輯）
		actualLanguage := h.templateEngine.GetDefaultLanguage(language)
		if actualLanguage != language {
			logger.Info("Language fallback applied", "telegram_handler",
				logger.String("requested", language),
				logger.String("actual", actualLanguage))
		}

		logger.Debug("Calling template engine with data", "telegram_handler",
			logger.String("actualLanguage", actualLanguage),
			logger.String("platform", "telegram"),
			logger.Bool("formatOptions.ShowGeneratorURL", templateData.FormatOptions.ShowGeneratorURL.Enabled),
			logger.Bool("formatOptions.ShowExternalURL", templateData.FormatOptions.ShowExternalURL.Enabled))

		message, err := h.templateEngine.RenderTemplateForPlatform(actualLanguage, "telegram", templateData)
		if err == nil {
			messagePreview := message
			if len(message) > 100 {
				messagePreview = message[:100] + "..."
			}
			logger.Info("Template rendered successfully", "telegram_handler",
				logger.String("language", actualLanguage),
				logger.String("available_languages", fmt.Sprintf("%v", h.templateEngine.GetAvailableLanguages())),
				logger.String("message_preview", messagePreview))
			return message
		}
		logger.Warn("Failed to render template, using built-in template", "telegram_handler",
			logger.String("language", actualLanguage),
			logger.Err(err))
	} else {
		logger.Warn("Template engine is nil, using built-in template", "telegram_handler")
	}

	// 如果模板引擎失敗，使用內建的模板邏輯
	return h.generateBuiltInMessage(webhook, language, firingCount, resolvedCount, alertName, env, severity, namespace)
}

// sendSeparateAlertMessages 分別發送觸發中和已解決的警報
func (h *Handler) sendSeparateAlertMessages(webhook *AlertManagerWebhook, language string, level int) error {
	// 分離觸發中和已解決的警報
	var firingAlerts []Alert
	var resolvedAlerts []Alert

	for _, alert := range webhook.Alerts {
		if alert.Status == "firing" {
			firingAlerts = append(firingAlerts, alert)
		} else if alert.Status == "resolved" {
			resolvedAlerts = append(resolvedAlerts, alert)
		}
	}

	// 發送觸發中的警報
	if len(firingAlerts) > 0 {
		firingWebhook := &AlertManagerWebhook{
			Receiver:          webhook.Receiver,
			Status:            "firing",
			Alerts:            firingAlerts,
			GroupLabels:       webhook.GroupLabels,
			CommonLabels:      webhook.CommonLabels,
			CommonAnnotations: webhook.CommonAnnotations,
			ExternalURL:       webhook.ExternalURL,
			Version:           webhook.Version,
			GroupKey:          webhook.GroupKey,
			TruncatedAlerts:   webhook.TruncatedAlerts,
		}

		logger.Debug("Generating firing alerts message", "telegram_handler",
			logger.Int("level", level),
			logger.Int("firing_count", len(firingAlerts)),
			logger.String("language", language))

		firingMessage := h.generateAlertManagerMessage(firingWebhook, language)

		logger.Debug("Attempting to send firing alerts", "telegram_handler",
			logger.Int("level", level),
			logger.Int("message_length", len(firingMessage)))

		if err := h.telegramService.SendMessage(level, firingMessage); err != nil {
			logger.Error("Failed to send firing alerts message", "telegram_handler",
				logger.Int("level", level),
				logger.String("error_detail", err.Error()))
			return fmt.Errorf("failed to send firing alerts: %v", err)
		}

		logger.Info("Firing alerts sent successfully", "telegram_handler",
			logger.Int("level", level),
			logger.Int("alerts_count", len(firingAlerts)))
	}

	// 發送已解決的警報
	if len(resolvedAlerts) > 0 {
		resolvedWebhook := &AlertManagerWebhook{
			Receiver:          webhook.Receiver,
			Status:            "resolved",
			Alerts:            resolvedAlerts,
			GroupLabels:       webhook.GroupLabels,
			CommonLabels:      webhook.CommonLabels,
			CommonAnnotations: webhook.CommonAnnotations,
			ExternalURL:       webhook.ExternalURL,
			Version:           webhook.Version,
			GroupKey:          webhook.GroupKey,
			TruncatedAlerts:   webhook.TruncatedAlerts,
		}

		logger.Debug("Generating resolved alerts message", "telegram_handler",
			logger.Int("level", level),
			logger.Int("resolved_count", len(resolvedAlerts)),
			logger.String("language", language))

		resolvedMessage := h.generateAlertManagerMessage(resolvedWebhook, language)

		logger.Debug("Attempting to send resolved alerts", "telegram_handler",
			logger.Int("level", level),
			logger.Int("message_length", len(resolvedMessage)))

		if err := h.telegramService.SendMessage(level, resolvedMessage); err != nil {
			logger.Error("Failed to send resolved alerts message", "telegram_handler",
				logger.Int("level", level),
				logger.String("error_detail", err.Error()))
			return fmt.Errorf("failed to send resolved alerts: %v", err)
		}

		logger.Info("Resolved alerts sent successfully", "telegram_handler",
			logger.Int("level", level),
			logger.Int("alerts_count", len(resolvedAlerts)))
	}

	return nil
}

// generateBuiltInMessage 生成內建的訊息模板（備用方案）
func (h *Handler) generateBuiltInMessage(webhook *AlertManagerWebhook, language string, firingCount, resolvedCount int, alertName, env, severity, namespace string) string {
	var message strings.Builder

	// 為 Telegram MarkdownV2 創建轉義函數
	escapeText := func(text string) string {
		// 轉義 MarkdownV2 特殊字符
		specialChars := []string{
			"_", "*", "[", "]", "(", ")", "~", "`", ">", "#", "+", "-", "=", "|", "{", "}", ".", "!",
		}

		result := text
		for _, char := range specialChars {
			result = strings.ReplaceAll(result, char, "\\"+char)
		}
		return result
	}

	// 讀取當前平台顯示開關（優先使用模板引擎載入的配置）
	var formatOptions template.FormatOptions
	if h.templateEngine != nil {
		formatOptions = h.templateEngine.GetCurrentFormatOptions()
	} else {
		formatOptions = h.getFormatOptionsForTelegram()
	}

	if language == "tw" {
		// 繁體中文模板
		if firingCount > 0 {
			message.WriteString("🚨 *警報通知*\n\n")
		} else if resolvedCount > 0 {
			message.WriteString("✅ *警報已解決*\n\n")
		}

		message.WriteString(fmt.Sprintf("*狀態:* %s\n", escapeText(webhook.Status)))
		message.WriteString(fmt.Sprintf("*警報名稱:* %s\n", escapeText(alertName)))
		message.WriteString(fmt.Sprintf("*環境:* %s\n", escapeText(env)))
		message.WriteString(fmt.Sprintf("*嚴重程度:* %s\n", escapeText(severity)))
		message.WriteString(fmt.Sprintf("*命名空間:* %s\n", escapeText(namespace)))
		message.WriteString(fmt.Sprintf("*總警報數:* %d\n", len(webhook.Alerts)))

		if firingCount > 0 {
			message.WriteString(fmt.Sprintf("*觸發中:* %d\n", firingCount))
		}
		if resolvedCount > 0 {
			message.WriteString(fmt.Sprintf("*已解決:* %d\n", resolvedCount))
		}

		// 詳細警報列表
		if firingCount > 0 {
			message.WriteString("\n*🚨 觸發中的警報:*\n")
			for i, alert := range webhook.Alerts {
				if alert.Status == "firing" {
					message.WriteString(fmt.Sprintf("\n*警報 %d:*\n", i+1))
					message.WriteString(fmt.Sprintf("• 摘要: %s\n", escapeText(alert.Annotations["summary"])))
					// 若提供 description，則補充描述
					if desc, ok := alert.Annotations["description"]; ok && desc != "" {
						message.WriteString(fmt.Sprintf("• 描述: %s\n", escapeText(desc)))
					}
					if pod := alert.Labels["pod"]; pod != "" {
						message.WriteString(fmt.Sprintf("• Pod: %s\n", escapeText(pod)))
					}
					message.WriteString(fmt.Sprintf("• 開始時間: %s\n", escapeText(h.formatTime(alert.StartsAt))))
					if alert.EndsAt != "0001-01-01T00:00:00Z" {
						message.WriteString(fmt.Sprintf("• 結束時間: %s\n", escapeText(h.formatTime(alert.EndsAt))))
					}
					if formatOptions.ShowGeneratorURL.Enabled && alert.GeneratorURL != "" {
						message.WriteString(fmt.Sprintf("• [查看詳情](%s)\n", alert.GeneratorURL))
					}
				}
			}
		}

		if resolvedCount > 0 {
			message.WriteString("\n*✅ 已解決的警報:*\n")
			for i, alert := range webhook.Alerts {
				if alert.Status == "resolved" {
					message.WriteString(fmt.Sprintf("\n*警報 %d:*\n", i+1))
					message.WriteString(fmt.Sprintf("• 摘要: %s\n", escapeText(alert.Annotations["summary"])))
					// 若提供 description，則補充描述
					if desc, ok := alert.Annotations["description"]; ok && desc != "" {
						message.WriteString(fmt.Sprintf("• 描述: %s\n", escapeText(desc)))
					}
					if pod := alert.Labels["pod"]; pod != "" {
						message.WriteString(fmt.Sprintf("• Pod: %s\n", escapeText(pod)))
					}
					message.WriteString(fmt.Sprintf("• 開始時間: %s\n", escapeText(h.formatTime(alert.StartsAt))))
					message.WriteString(fmt.Sprintf("• 結束時間: %s\n", escapeText(h.formatTime(alert.EndsAt))))
					if formatOptions.ShowGeneratorURL.Enabled && alert.GeneratorURL != "" {
						message.WriteString(fmt.Sprintf("• [查看詳情](%s)\n", alert.GeneratorURL))
					}
				}
			}
		}

		if formatOptions.ShowExternalURL.Enabled && webhook.ExternalURL != "" {
			message.WriteString(fmt.Sprintf("\n[查看所有警報詳情](%s)", webhook.ExternalURL))
		}
	} else {
		// 英文模板
		if firingCount > 0 {
			message.WriteString("🚨 *Alert Notification*\n\n")
		} else if resolvedCount > 0 {
			message.WriteString("✅ *Alert Resolved*\n\n")
		}

		message.WriteString(fmt.Sprintf("*Status:* %s\n", escapeText(webhook.Status)))
		message.WriteString(fmt.Sprintf("*Alert Name:* %s\n", escapeText(alertName)))
		message.WriteString(fmt.Sprintf("*Environment:* %s\n", escapeText(env)))
		message.WriteString(fmt.Sprintf("*Severity:* %s\n", escapeText(severity)))
		message.WriteString(fmt.Sprintf("*Namespace:* %s\n", escapeText(namespace)))
		message.WriteString(fmt.Sprintf("*Total Alerts:* %d\n", len(webhook.Alerts)))

		if firingCount > 0 {
			message.WriteString(fmt.Sprintf("*Firing:* %d\n", firingCount))
		}
		if resolvedCount > 0 {
			message.WriteString(fmt.Sprintf("*Resolved:* %d\n", resolvedCount))
		}

		// 詳細警報列表
		if firingCount > 0 {
			message.WriteString("\n*🚨 Firing Alerts:*\n")
			for i, alert := range webhook.Alerts {
				if alert.Status == "firing" {
					message.WriteString(fmt.Sprintf("\n*Alert %d:*\n", i+1))
					message.WriteString(fmt.Sprintf("• Summary: %s\n", escapeText(alert.Annotations["summary"])))
					// If description provided, add it
					if desc, ok := alert.Annotations["description"]; ok && desc != "" {
						message.WriteString(fmt.Sprintf("• Description: %s\n", escapeText(desc)))
					}
					if pod := alert.Labels["pod"]; pod != "" {
						message.WriteString(fmt.Sprintf("• Pod: %s\n", escapeText(pod)))
					}
					message.WriteString(fmt.Sprintf("• Started: %s\n", escapeText(h.formatTime(alert.StartsAt))))
					if alert.EndsAt != "0001-01-01T00:00:00Z" {
						message.WriteString(fmt.Sprintf("• Ended: %s\n", escapeText(h.formatTime(alert.EndsAt))))
					}
					if formatOptions.ShowGeneratorURL.Enabled && alert.GeneratorURL != "" {
						message.WriteString(fmt.Sprintf("• [View Details](%s)\n", alert.GeneratorURL))
					}
				}
			}
		}

		if resolvedCount > 0 {
			message.WriteString("\n*✅ Resolved Alerts:*\n")
			for i, alert := range webhook.Alerts {
				if alert.Status == "resolved" {
					message.WriteString(fmt.Sprintf("\n*Alert %d:*\n", i+1))
					message.WriteString(fmt.Sprintf("• Summary: %s\n", escapeText(alert.Annotations["summary"])))
					// If description provided, add it
					if desc, ok := alert.Annotations["description"]; ok && desc != "" {
						message.WriteString(fmt.Sprintf("• Description: %s\n", escapeText(desc)))
					}
					if pod := alert.Labels["pod"]; pod != "" {
						message.WriteString(fmt.Sprintf("• Pod: %s\n", escapeText(pod)))
					}
					message.WriteString(fmt.Sprintf("• Started: %s\n", escapeText(h.formatTime(alert.StartsAt))))
					message.WriteString(fmt.Sprintf("• Ended: %s\n", escapeText(h.formatTime(alert.EndsAt))))
					if formatOptions.ShowGeneratorURL.Enabled && alert.GeneratorURL != "" {
						message.WriteString(fmt.Sprintf("• [View Details](%s)\n", alert.GeneratorURL))
					}
				}
			}
		}

		if formatOptions.ShowExternalURL.Enabled && webhook.ExternalURL != "" {
			message.WriteString(fmt.Sprintf("\n[View All Alert Details](%s)", webhook.ExternalURL))
		}
	}

	return message.String()
}

// formatTime 格式化時間字符串
func (h *Handler) formatTime(timeStr string) string {
	if timeStr == "" || timeStr == "0001-01-01T00:00:00Z" {
		return "未設定"
	}

	// 嘗試解析 ISO 8601 格式的時間
	t, err := time.Parse(time.RFC3339, timeStr)
	if err != nil {
		// 如果解析失敗，返回原始字符串
		return timeStr
	}

	// 格式化為本地時間
	return t.Format("2006-01-02 15:04:05")
}

// getFormatOptionsForTelegram 根據 Telegram 配置返回對應的 FormatOptions
func (h *Handler) getFormatOptionsForTelegram() template.FormatOptions {
	templateMode := config.Conf.Telegram.TemplateMode
	if templateMode == "" {
		templateMode = "full" // Default to full mode
	}

	if templateMode == "minimal" {
		// 從 template engine 載入 minimal 配置，而不是硬編碼
		if h.templateEngine != nil {
			minimalConfig := h.templateEngine.GetMinimalDefaultConfig()
			if minimalConfig != nil {
				logger.Debug("Using minimal config FormatOptions for Telegram", "TelegramHandler",
					logger.Bool("ShowEmoji", minimalConfig.FormatOptions.ShowEmoji.Enabled),
					logger.Bool("ShowTimestamps", minimalConfig.FormatOptions.ShowTimestamps.Enabled),
					logger.Bool("ShowGeneratorURL", minimalConfig.FormatOptions.ShowGeneratorURL.Enabled),
					logger.Bool("ShowExternalURL", minimalConfig.FormatOptions.ShowExternalURL.Enabled))
				return minimalConfig.FormatOptions
			}
		}

		// 回退到硬編碼配置（如果模板引擎不可用）
		logger.Debug("Fallback to hardcoded minimal FormatOptions for Telegram", "TelegramHandler")
		return template.FormatOptions{
			ShowLinks: struct {
				Enabled     bool   `yaml:"enabled"`
				Description string `yaml:"description"`
			}{Enabled: false, Description: "是否顯示超連結"},
			ShowTimestamps: struct {
				Enabled     bool   `yaml:"enabled"`
				Description string `yaml:"description"`
			}{Enabled: true, Description: "是否顯示時間戳"}, // 與 minimal config 一致
			ShowExternalURL: struct {
				Enabled     bool   `yaml:"enabled"`
				Description string `yaml:"description"`
			}{Enabled: false, Description: "是否顯示外部連結"},
			ShowGeneratorURL: struct {
				Enabled     bool   `yaml:"enabled"`
				Description string `yaml:"description"`
			}{Enabled: false, Description: "是否顯示生成器連結"},
			ShowEmoji: struct {
				Enabled     bool   `yaml:"enabled"`
				Description string `yaml:"description"`
			}{Enabled: true, Description: "是否顯示表情符號"}, // 與 minimal config 一致
			CompactMode: struct {
				Enabled     bool   `yaml:"enabled"`
				Description string `yaml:"description"`
			}{Enabled: true, Description: "緊湊模式（簡化顯示）"},
		}
	} else {
		// Full mode: enable all options
		return template.FormatOptions{
			ShowLinks: struct {
				Enabled     bool   `yaml:"enabled"`
				Description string `yaml:"description"`
			}{Enabled: true, Description: "是否顯示超連結"},
			ShowTimestamps: struct {
				Enabled     bool   `yaml:"enabled"`
				Description string `yaml:"description"`
			}{Enabled: true, Description: "是否顯示時間戳"},
			ShowExternalURL: struct {
				Enabled     bool   `yaml:"enabled"`
				Description string `yaml:"description"`
			}{Enabled: true, Description: "是否顯示外部連結"},
			ShowGeneratorURL: struct {
				Enabled     bool   `yaml:"enabled"`
				Description string `yaml:"description"`
			}{Enabled: true, Description: "是否顯示生成器連結"},
			ShowEmoji: struct {
				Enabled     bool   `yaml:"enabled"`
				Description string `yaml:"description"`
			}{Enabled: true, Description: "是否顯示表情符號"},
			CompactMode: struct {
				Enabled     bool   `yaml:"enabled"`
				Description string `yaml:"description"`
			}{Enabled: false, Description: "緊湊模式（簡化顯示）"},
		}
	}
}

// getBodyPreview 取得請求體的預覽內容，限制長度避免日誌過長
func getBodyPreview(body []byte) string {
	const maxPreviewLength = 200
	if len(body) <= maxPreviewLength {
		return string(body)
	}
	return string(body[:maxPreviewLength]) + "..."
}
