package template

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"alert-webhooks/pkg/logger"

	"gopkg.in/yaml.v3"
)

// TemplateConfig 模板配置結構
type TemplateConfig struct {
	Version             string            `yaml:"version"`
	SupportedLanguages  []LanguageConfig  `yaml:"supported_languages"`
	FallbackOrder      []string          `yaml:"fallback_order"`
	NamingConvention   NamingConfig      `yaml:"naming_convention"`
	FormatOptions      FormatOptions     `yaml:"format_options"`
}

// LanguageConfig 語言配置
type LanguageConfig struct {
	Code        string `yaml:"code"`
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
	Fallback    bool   `yaml:"fallback"`
}

// NamingConfig 命名規則配置
type NamingConfig struct {
	Prefix              string   `yaml:"prefix"`
	SupportedExtensions []string `yaml:"supported_extensions"`
	PriorityOrder       []string `yaml:"priority_order"`
}

// FormatOptions 格式化選項結構
type FormatOptions struct {
	ShowLinks struct {
		Enabled     bool   `yaml:"enabled"`
		Description string `yaml:"description"`
	} `yaml:"show_links"`
	ShowTimestamps struct {
		Enabled     bool   `yaml:"enabled"`
		Description string `yaml:"description"`
	} `yaml:"show_timestamps"`
	ShowExternalURL struct {
		Enabled     bool   `yaml:"enabled"`
		Description string `yaml:"description"`
	} `yaml:"show_external_url"`
	ShowGeneratorURL struct {
		Enabled     bool   `yaml:"enabled"`
		Description string `yaml:"description"`
	} `yaml:"show_generator_url"`
	ShowEmoji struct {
		Enabled     bool   `yaml:"enabled"`
		Description string `yaml:"description"`
	} `yaml:"show_emoji"`
	CompactMode struct {
		Enabled     bool   `yaml:"enabled"`
		Description string `yaml:"description"`
	} `yaml:"compact_mode"`
	MaxSummaryLength struct {
		Value       int    `yaml:"value"`
		Description string `yaml:"description"`
	} `yaml:"max_summary_length"`
}

// TemplateEngine 模板引擎
type TemplateEngine struct {
	templates map[string]*template.Template
	config    *TemplateConfig
}

// TemplateData 模板數據結構
type TemplateData struct {
	Status        string
	AlertName     string
	Env           string
	Severity      string
	Namespace     string
	TotalAlerts   int
	FiringCount   int
	ResolvedCount int
	Alerts        []AlertData
	ExternalURL   string
	FormatOptions FormatOptions
	Platform      string // 目標平台：telegram, slack
}

// AlertData 警報數據結構
type AlertData struct {
	Status       string
	Labels       map[string]string
	Annotations  map[string]string
	StartsAt     string
	EndsAt       string
	GeneratorURL string
}

// NewTemplateEngine 創建新的模板引擎
func NewTemplateEngine() *TemplateEngine {
	return &TemplateEngine{
		templates: make(map[string]*template.Template),
		config:    getDefaultConfig(),
	}
}

// LoadConfig 載入模板配置檔案
func (te *TemplateEngine) LoadConfig(configPath string) error {
	logger.Info("Loading template config", "template_engine",
		logger.String("config_path", configPath))
	
	data, err := os.ReadFile(configPath)
	if err != nil {
		logger.Warn("Failed to read config file, using default config", "template_engine",
			logger.String("config_path", configPath),
			logger.Err(err))
		te.config = getDefaultConfig()
		return nil // 不返回錯誤，使用預設配置
	}
	
	var fullConfig struct {
		TemplateConfig TemplateConfig `yaml:"template_config"`
	}
	
	if err := yaml.Unmarshal(data, &fullConfig); err != nil {
		logger.Warn("Failed to parse config file, using default config", "template_engine",
			logger.String("config_path", configPath),
			logger.Err(err))
		te.config = getDefaultConfig()
		return nil // 不返回錯誤，使用預設配置
	}
	
	te.config = &fullConfig.TemplateConfig
	logger.Info("Template config loaded successfully", "template_engine",
		logger.String("version", te.config.Version),
		logger.Int("supported_languages", len(te.config.SupportedLanguages)))
	
	return nil
}

// LoadConfigFromConfigs 從 configs 目錄載入模板配置
func (te *TemplateEngine) LoadConfigFromConfigs() error {
	configPaths := []string{
		"configs/alert_config.yaml",
		"./configs/alert_config.yaml",
		"../configs/alert_config.yaml",
	}
	
	for _, configPath := range configPaths {
		if err := te.LoadConfig(configPath); err == nil {
			logger.Info("Template config loaded from configs directory", "template_engine",
				logger.String("config_path", configPath))
			return nil
		}
	}
	
	logger.Warn("Failed to load template config from configs directory, using default config", "template_engine")
	te.config = getDefaultConfig()
	return nil
}

// LoadConfigWithProfile 載入指定配置檔案
func (te *TemplateEngine) LoadConfigWithProfile(profile string) error {
	var configPath string
	if profile == "" {
		configPath = filepath.Join("configs", "alert_config.yaml")
	} else {
		configPath = filepath.Join("configs", fmt.Sprintf("alert_config.%s.yaml", profile))
	}
	
	// 檢查檔案是否存在
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		logger.Warn("Config file not found, using default", "template",
			logger.String("config_path", configPath))
		return te.LoadConfigFromConfigs() // 回退到預設配置
	}
	
	logger.Info("Loading template config with profile", "template",
		logger.String("profile", profile),
		logger.String("config_path", configPath))
	
	return te.LoadConfig(configPath)
}

// ReloadConfigWithProfile 重新載入指定配置檔案
func (te *TemplateEngine) ReloadConfigWithProfile(profile string) error {
	if err := te.LoadConfigWithProfile(profile); err != nil {
		return err
	}
	
	logger.Info("Template config reloaded", "template",
		logger.String("profile", profile))
	return nil
}

// loadTemplateConfigFromFile 從配置文件載入模板配置
func loadTemplateConfigFromFile(profile string) *TemplateConfig {
	var configPath string
	if profile == "" || profile == "full" {
		configPath = filepath.Join("configs", "alert_config.yaml")
	} else {
		configPath = filepath.Join("configs", fmt.Sprintf("alert_config.%s.yaml", profile))
	}
	
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil
	}
	
	var fullConfig struct {
		TemplateConfig TemplateConfig `yaml:"template_config"`
	}
	
	if err := yaml.Unmarshal(data, &fullConfig); err != nil {
		return nil
	}
	
	return &fullConfig.TemplateConfig
}

// getDefaultConfigForMode 根據模式獲取預設配置
func getDefaultConfigForMode(mode string) *TemplateConfig {
	// 嘗試根據模式從對應的配置文件讀取配置
	var profile string
	if mode == "minimal" {
		profile = "minimal"
	} else {
		profile = "" // 預設使用 full 模式
	}
	
	if config := loadTemplateConfigFromFile(profile); config != nil {
		return config
	}
	
	// 如果無法讀取配置文件，根據模式返回適當的硬編碼配置
	if mode == "minimal" {
		return getMinimalDefaultConfig()
	}
	
	return getFullDefaultConfig()
}

// getDefaultConfig 獲取預設配置，優先從已載入的模板配置中讀取
func getDefaultConfig() *TemplateConfig {
	// 預設使用 full 模式
	return getDefaultConfigForMode("full")
}

// getFullDefaultConfig 獲取完整模式的硬編碼預設配置
func getFullDefaultConfig() *TemplateConfig {
	return &TemplateConfig{
		Version: "1.0.0",
		SupportedLanguages: []LanguageConfig{
			{Code: "eng", Name: "English", Description: "English template", Fallback: true},
			{Code: "tw", Name: "繁體中文", Description: "Traditional Chinese template", Fallback: false},
			{Code: "zh", Name: "简体中文", Description: "Simplified Chinese template", Fallback: false},
			{Code: "ko", Name: "한국어", Description: "Korean template", Fallback: false},
			{Code: "ja", Name: "日本語", Description: "Japanese template", Fallback: false},
		},
		FallbackOrder: []string{"eng", "tw", "zh", "ko", "ja", "en"},
		NamingConvention: NamingConfig{
			Prefix:              "alert_template_",
			SupportedExtensions: []string{".tmpl", ".j2"},
			PriorityOrder:       []string{".tmpl", ".j2"},
		},
		FormatOptions: FormatOptions{
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
			MaxSummaryLength: struct {
				Value       int    `yaml:"value"`
				Description string `yaml:"description"`
			}{Value: 200, Description: "摘要最大長度"},
		},
	}
}

// getMinimalDefaultConfig 獲取極簡模式的硬編碼預設配置
func getMinimalDefaultConfig() *TemplateConfig {
	return &TemplateConfig{
		Version: "1.0.0",
		SupportedLanguages: []LanguageConfig{
			{Code: "eng", Name: "English", Description: "English template", Fallback: true},
			{Code: "tw", Name: "繁體中文", Description: "Traditional Chinese template", Fallback: false},
			{Code: "zh", Name: "简体中文", Description: "Simplified Chinese template", Fallback: false},
			{Code: "ko", Name: "한국어", Description: "Korean template", Fallback: false},
			// 注意：minimal 模式不包含 ja (日本語)
		},
		FallbackOrder: []string{"eng", "tw", "zh", "en"},
		NamingConvention: NamingConfig{
			Prefix:              "alert_template_",
			SupportedExtensions: []string{".tmpl", ".j2"},
			PriorityOrder:       []string{".tmpl", ".j2"},
		},
		FormatOptions: FormatOptions{
			ShowLinks: struct {
				Enabled     bool   `yaml:"enabled"`
				Description string `yaml:"description"`
			}{Enabled: false, Description: "是否顯示超連結"},
			ShowTimestamps: struct {
				Enabled     bool   `yaml:"enabled"`
				Description string `yaml:"description"`
			}{Enabled: false, Description: "是否顯示時間戳"},
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
			}{Enabled: false, Description: "是否顯示表情符號"},
			CompactMode: struct {
				Enabled     bool   `yaml:"enabled"`
				Description string `yaml:"description"`
			}{Enabled: true, Description: "緊湊模式（簡化顯示）"},
			MaxSummaryLength: struct {
				Value       int    `yaml:"value"`
				Description string `yaml:"description"`
			}{Value: 100, Description: "摘要最大長度"},
		},
	}
}

// LoadTemplates 載入模板檔案 - 自動掃描並載入所有語系模板
func (te *TemplateEngine) LoadTemplates(templateDir string) error {
	logger.Info("Starting dynamic template loading", "template_engine",
		logger.String("template_dir", templateDir))
	
	// 檢查目錄是否存在
	if _, err := os.Stat(templateDir); err != nil {
		return fmt.Errorf("template directory does not exist: %s (%v)", templateDir, err)
	}
	
	// 掃描目錄中的模板檔案
	templateFiles, err := te.scanTemplateFiles(templateDir)
	if err != nil {
		return fmt.Errorf("failed to scan template files: %v", err)
	}
	
	if len(templateFiles) == 0 {
		return fmt.Errorf("no template files found in directory: %s", templateDir)
	}
	
	// 載入找到的模板檔案
	loadedCount := 0
	for language, templatePath := range templateFiles {
		if template, err := te.loadTemplate(templatePath); err == nil {
			te.templates[language] = template
			logger.Info("Template loaded successfully", "template_engine",
				logger.String("language", language),
				logger.String("path", templatePath))
			loadedCount++
		} else {
			logger.Warn("Failed to load template", "template_engine",
				logger.String("language", language),
				logger.String("path", templatePath),
				logger.Err(err))
		}
	}
	
	if loadedCount == 0 {
		return fmt.Errorf("failed to load any templates from directory: %s", templateDir)
	}
	
	logger.Info("Template loading completed", "template_engine",
		logger.String("template_dir", templateDir),
		logger.Int("loaded_count", loadedCount),
		logger.Int("total_found", len(templateFiles)))

	return nil
}

// scanTemplateFiles 掃描模板目錄，找到所有語系的模板檔案
func (te *TemplateEngine) scanTemplateFiles(templateDir string) (map[string]string, error) {
	templateFiles := make(map[string]string)
	
	// 從配置獲取設定
	supportedExtensions := te.config.NamingConvention.SupportedExtensions
	templatePrefix := te.config.NamingConvention.Prefix
	priorityOrder := te.config.NamingConvention.PriorityOrder
	
	// 讀取目錄內容
	files, err := os.ReadDir(templateDir)
	if err != nil {
		return nil, err
	}
	
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		
		fileName := file.Name()
		
		// 檢查是否是模板檔案
		if !strings.HasPrefix(fileName, templatePrefix) {
			continue
		}
		
		// 檢查副檔名
		var isSupported bool
		var ext string
		for _, supportedExt := range supportedExtensions {
			if strings.HasSuffix(fileName, supportedExt) {
				isSupported = true
				ext = supportedExt
				break
			}
		}
		
		if !isSupported {
			continue
		}
		
		// 提取語言代碼
		// 檔案名格式: alert_template_{language}.{ext}
		baseName := strings.TrimPrefix(fileName, templatePrefix)
		language := strings.TrimSuffix(baseName, ext)
		
		if language == "" {
			logger.Warn("Invalid template file name format", "template_engine",
				logger.String("file_name", fileName))
			continue
		}
		
		templatePath := filepath.Join(templateDir, fileName)
		
		// 根據配置的優先權順序選擇檔案
		if existing, exists := templateFiles[language]; exists {
			existingExt := filepath.Ext(existing)
			currentPriority := te.getExtensionPriority(ext, priorityOrder)
			existingPriority := te.getExtensionPriority(existingExt, priorityOrder)
			
			// 如果當前檔案優先權更高，替換現有檔案
			if currentPriority < existingPriority {
				templateFiles[language] = templatePath
				logger.Info("Template file found (replacing existing)", "template_engine",
					logger.String("language", language),
					logger.String("new_path", templatePath),
					logger.String("old_path", existing),
					logger.String("reason", "higher_priority"))
			} else {
				logger.Info("Template file found (keeping existing)", "template_engine",
					logger.String("language", language),
					logger.String("existing_path", existing),
					logger.String("skipped_path", templatePath),
					logger.String("reason", "lower_priority"))
			}
		} else {
			templateFiles[language] = templatePath
			logger.Info("Template file found", "template_engine",
				logger.String("language", language),
				logger.String("path", templatePath))
		}
	}
	
	return templateFiles, nil
}

// getExtensionPriority 獲取副檔名的優先權（數字越小優先權越高）
func (te *TemplateEngine) getExtensionPriority(ext string, priorityOrder []string) int {
	for i, priorityExt := range priorityOrder {
		if ext == priorityExt {
			return i
		}
	}
	return len(priorityOrder) // 如果不在優先順序中，設為最低優先權
}

// loadTemplate 載入單個模板檔案
func (te *TemplateEngine) loadTemplate(templatePath string) (*template.Template, error) {
	content, err := os.ReadFile(templatePath)
	if err != nil {
		return nil, err
	}

	var goTemplateContent string
	
	// 檢查檔案副檔名
	if strings.HasSuffix(templatePath, ".tmpl") {
		// 直接使用 Go template 語法
		goTemplateContent = string(content)
		logger.Info("Using direct Go template", "template_engine",
			logger.String("template_path", templatePath))
	} else {
		// 將 Jinja2 語法轉換為 Go template 語法
		goTemplateContent = te.convertJinja2ToGoTemplate(string(content))
		logger.Info("Using converted Jinja2 template", "template_engine",
			logger.String("template_path", templatePath))
	}
	
	// 添加調試信息
	logger.Info("Template conversion debug", "template_engine",
		logger.String("template_path", templatePath),
		logger.String("original_length", fmt.Sprintf("%d", len(content))),
		logger.String("converted_length", fmt.Sprintf("%d", len(goTemplateContent))))
	
	// 顯示轉換後的模板前幾行（調試用）
	lines := strings.Split(goTemplateContent, "\n")
	if len(lines) > 5 {
		logger.Info("Converted template preview", "template_engine",
			logger.String("preview", strings.Join(lines[:5], "\n")))
	}

	tmpl, err := template.New(filepath.Base(templatePath)).Funcs(template.FuncMap{
		"format_time": te.formatTimeForPlatform,
		"format_time_simple": te.formatTime,
		"add": func(a, b int) int { return a + b },
		"index": func(m map[string]string, key string) string { return m[key] },
		"format_text": te.formatTextForPlatform,
		"format_bold": te.formatBoldForPlatform,
		"format_italic": te.formatItalicForPlatform,
		"format_code": te.formatCodeForPlatform,
		"format_link": te.formatLinkForPlatform,
		"printf": fmt.Sprintf,
	}).Parse(goTemplateContent)
	if err != nil {
		return nil, fmt.Errorf("failed to parse template %s: %v", templatePath, err)
	}

	return tmpl, nil
}

// convertJinja2ToGoTemplate 將 Jinja2 語法轉換為 Go template 語法
func (te *TemplateEngine) convertJinja2ToGoTemplate(content string) string {
	// 簡化的轉換邏輯
	result := content
	
	// 處理變數引用（先處理，避免與控制結構衝突）
	result = strings.ReplaceAll(result, "{{ status }}", "{{ .Status }}")
	result = strings.ReplaceAll(result, "{{ alert_name }}", "{{ .AlertName }}")
	result = strings.ReplaceAll(result, "{{ env }}", "{{ .Env }}")
	result = strings.ReplaceAll(result, "{{ severity }}", "{{ .Severity }}")
	result = strings.ReplaceAll(result, "{{ namespace }}", "{{ .Namespace }}")
	result = strings.ReplaceAll(result, "{{ total_alerts }}", "{{ .TotalAlerts }}")
	result = strings.ReplaceAll(result, "{{ firing_count }}", "{{ .FiringCount }}")
	result = strings.ReplaceAll(result, "{{ resolved_count }}", "{{ .ResolvedCount }}")
	result = strings.ReplaceAll(result, "{{ externalURL }}", "{{ .ExternalURL }}")
	
	// 處理簡單的 if 語句
	result = strings.ReplaceAll(result, "{% if firing_count > 0 %}", "{{ if gt .FiringCount 0 }}")
	result = strings.ReplaceAll(result, "{% if resolved_count > 0 %}", "{{ if gt .ResolvedCount 0 }}")
	result = strings.ReplaceAll(result, "{% elif resolved_count > 0 %}", "{{ else if gt .ResolvedCount 0 }}")
	result = strings.ReplaceAll(result, "{% else %}", "{{ else }}")
	result = strings.ReplaceAll(result, "{% endif %}", "{{ end }}")
	
	// 處理 for 迴圈
	result = strings.ReplaceAll(result, "{% for alert in alerts %}", "{{ range $index, $alert := .Alerts }}")
	result = strings.ReplaceAll(result, "{% endfor %}", "{{ end }}")
	
	// 處理迴圈內的變數
	result = strings.ReplaceAll(result, "{{ loop.index }}", "{{ add $index 1 }}")
	result = strings.ReplaceAll(result, "alert.status", "$alert.Status")
	result = strings.ReplaceAll(result, "alert.annotations.summary", "index $alert.Annotations \"summary\"")
	result = strings.ReplaceAll(result, "alert.labels.pod", "index $alert.Labels \"pod\"")
	result = strings.ReplaceAll(result, "alert.startsAt", "$alert.StartsAt")
	result = strings.ReplaceAll(result, "alert.endsAt", "$alert.EndsAt")
	result = strings.ReplaceAll(result, "alert.generatorURL", "$alert.GeneratorURL")
	
	// 處理條件語句
	result = strings.ReplaceAll(result, "{% if alert.status == \"firing\" %}", "{{ if eq $alert.Status \"firing\" }}")
	result = strings.ReplaceAll(result, "{% if alert.status == \"resolved\" %}", "{{ if eq $alert.Status \"resolved\" }}")
	result = strings.ReplaceAll(result, "{% if alert.endsAt != \"0001-01-01T00:00:00Z\" %}", "{{ if ne $alert.EndsAt \"0001-01-01T00:00:00Z\" }}")
	result = strings.ReplaceAll(result, "{% if alert.generatorURL %}", "{{ if $alert.GeneratorURL }}")
	result = strings.ReplaceAll(result, "{% if externalURL %}", "{{ if .ExternalURL }}")
	
	// 處理函數調用
	result = strings.ReplaceAll(result, "format_time(alert.startsAt)", "format_time $alert.StartsAt")
	result = strings.ReplaceAll(result, "format_time(alert.endsAt)", "format_time $alert.EndsAt")

	return result
}

// RenderTemplate 渲染模板
func (te *TemplateEngine) RenderTemplate(language string, data TemplateData) (string, error) {
	tmpl, exists := te.templates[language]
	if !exists {
		return "", fmt.Errorf("template for language '%s' not found", language)
	}

	// 只有在 FormatOptions 為空時才使用配置文件的默認值
	// 這樣可以保留平台 handler 傳遞的自定義 FormatOptions
	if te.config != nil && data.FormatOptions == (FormatOptions{}) {
		data.FormatOptions = te.config.FormatOptions
		logger.Debug("Using config FormatOptions as fallback", "template_engine")
	} else {
		logger.Debug("Using provided FormatOptions (not overriding)", "template_engine")
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute template: %v", err)
	}

	return buf.String(), nil
}

// RenderTemplateForPlatform 為特定平台渲染模板
func (te *TemplateEngine) RenderTemplateForPlatform(language, platform string, data TemplateData) (string, error) {
	// 設置平台信息
	data.Platform = platform
	
	// Debug: 檢查 FormatOptions
	logger.Debug("Template engine received FormatOptions", "template_engine",
		logger.String("platform", platform),
		logger.Bool("ShowEmoji", data.FormatOptions.ShowEmoji.Enabled),
		logger.Bool("ShowTimestamps", data.FormatOptions.ShowTimestamps.Enabled),
		logger.Bool("ShowGeneratorURL", data.FormatOptions.ShowGeneratorURL.Enabled),
		logger.Bool("ShowExternalURL", data.FormatOptions.ShowExternalURL.Enabled))
	
	// 詳細調試：檢查整個 FormatOptions 結構
	logger.Debug("FormatOptions structure debug", "template_engine",
		logger.String("ShowEmoji.Description", data.FormatOptions.ShowEmoji.Description),
		logger.String("ShowTimestamps.Description", data.FormatOptions.ShowTimestamps.Description))
	
	// 現在每個平台的 handler 都會正確提供 FormatOptions，不需要再檢查
	logger.Debug("Using provided FormatOptions", "template_engine", logger.String("platform", platform))
	
	// 額外調試：檢查傳遞給模板的完整數據
	logger.Debug("Template data debug", "template_engine",
		logger.String("Status", data.Status),
		logger.String("AlertName", data.AlertName),
		logger.Int("FiringCount", data.FiringCount),
		logger.Bool("FormatOptions.ShowEmoji.Enabled", data.FormatOptions.ShowEmoji.Enabled))
	
	renderedMessage, err := te.RenderTemplate(language, data)
	if err != nil {
		logger.Error("Template rendering failed", "template_engine", logger.Err(err))
		return "", err
	}
	
	// 調試渲染結果
	logger.Debug("Template rendered result", "template_engine",
		logger.String("platform", platform),
		logger.String("language", language),
		logger.Int("result_length", len(renderedMessage)),
		logger.String("result_preview", func() string {
			if len(renderedMessage) > 200 {
				return renderedMessage[:200] + "..."
			}
			return renderedMessage
		}()))
	
	return renderedMessage, nil
}

// GetAvailableLanguages 獲取已載入模板的語言列表
func (te *TemplateEngine) GetAvailableLanguages() []string {
	var languages []string
	for lang := range te.templates {
		languages = append(languages, lang)
	}
	return languages
}

// GetSupportedLanguages 獲取配置中定義的支援語言列表
func (te *TemplateEngine) GetSupportedLanguages() []string {
	if te.config == nil {
		return []string{}
	}
	
	var languages []string
	for _, langConfig := range te.config.SupportedLanguages {
		languages = append(languages, langConfig.Code)
	}
	return languages
}

// GetSupportedLanguageDetails 獲取支援語言的詳細資訊
func (te *TemplateEngine) GetSupportedLanguageDetails() []LanguageConfig {
	if te.config == nil {
		return []LanguageConfig{}
	}
	return te.config.SupportedLanguages
}

// HasLanguage 檢查是否支援指定語言
func (te *TemplateEngine) HasLanguage(language string) bool {
	_, exists := te.templates[language]
	return exists
}

// GetDefaultLanguage 獲取預設語言（如果指定語言不存在）
func (te *TemplateEngine) GetDefaultLanguage(preferredLanguage string) string {
	// 如果指定語言存在，直接返回
	if te.HasLanguage(preferredLanguage) {
		return preferredLanguage
	}
	
	// 使用配置的語言回退順序
	for _, fallback := range te.config.FallbackOrder {
		if te.HasLanguage(fallback) {
			return fallback
		}
	}
	
	// 如果都沒有，返回第一個可用的語言
	languages := te.GetAvailableLanguages()
	if len(languages) > 0 {
		return languages[0]
	}
	
	return preferredLanguage // 如果沒有任何模板，返回原始語言
}

// ReloadTemplates 重新載入模板（用於動態更新）
func (te *TemplateEngine) ReloadTemplates(templateDir string) error {
	// 清空現有模板
	te.templates = make(map[string]*template.Template)
	
	// 重新載入
	return te.LoadTemplates(templateDir)
}

// ValidateTemplate 驗證模板語法
func (te *TemplateEngine) ValidateTemplate(templatePath string) error {
	_, err := te.loadTemplate(templatePath)
	return err
}

// formatTime 格式化時間字符串
func (te *TemplateEngine) formatTime(timeStr string) string {
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

// escapeMarkdownV2 轉義 MarkdownV2 特殊字符
func (te *TemplateEngine) escapeMarkdownV2(text string) string {
	// MarkdownV2 需要轉義的字符
	specialChars := []string{
		"_", "*", "[", "]", "(", ")", "~", "`", ">", "#", "+", "-", "=", "|", "{", "}", ".", "!",
	}
	
	result := text
	for _, char := range specialChars {
		result = strings.ReplaceAll(result, char, "\\"+char)
	}
	
	return result
}

// formatTextForPlatform 根據平台格式化普通文字
func (te *TemplateEngine) formatTextForPlatform(platform, text string) string {
	// 回退到簡單處理，避免過度轉義
	return text
}

// formatBoldForPlatform 根據平台格式化粗體文字
func (te *TemplateEngine) formatBoldForPlatform(platform, text string) string {
	switch platform {
	case "telegram":
		// Telegram HTML 格式
		return "<b>" + text + "</b>"
	default:
		// 其他平台使用標準 Markdown 粗體格式
		return "*" + text + "*"
	}
}

// formatItalicForPlatform 根據平台格式化斜體文字
func (te *TemplateEngine) formatItalicForPlatform(platform, text string) string {
	switch platform {
	case "telegram":
		// Telegram HTML 格式
		return "<i>" + text + "</i>"
	default:
		// 其他平台使用標準 Markdown 斜體格式
		return "_" + text + "_"
	}
}

// formatCodeForPlatform 根據平台格式化代碼文字
func (te *TemplateEngine) formatCodeForPlatform(platform, text string) string {
	switch platform {
	case "telegram":
		// Telegram HTML 格式
		return "<code>" + text + "</code>"
	default:
		// 其他平台使用標準 Markdown 代碼格式
		return "`" + text + "`"
	}
}

// formatLinkForPlatform 根據平台格式化連結
func (te *TemplateEngine) formatLinkForPlatform(platform, url, text string) string {
	if text == "" {
		text = url
	}
	
	switch platform {
	case "telegram":
		// Telegram HTML 格式：<a href="url">text</a>
		// HTML 實體轉義 URL 中的特殊字符
		escapedURL := strings.ReplaceAll(url, "&", "&amp;")
		escapedURL = strings.ReplaceAll(escapedURL, "<", "&lt;")
		escapedURL = strings.ReplaceAll(escapedURL, ">", "&gt;")
		return "<a href=\"" + escapedURL + "\">" + text + "</a>"
	case "slack":
		// Slack 格式：<url|text>
		return "<" + url + "|" + text + ">"
	case "discord":
		// Discord 支援標準 Markdown
		return "[" + text + "](" + url + ")"
	default:
		// 預設使用標準 Markdown
		return "[" + text + "](" + url + ")"
	}
}

// escapeMarkdownV2Text 轉義 MarkdownV2 文字中的特殊字符
func (te *TemplateEngine) escapeMarkdownV2Text(text string) string {
	// MarkdownV2 需要轉義的字符，但要謹慎處理
	// 只轉義在普通文字中需要轉義的字符
	specialChars := []string{
		"_", "*", "[", "]", "(", ")", "~", "`", ">", "#", "+", "=", "|", "{", "}", ".", "!",
		// 暫時移除 "-"，因為它在很多情況下不需要轉義
	}
	
	result := text
	for _, char := range specialChars {
		result = strings.ReplaceAll(result, char, "\\"+char)
	}
	
	// 只在特定情況下轉義 "-"（避免過度轉義）
	// 在 MarkdownV2 中，"-" 只在行首作為列表項時需要轉義
	// 在普通文字中通常不需要轉義
	
	return result
}

// formatTimeForPlatform 根據平台格式化時間
func (te *TemplateEngine) formatTimeForPlatform(platform, timeStr string) string {
	// 所有平台都使用標準 Markdown，不需要特殊轉義
	return te.formatTime(timeStr)
}

// GetMinimalDefaultConfig returns the minimal mode default configuration
func (te *TemplateEngine) GetMinimalDefaultConfig() *TemplateConfig {
	// Try to load from alert_config.minimal.yaml first
	minimalConfig, err := te.loadMinimalConfig()
	if err != nil {
		logger.Warn("Failed to load minimal config, using hardcoded defaults", "template_engine", logger.Err(err))
		return getMinimalDefaultConfig()
	}
	return minimalConfig
}

// loadMinimalConfig loads the minimal configuration from alert_config.minimal.yaml
func (te *TemplateEngine) loadMinimalConfig() (*TemplateConfig, error) {
	configPaths := []string{
		"configs/alert_config.minimal.yaml",
		"./configs/alert_config.minimal.yaml",
		"../configs/alert_config.minimal.yaml",
	}
	
	for _, configPath := range configPaths {
		config, err := te.loadConfigFromFile(configPath)
		if err == nil {
			logger.Info("Minimal config loaded successfully", "template_engine",
				logger.String("config_path", configPath),
				logger.Bool("ShowTimestamps", config.FormatOptions.ShowTimestamps.Enabled))
			return config, nil
		}
	}
	
	return nil, fmt.Errorf("minimal config file not found in any expected location")
}

// loadConfigFromFile loads configuration from a specific file
func (te *TemplateEngine) loadConfigFromFile(configPath string) (*TemplateConfig, error) {
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("config file does not exist: %s", configPath)
	}
	
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %v", err)
	}
	
	var fullConfig struct {
		TemplateConfig TemplateConfig `yaml:"template_config"`
	}
	
	if err := yaml.Unmarshal(data, &fullConfig); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %v", err)
	}
	
	return &fullConfig.TemplateConfig, nil
}

// GetFullDefaultConfig returns the full mode default configuration  
func (te *TemplateEngine) GetFullDefaultConfig() *TemplateConfig {
	return getFullDefaultConfig()
}

// GetCurrentFormatOptions 取得目前載入配置的 FormatOptions
func (te *TemplateEngine) GetCurrentFormatOptions() FormatOptions {
	if te != nil && te.config != nil {
		return te.config.FormatOptions
	}
	return getFullDefaultConfig().FormatOptions
}

// getFormatOptionsForPlatform 根據平台配置返回對應的 FormatOptions
func (te *TemplateEngine) getFormatOptionsForPlatform(platform string) FormatOptions {
	// 由於循環依賴問題，這裡暫時使用配置文件中的默認值
	// 每個平台的 handler 應該自己設置正確的 FormatOptions
	
	// 如果模板引擎有配置，使用它；否則使用 full 模式
	if te.config != nil {
		return te.config.FormatOptions
	}
	
	// 默認使用 full 模式
	return getFullDefaultConfig().FormatOptions
}
