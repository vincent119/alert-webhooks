package slack

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"alert-webhooks/config"
	"alert-webhooks/pkg/alertmodel"
	"alert-webhooks/pkg/logger"
	"alert-webhooks/pkg/service"
	"alert-webhooks/pkg/template"

	"github.com/gin-gonic/gin"
)

// Handler Slack 路由處理器
type Handler struct {
	slackService   *service.SlackService
	templateEngine *template.TemplateEngine
}

// NewHandler 創建新的 Slack 路由處理器
func NewHandler(slackService *service.SlackService) *Handler {
	logger.Info("Creating new Slack handler", "slack_handler")

	// 從 ServiceManager 獲取模板引擎
	serviceManager := service.GetServiceManager()
	templateEngine := serviceManager.GetTemplateEngine()

	if templateEngine == nil {
		logger.Error("Template engine not available from service manager", "slack_handler")
		return &Handler{
			slackService:   slackService,
			templateEngine: template.NewTemplateEngine(), // 後備方案
		}
	}

	logger.Info("Template engine obtained from service manager", "slack_handler")

	return &Handler{
		slackService:   slackService,
		templateEngine: templateEngine,
	}
}

// SendMessageRequest 發送訊息請求結構
type SendMessageRequest struct {
	Message          string                 `json:"message,omitempty"`           // 簡單文字訊息
	AlertManagerData map[string]interface{} `json:"alertmanager_data,omitempty"` // AlertManager webhook 數據（包裝格式）
	Username         string                 `json:"username,omitempty"`          // 自定義 Bot 名稱
	IconURL          string                 `json:"icon_url,omitempty"`          // 自定義 Bot 頭像 URL
	IconEmoji        string                 `json:"icon_emoji,omitempty"`        // 自定義 Bot 表情符號
	ThreadTS         string                 `json:"thread_ts,omitempty"`         // 線程時間戳（回覆特定訊息）

	// 原始 AlertManager JSON 格式（直接接受）
	Receiver          string                   `json:"receiver,omitempty"`
	Status            string                   `json:"status,omitempty"`
	Alerts            []map[string]interface{} `json:"alerts,omitempty"`
	GroupLabels       map[string]interface{}   `json:"groupLabels,omitempty"`
	CommonLabels      map[string]interface{}   `json:"commonLabels,omitempty"`
	CommonAnnotations map[string]interface{}   `json:"commonAnnotations,omitempty"`
	ExternalURL       string                   `json:"externalURL,omitempty"`
	Version           string                   `json:"version,omitempty"`
	GroupKey          string                   `json:"groupKey,omitempty"`
	TruncatedAlerts   int                      `json:"truncatedAlerts,omitempty"`
}

// RichMessageRequest 富文本訊息請求結構
type RichMessageRequest struct {
	Title       string                  `json:"title"`                 // 訊息標題
	Message     string                  `json:"message"`               // 訊息內容
	Color       string                  `json:"color,omitempty"`       // 顏色 (good, warning, danger, 或 hex)
	Fields      []service.Field         `json:"fields,omitempty"`      // 字段列表
	Username    string                  `json:"username,omitempty"`    // 自定義 Bot 名稱
	IconURL     string                  `json:"icon_url,omitempty"`    // 自定義 Bot 頭像 URL
	IconEmoji   string                  `json:"icon_emoji,omitempty"`  // 自定義 Bot 表情符號
	FooterText  string                  `json:"footer_text,omitempty"` // 頁腳文字
	FooterIcon  string                  `json:"footer_icon,omitempty"` // 頁腳圖示
}

// SendMessageResponse 發送訊息響應結構
type SendMessageResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Level   string `json:"level"`
}

// StatusResponse Slack 服務狀態響應
type StatusResponse struct {
	Success   bool              `json:"success"`
	Enabled   bool              `json:"enabled"`
	Connected bool              `json:"connected"`
	Channels  map[string]string `json:"channels,omitempty"`
}

// ChannelsResponse 頻道列表響應
type ChannelsResponse struct {
	Success  bool              `json:"success"`
	Channels map[string]string `json:"channels"`
}

// SendMessageToChannel 發送訊息到指定頻道
// @Summary 發送 Slack 訊息到指定頻道
// @Description 發送訊息到指定的 Slack 頻道
// @Tags slack
// @Accept json
// @Produce json
// @Security BasicAuth
// @Param channel path string true "頻道名稱 (例如: alerts, emergency)"
// @Param request body SendMessageRequest true "發送訊息請求"
// @Success 200 {object} SendMessageResponse
// @Failure 400 {object} SendMessageResponse
// @Failure 401 {object} SendMessageResponse
// @Failure 500 {object} SendMessageResponse
// @Router /slack/channel/{channel} [post]
func (h *Handler) SendMessageToChannel(c *gin.Context) {
	// 檢查 Slack 服務是否可用
	if h.slackService == nil {
		c.JSON(http.StatusServiceUnavailable, SendMessageResponse{
			Success: false,
			Message: "Slack service not available",
		})
		return
	}

	channel := c.Param("channel")
	if channel == "" {
		c.JSON(http.StatusBadRequest, SendMessageResponse{
			Success: false,
			Message: "Channel parameter is required",
		})
		return
	}

	// 確保頻道名稱以 # 開頭
	if channel[0] != '#' && channel[0] != '@' {
		channel = "#" + channel
	}

	var req SendMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, SendMessageResponse{
			Success: false,
			Message: "Invalid request format: " + err.Error(),
		})
		return
	}

	// 檢查是否提供了訊息內容（支援三種格式）
	isRawAlertManager := req.Alerts != nil || req.Status != ""
	if req.Message == "" && req.AlertManagerData == nil && !isRawAlertManager {
		c.JSON(http.StatusBadRequest, SendMessageResponse{
			Success: false,
			Message: "Either message, alertmanager_data, or raw AlertManager JSON must be provided",
		})
		return
	}

	var message string
	if req.Message != "" {
		message = req.Message
	} else if isRawAlertManager {
		// 處理原始 AlertManager JSON 格式
		message = h.formatAlertManagerMessage(&req)
	} else {
		// 處理包裝格式的 AlertManager 數據
		message = "AlertManager notification (wrapped format - template integration pending)"
	}

	// 構建 Slack 訊息選項
	options := &service.SlackMessage{
		Username:  req.Username,
		IconURL:   req.IconURL,
		IconEmoji: req.IconEmoji,
		ThreadTS:  req.ThreadTS,
	}

	// 發送訊息
	if err := h.slackService.SendMessageWithOptions(channel, message, options); err != nil {
		logger.Error("Failed to send Slack message", "slack_handler",
			logger.String("channel", channel),
			logger.String("message", message),
			logger.Err(err))

		c.JSON(http.StatusInternalServerError, SendMessageResponse{
			Success: false,
			Message: "Failed to send message: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, SendMessageResponse{
		Success: true,
		Message: "Slack Message sent successfully",
	})
}

// SendMessageToLevel 發送訊息到指定等級
// @Summary 發送 Slack 訊息到指定等級
// @Description 發送訊息到指定等級對應的 Slack 頻道
// @Tags slack
// @Accept json
// @Produce json
// @Security BasicAuth
// @Param level path string true "等級名稱 (例如: emergency, critical, warning, info)"
// @Param request body SendMessageRequest true "發送訊息請求"
// @Success 200 {object} SendMessageResponse
// @Failure 400 {object} SendMessageResponse
// @Failure 401 {object} SendMessageResponse
// @Failure 500 {object} SendMessageResponse
// @Router /api/v1/slack/level/{level} [post]
func (h *Handler) SendMessageToLevel(c *gin.Context) {
	// 檢查 Slack 服務是否可用
	if h.slackService == nil {
		c.JSON(http.StatusServiceUnavailable, SendMessageResponse{
			Success: false,
			Message: "Slack service not available",
		})
		return
	}

	level := c.Param("level")
	if level == "" {
		c.JSON(http.StatusBadRequest, SendMessageResponse{
			Success: false,
			Message: "Level parameter is required",
		})
		return
	}

	// 將數字等級轉換為 "L{數字}" 格式以匹配配置
	if len(level) > 0 && level[0] >= '0' && level[0] <= '9' {
		level = "L" + level
	}

	var req SendMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, SendMessageResponse{
			Success: false,
			Message: "Invalid request format: " + err.Error(),
		})
		return
	}

	// 檢查是否提供了訊息內容（支援三種格式）
	isRawAlertManager := req.Alerts != nil || req.Status != ""
	if req.Message == "" && req.AlertManagerData == nil && !isRawAlertManager {
		c.JSON(http.StatusBadRequest, SendMessageResponse{
			Success: false,
			Message: "Either message, alertmanager_data, or raw AlertManager JSON must be provided",
		})
		return
	}

	var message string
	if req.Message != "" {
		message = req.Message
	} else if isRawAlertManager {
		// 使用共用模型 + 模板引擎渲染 Slack 訊息
		serviceManager := service.GetServiceManager()
		te := serviceManager.GetTemplateEngine()
		formatOptions := template.FormatOptions{}
		if te != nil {
			formatOptions = te.GetCurrentFormatOptions()
		}
		data := alertmodel.BuildTemplateData(
			req.Status,
			req.Alerts,
			req.GroupLabels,
			req.CommonLabels,
			req.CommonAnnotations,
			req.ExternalURL,
			formatOptions,
		)
		lang := config.Slack.TemplateLanguage
		if lang == "" { lang = "eng" }
		if te != nil {
			actual := te.GetDefaultLanguage(lang)
			if msg, err := te.RenderTemplateForPlatform(actual, "slack", data); err == nil {
				message = msg
			} else {
				message = h.generateBuiltInSlackMessage(&req, data.FiringCount, data.ResolvedCount, data.AlertName, data.Env, data.Severity, data.Namespace)
			}
		} else {
			message = h.generateBuiltInSlackMessage(&req, data.FiringCount, data.ResolvedCount, data.AlertName, data.Env, data.Severity, data.Namespace)
		}
	} else {
		// 處理包裝格式的 AlertManager 數據
		message = "AlertManager notification (wrapped format - template integration pending)"
	}

	// 發送訊息到指定等級
	if err := h.slackService.SendMessageToLevel(c.Request.Context(), level, message); err != nil {
		logger.Error("Failed to send Slack message to level", "slack_handler",
			logger.String("level", level),
			logger.String("message", message),
			logger.Err(err))

		c.JSON(http.StatusInternalServerError, SendMessageResponse{
			Success: false,
			Message: "Failed to send message: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, SendMessageResponse{
		Success: true,
		Message: "Slack Message sent successfully to level ",
		Level:   level,
	})
}

// SendRichMessage 發送富文本訊息
// @Summary 發送富文本 Slack 訊息
// @Description 發送包含附件和字段的富文本訊息到指定 Slack 頻道
// @Tags slack
// @Accept json
// @Produce json
// @Security BasicAuth
// @Param channel path string true "頻道名稱 (例如: alerts, emergency)"
// @Param request body RichMessageRequest true "富文本訊息請求"
// @Success 200 {object} SendMessageResponse
// @Failure 400 {object} SendMessageResponse
// @Failure 401 {object} SendMessageResponse
// @Failure 500 {object} SendMessageResponse
// @Router /api/v1/slack/rich/{channel} [post]
func (h *Handler) SendRichMessage(c *gin.Context) {
	// 檢查 Slack 服務是否可用
	if h.slackService == nil {
		c.JSON(http.StatusServiceUnavailable, SendMessageResponse{
			Success: false,
			Message: "Slack service not available",
		})
		return
	}

	channel := c.Param("channel")
	if channel == "" {
		c.JSON(http.StatusBadRequest, SendMessageResponse{
			Success: false,
			Message: "Channel parameter is required",
		})
		return
	}

	// 確保頻道名稱以 # 開頭
	if channel[0] != '#' && channel[0] != '@' {
		channel = "#" + channel
	}

	var req RichMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, SendMessageResponse{
			Success: false,
			Message: "Invalid request format: " + err.Error(),
		})
		return
	}

	// 檢查必需字段
	if req.Title == "" {
		c.JSON(http.StatusBadRequest, SendMessageResponse{
			Success: false,
			Message: "Title is required for rich messages",
		})
		return
	}

	// 設定預設顏色
	color := req.Color
	if color == "" {
		color = "good" // 預設為綠色
	}

	// 發送富文本訊息
	if err := h.slackService.SendRichMessage(channel, req.Title, req.Message, color, req.Fields); err != nil {
		logger.Error("Failed to send rich Slack message", "slack_handler",
			logger.String("channel", channel),
			logger.String("title", req.Title),
			logger.Err(err))

		c.JSON(http.StatusInternalServerError, SendMessageResponse{
			Success: false,
			Message: "Failed to send rich message: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, SendMessageResponse{
		Success: true,
		Message: "Rich message sent successfully",
	})
}

// GetStatus 獲取 Slack 服務狀態
// @Summary 獲取 Slack 服務狀態
// @Description 獲取 Slack 服務的當前狀態和配置信息
// @Tags slack
// @Produce json
// @Security BasicAuth
// @Success 200 {object} StatusResponse
// @Failure 401 {object} SendMessageResponse
// @Router /slack/status [get]
func (h *Handler) GetStatus(c *gin.Context) {
	status := StatusResponse{
		Success:   true,
		Enabled:   config.Slack.Enable,
		Connected: h.slackService != nil,
	}

	if h.slackService != nil {
		status.Channels = h.slackService.GetChannels()
	}

	c.JSON(http.StatusOK, status)
}

// GetChannels 獲取頻道列表
// @Summary 獲取 Slack 頻道配置
// @Description 獲取當前配置的所有 Slack 頻道映射
// @Tags slack
// @Produce json
// @Security BasicAuth
// @Success 200 {object} ChannelsResponse
// @Failure 401 {object} SendMessageResponse
// @Failure 503 {object} SendMessageResponse
// @Router /slack/channels [get]
func (h *Handler) GetChannels(c *gin.Context) {
	if h.slackService == nil {
		c.JSON(http.StatusServiceUnavailable, SendMessageResponse{
			Success: false,
			Message: "Slack service not available",
		})
		return
	}

	channels := h.slackService.GetChannels()
	c.JSON(http.StatusOK, ChannelsResponse{
		Success:  true,
		Channels: channels,
	})
}

// TestConnection 測試 Slack 連接
// @Summary 測試 Slack 連接
// @Description 發送測試訊息到預設頻道以驗證 Slack 連接
// @Tags slack
// @Accept json
// @Produce json
// @Security BasicAuth
// @Success 200 {object} SendMessageResponse
// @Failure 401 {object} SendMessageResponse
// @Failure 503 {object} SendMessageResponse
// @Router /slack/test [post]
func (h *Handler) TestConnection(c *gin.Context) {
	if h.slackService == nil {
		c.JSON(http.StatusServiceUnavailable, SendMessageResponse{
			Success: false,
			Message: "Slack service not available",
		})
		return
	}

	// 構建測試訊息
	testMessage := "🧪 Slack 連接測試成功！\n" +
		"時間: " + strconv.FormatInt(int64(1), 10) + "\n" +
		"服務: Alert Webhooks"

	// 嘗試發送到預設頻道
	var testChannel string
	if config.Slack.Channel != "" {
		testChannel = config.Slack.Channel
	} else {
		// 如果沒有預設頻道，使用第一個配置的頻道
		channels := h.slackService.GetChannels()
		for _, channel := range channels {
			testChannel = channel
			break
		}
	}

	if testChannel == "" {
		c.JSON(http.StatusBadRequest, SendMessageResponse{
			Success: false,
			Message: "No channel configured for testing",
		})
		return
	}

	// 發送測試訊息
	if err := h.slackService.SendMessage(c.Request.Context(), testChannel, testMessage); err != nil {
		logger.Error("Slack connection test failed", "slack_handler",
			logger.String("channel", testChannel),
			logger.Err(err))

		c.JSON(http.StatusInternalServerError, SendMessageResponse{
			Success: false,
			Message: "Connection test failed: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, SendMessageResponse{
		Success: true,
		Message: "Connection test successful, message sent to " + testChannel,
	})
}

// ValidateChannel 驗證 Bot 是否已加入頻道
// @Summary 驗證 Slack 頻道
// @Description 檢查 Bot 是否已加入指定頻道，如果未加入會提供加入指引
// @Tags slack
// @Produce json
// @Security BasicAuth
// @Param channel path string true "頻道名稱 (例如: alerts, emergency)"
// @Success 200 {object} SendMessageResponse
// @Failure 400 {object} SendMessageResponse
// @Failure 401 {object} SendMessageResponse
// @Failure 503 {object} SendMessageResponse
// @Router /slack/validate/{channel} [post]
func (h *Handler) ValidateChannel(c *gin.Context) {
	if h.slackService == nil {
		c.JSON(http.StatusServiceUnavailable, SendMessageResponse{
			Success: false,
			Message: "Slack service not available",
		})
		return
	}

	channel := c.Param("channel")
	if channel == "" {
		c.JSON(http.StatusBadRequest, SendMessageResponse{
			Success: false,
			Message: "Channel parameter is required",
		})
		return
	}

	// 確保頻道名稱以 # 開頭
	if channel[0] != '#' && channel[0] != '@' {
		channel = "#" + channel
	}

	// 嘗試發送一個非常簡短的驗證訊息
	testMessage := "✅ Bot 已加入此頻道"

	if err := h.slackService.SendMessage(c.Request.Context(), channel, testMessage); err != nil {
		// 分析錯誤類型並提供具體指引
		if strings.Contains(err.Error(), "not_in_channel") {
			c.JSON(http.StatusBadRequest, SendMessageResponse{
				Success: false,
				Message: fmt.Sprintf("❌ Bot is not in channel %s\n\nSolution:\n1. Open %s channel in Slack\n2. Type: /invite @your_bot_name\n3. Or add Bot in channel settings", channel, channel),
			})
		} else if strings.Contains(err.Error(), "channel_not_found") {
			c.JSON(http.StatusBadRequest, SendMessageResponse{
				Success: false,
				Message: fmt.Sprintf("❌ Channel %s does not exist\n\nPlease check:\n1. Is the channel name correct\n2. Is it a private channel (Bot needs to be invited)", channel),
			})
		} else {
			c.JSON(http.StatusInternalServerError, SendMessageResponse{
				Success: false,
				Message: fmt.Sprintf("❌ Channel validation failed: %s", err.Error()),
			})
		}
		return
	}

	c.JSON(http.StatusOK, SendMessageResponse{
		Success: true,
		Message: fmt.Sprintf("✅ Channel %s validation successful! Bot has joined and can send messages", channel),
	})
}

// formatAlertManagerMessage 格式化原始 AlertManager JSON 為 Slack 訊息
func (h *Handler) formatAlertManagerMessage(req *SendMessageRequest) string {
	// 使用 Slack 配置中的模板語言
	templateLanguage := config.Slack.TemplateLanguage
	if templateLanguage == "" {
		templateLanguage = "eng" // 預設使用英文
	}

	// 統計警報
	firingCount := 0
	resolvedCount := 0
	for _, alert := range req.Alerts {
		if status, ok := alert["status"].(string); ok {
			if status == "firing" {
				firingCount++
			} else if status == "resolved" {
				resolvedCount++
			}
		}
	}

	// 獲取基本信息
	var alertName, env, severity, namespace string
	if commonLabels := req.CommonLabels; commonLabels != nil {
		if alertname, ok := commonLabels["alertname"].(string); ok {
			alertName = alertname
		}
		if envValue, ok := commonLabels["env"].(string); ok {
			env = envValue
		}
		if severityValue, ok := commonLabels["severity"].(string); ok {
			severity = severityValue
		}
		if namespaceValue, ok := commonLabels["namespace"].(string); ok {
			namespace = namespaceValue
		}
	}

	// 如果 commonLabels 中沒有 namespace，嘗試從第一個 alert 的 labels 中獲取
	if namespace == "" && len(req.Alerts) > 0 {
		if firstAlert := req.Alerts[0]; firstAlert != nil {
			if labels, ok := firstAlert["labels"].(map[string]interface{}); ok {
				if namespaceValue, ok := labels["namespace"].(string); ok {
					namespace = namespaceValue
				}
			}
		}
	}

	// 轉換警報數據為模板格式
	var alertData []template.AlertData
	for _, alert := range req.Alerts {
		alertDataItem := template.AlertData{}

		if status, ok := alert["status"].(string); ok {
			alertDataItem.Status = status
		}

		if labels, ok := alert["labels"].(map[string]interface{}); ok {
			alertDataItem.Labels = make(map[string]string)
			for k, v := range labels {
				if str, ok := v.(string); ok {
					alertDataItem.Labels[k] = str
				}
			}
		}

		if annotations, ok := alert["annotations"].(map[string]interface{}); ok {
			alertDataItem.Annotations = make(map[string]string)
			for k, v := range annotations {
				if str, ok := v.(string); ok {
					alertDataItem.Annotations[k] = str
				}
			}
		}

		if startsAt, ok := alert["startsAt"].(string); ok {
			alertDataItem.StartsAt = startsAt
		}

		if endsAt, ok := alert["endsAt"].(string); ok {
			alertDataItem.EndsAt = endsAt
		}

		if generatorURL, ok := alert["generatorURL"].(string); ok {
			alertDataItem.GeneratorURL = generatorURL
		}

		alertData = append(alertData, alertDataItem)
	}

	// 準備模板數據
	templateData := template.TemplateData{
		Status:        req.Status,
		AlertName:     alertName,
		Env:           env,
		Severity:      severity,
		Namespace:     namespace,
		TotalAlerts:   len(req.Alerts),
		FiringCount:   firingCount,
		ResolvedCount: resolvedCount,
		Alerts:        alertData,
		ExternalURL:   req.ExternalURL,
	}

	// 動態獲取最新的模板引擎（支援熱重載）
	serviceManager := service.GetServiceManager()
	currentTemplateEngine := serviceManager.GetTemplateEngine()

	// 嘗試使用模板引擎渲染
	if currentTemplateEngine != nil {
		// 獲取合適的語言（包含回退邏輯）
		actualLanguage := currentTemplateEngine.GetDefaultLanguage(templateLanguage)
		if actualLanguage != templateLanguage {
			logger.Info("Language fallback applied for Slack", "slack_handler",
				logger.String("requested", templateLanguage),
				logger.String("actual", actualLanguage))
		}

		message, err := currentTemplateEngine.RenderTemplateForPlatform(actualLanguage, "slack", templateData)
		if err == nil {
			messagePreview := message
			if len(message) > 100 {
				messagePreview = message[:100] + "..."
			}
			logger.Info("Slack template rendered successfully", "slack_handler",
				logger.String("language", actualLanguage),
				logger.String("available_languages", fmt.Sprintf("%v", currentTemplateEngine.GetAvailableLanguages())),
				logger.String("message_preview", messagePreview))
			return message
		}
		logger.Warn("Failed to render Slack template, using built-in template", "slack_handler",
			logger.String("language", actualLanguage),
			logger.Err(err))
	} else {
		logger.Warn("Template engine is nil, using built-in template for Slack", "slack_handler")
	}

	// 如果模板引擎失敗，使用內建的模板邏輯
	return h.generateBuiltInSlackMessage(req, firingCount, resolvedCount, alertName, env, severity, namespace)
}

// generateBuiltInSlackMessage 生成內建的 Slack 訊息格式（後備方案）
func (h *Handler) generateBuiltInSlackMessage(req *SendMessageRequest, firingCount, resolvedCount int, alertName, env, severity, namespace string) string {
	var message strings.Builder

	// 警報標題和基本信息
	if req.Status == "firing" && firingCount > 0 {
		message.WriteString("🚨 *警報通知*\n")
	} else if req.Status == "resolved" || resolvedCount > 0 {
		message.WriteString("✅ *警報解決*\n")
	} else {
		message.WriteString("📊 *警報狀態更新*\n")
	}

	if alertName != "" {
		message.WriteString(fmt.Sprintf("*Alert Name:* %s\n", alertName))
	}
	if env != "" {
		message.WriteString(fmt.Sprintf("*Environment:* %s\n", env))
	}
	if severity != "" {
		message.WriteString(fmt.Sprintf("*嚴重程度:* %s\n", severity))
	}
	if namespace != "" {
		message.WriteString(fmt.Sprintf("*命名空間:* %s\n", namespace))
	}

	message.WriteString(fmt.Sprintf("*總警報數:* %d\n", len(req.Alerts)))
	if firingCount > 0 {
		message.WriteString(fmt.Sprintf("*觸發中:* %d\n", firingCount))
	}
	if resolvedCount > 0 {
		message.WriteString(fmt.Sprintf("*已解決:* %d\n", resolvedCount))
	}

	// 觸發中的警報詳情
	if firingCount > 0 {
		message.WriteString("\n🔥 *觸發中的警報:*\n")
		count := 0
		for _, alert := range req.Alerts {
			if status, ok := alert["status"].(string); ok && status == "firing" {
				count++
				message.WriteString(fmt.Sprintf("*警報 %d:*\n", count))

				// 顯示摘要
				if annotations, ok := alert["annotations"].(map[string]interface{}); ok {
					if summary, ok := annotations["summary"].(string); ok {
						message.WriteString(fmt.Sprintf("• 摘要: %s\n", summary))
					}
				}

				// 顯示啟動時間
				if startsAt, ok := alert["startsAt"].(string); ok && startsAt != "0001-01-01T00:00:00Z" {
					message.WriteString(fmt.Sprintf("• 開始時間: %s\n", startsAt))
				}

				// 顯示標籤
				if labels, ok := alert["labels"].(map[string]interface{}); ok {
					if pod, ok := labels["pod"].(string); ok && pod != "" {
						message.WriteString(fmt.Sprintf("• Pod: %s\n", pod))
					}
				}
				message.WriteString("\n")
			}
		}
	}

	// 已解決的警報詳情
	if resolvedCount > 0 {
		message.WriteString("✅ *已解決的警報:*\n")
		count := 0
		for _, alert := range req.Alerts {
			if status, ok := alert["status"].(string); ok && status == "resolved" {
				count++
				message.WriteString(fmt.Sprintf("*警報 %d:*\n", count))

				// 顯示摘要
				if annotations, ok := alert["annotations"].(map[string]interface{}); ok {
					if summary, ok := annotations["summary"].(string); ok {
						message.WriteString(fmt.Sprintf("• 摘要: %s\n", summary))
					}
				}

				// 顯示結束時間
				if endsAt, ok := alert["endsAt"].(string); ok && endsAt != "0001-01-01T00:00:00Z" {
					message.WriteString(fmt.Sprintf("• 解決時間: %s\n", endsAt))
				}

				// 顯示標籤
				if labels, ok := alert["labels"].(map[string]interface{}); ok {
					if pod, ok := labels["pod"].(string); ok && pod != "" {
						message.WriteString(fmt.Sprintf("• Pod: %s\n", pod))
					}
				}
				message.WriteString("\n")
			}
		}
	}

	// 添加外部連結
	if req.ExternalURL != "" {
		// 僅在配置允許時顯示外部連結（使用引擎當前配置）
		serviceManager := service.GetServiceManager()
		if te := serviceManager.GetTemplateEngine(); te != nil {
			if te.GetCurrentFormatOptions().ShowExternalURL.Enabled {
				message.WriteString(fmt.Sprintf("\n🔗 <%s|查看詳情>", req.ExternalURL))
			}
		}
	}

	return message.String()
}
