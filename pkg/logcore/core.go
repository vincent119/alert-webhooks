package logcore

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"alert-webhooks/config"
)

var (
	// Log 全局日誌實例
	Log  *zap.Logger
	once sync.Once
	// zapGlobalLevel 全局日誌級別，支援動態修改
	zapGlobalLevel = zap.NewAtomicLevel()
)

// Field 是 zap.Field 的別名，使用時更簡潔
type Field = zap.Field

// LogConfig 日誌配置結構
type LogConfig struct {
	Level         string   // 日誌級別
	Format        string   // 輸出格式：json 或 console
	Outputs       []string // 輸出目標：console, file
	LogPath       string   // 日誌檔案路徑
	FileName      string   // 日誌檔案名稱
	MaxSize       int      // 單個日誌檔案最大大小（MB）
	MaxAge        int      // 日誌檔案保存天數
	MaxBackups    int      // 最大備份數
	Compress      bool     // 是否壓縮舊日誌
	AddCaller     bool     // 是否添加調用者信息
	AddStacktrace bool     // 是否添加堆疊追蹤
}

// sqlProcessingCore 是一個處理 SQL 字段的核心包裝器
type sqlProcessingCore struct {
	zapcore.Core
}

// With 實現 zapcore.Core 接口
func (c *sqlProcessingCore) With(fields []zapcore.Field) zapcore.Core {
	// 處理字段中的 SQL
	for i := range fields {
		if fields[i].Key == "sql" && fields[i].Type == zapcore.StringType {
			fields[i].String = processSQLString(fields[i].String)
		}
	}
	return &sqlProcessingCore{Core: c.Core.With(fields)}
}

// Check 實現 zapcore.Core 接口
func (c *sqlProcessingCore) Check(ent zapcore.Entry, ce *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	if c.Core.Enabled(ent.Level) {
		return ce.AddCore(ent, c)
	}
	return ce
}

// Write 實現 zapcore.Core 接口
func (c *sqlProcessingCore) Write(ent zapcore.Entry, fields []zapcore.Field) error {
	// 處理消息中的反斜線
	ent.Message = strings.ReplaceAll(ent.Message, "\\", "")

	// 處理字段中的 SQL
	for i := range fields {
		if fields[i].Key == "sql" && fields[i].Type == zapcore.StringType {
			fields[i].String = processSQLString(fields[i].String)
		}
	}

	return c.Core.Write(ent, fields)
}

// InitLogger 初始化日誌配置
func InitLogger(level string, development bool) {
	once.Do(func() {
		// 構建日誌配置
		logConfig := getLogConfig(level, development)

		// 設置日誌級別
		var logLevel zapcore.Level
		switch strings.ToLower(logConfig.Level) {
		case "debug":
			logLevel = zap.DebugLevel
		case "info":
			logLevel = zap.InfoLevel
		case "warn":
			logLevel = zap.WarnLevel
		case "error":
			logLevel = zap.ErrorLevel
		case "fatal":
			logLevel = zap.FatalLevel
		default:
			logLevel = zap.InfoLevel
		}

		// 設置全局級別，支援動態修改
		zapGlobalLevel.SetLevel(logLevel)

		// 配置編碼器
		encoderConfig := zapcore.EncoderConfig{
			TimeKey:        "ts",
			LevelKey:       "level",
			NameKey:        "logger",
			CallerKey:      "caller",
			FunctionKey:    zapcore.OmitKey,
			MessageKey:     "msg",
			StacktraceKey:  "stacktrace",
			LineEnding:     zapcore.DefaultLineEnding,
			EncodeLevel:    zapcore.CapitalLevelEncoder,
			EncodeTime:     zapcore.ISO8601TimeEncoder,
			EncodeDuration: zapcore.StringDurationEncoder,
			EncodeCaller:   zapcore.ShortCallerEncoder,
			// // 添加自定義字符串編碼器，處理特殊字符
			// EncodeString: func(s string, enc zapcore.PrimitiveArrayEncoder) {
			// 	// 處理字串中的反斜線
			// 	cleanedStr := strings.ReplaceAll(s, "\\\\", "\\")
			// 	enc.AppendString(cleanedStr)
			// },
		}

		// 定義日誌輸出
		var outputs []zapcore.Core

		// 處理輸出目標
		for _, output := range logConfig.Outputs {
			switch strings.ToLower(output) {
			case "console":
				// 控制台輸出
				var encoder zapcore.Encoder
				if strings.ToLower(logConfig.Format) == "json" {
					// 創建自定義的 JSON 編碼器配置
					jsonEncoderConfig := encoderConfig
					// 自定義 JSON 編碼選項
					jsonEncoderConfig.EncodeTime = func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
						enc.AppendString(t.Format(time.RFC3339))
					}
					// 使用不轉義 HTML 字符的選項創建 JSON 編碼器
					encoder = zapcore.NewJSONEncoder(jsonEncoderConfig)
				} else {
					encoder = zapcore.NewConsoleEncoder(encoderConfig)
				}
				consoleOutput := zapcore.Lock(os.Stdout)
				outputs = append(outputs, zapcore.NewCore(encoder, consoleOutput, zapGlobalLevel))

			case "file":
				// 檔案輸出
				var encoder zapcore.Encoder
				if strings.ToLower(logConfig.Format) == "json" {
					// 創建自定義的 JSON 編碼器配置
					jsonEncoderConfig := encoderConfig
					// 自定義 JSON 編碼選項
					jsonEncoderConfig.EncodeTime = func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
						enc.AppendString(t.Format(time.RFC3339))
					}
					// 不使用 WithEscapeHTMLOption，而是直接創建編碼器
					encoder = zapcore.NewJSONEncoder(jsonEncoderConfig)
				} else {
					encoder = zapcore.NewConsoleEncoder(encoderConfig)
				}

				// 確保日誌目錄存在
				logDir := logConfig.LogPath
				if logDir == "" {
					logDir = "./logs"
				}
				if err := os.MkdirAll(logDir, 0755); err != nil {
					panic("無法建立日誌目錄: " + err.Error())
				}

				// 確定檔案名稱
				var logFileName string
				if logConfig.FileName != "" {
					logFileName = logConfig.FileName
				} else {
					now := time.Now()
					logFileName = now.Format("2006-01-02") + ".log"
				}

				// 開啟日誌檔案
				logFilePath := filepath.Join(logDir, logFileName)
				logFile, err := os.OpenFile(
					logFilePath,
					os.O_CREATE|os.O_APPEND|os.O_WRONLY,
					0644,
				)
				if err != nil {
					panic("無法開啟日誌檔案: " + err.Error())
				}
				// 建立日誌實例

				// 建立核心日誌
				core := zapcore.NewTee(outputs...)

				Log = zap.New(core)

				// 使用 WrapCore 來替代 EncodeValue
				Log = Log.WithOptions(zap.WrapCore(func(c zapcore.Core) zapcore.Core {
					return &sqlProcessingCore{Core: c}
				}))

				// 添加其他選項
				//var options []zap.Option

				fileOutput := zapcore.Lock(logFile)
				outputs = append(outputs, zapcore.NewCore(encoder, fileOutput, zapGlobalLevel))
			}
		}

		// 如果沒有指定任何輸出，預設使用控制台輸出
		if len(outputs) == 0 {
			consoleEncoder := zapcore.NewConsoleEncoder(encoderConfig)
			consoleOutput := zapcore.Lock(os.Stdout)
			outputs = append(outputs, zapcore.NewCore(consoleEncoder, consoleOutput, zapGlobalLevel))
		}

		// 建立核心日誌
		core := zapcore.NewTee(outputs...)

		// 建立日誌實例
		Log = zap.New(core)
		// 添加處理反斜線的 hook
		Log = Log.WithOptions(zap.Hooks(func(entry zapcore.Entry) error {
			cleanedMessage := strings.ReplaceAll(entry.Message, "\\", "")
			entry.Message = cleanedMessage
			return nil
		}))

		var options []zap.Option

		if logConfig.AddCaller {
			options = append(options, zap.AddCaller())
		}

		if logConfig.AddStacktrace {
			options = append(options, zap.AddStacktrace(zapcore.ErrorLevel))
		}

		if development {
			options = append(options, zap.Development())
		}

		if len(options) > 0 {
			Log = Log.WithOptions(options...)
		}

		// 替換全局 logger
		zap.ReplaceGlobals(Log)

		// 記錄日誌系統初始化信息
		Log.Info("日誌系統初始化完成",
			String("level", logConfig.Level),
			String("format", logConfig.Format),
			Any("outputs", logConfig.Outputs),
			String("path", logConfig.LogPath),
			String("file", logConfig.FileName),
		)
	})
}

// 從配置中獲取日誌配置
func getLogConfig(level string, development bool) LogConfig {
	// 建立預設配置
	logConfig := LogConfig{
		Level:         level,
		Format:        "console",           // 預設為控制台格式
		Outputs:       []string{"console"}, // 預設僅輸出到控制台
		LogPath:       "./logs",
		MaxSize:       100,
		MaxAge:        30,
		MaxBackups:    10,
		Compress:      true,
		AddCaller:     true,
		AddStacktrace: development,
	}

	// 從配置檔案中讀取設定
	conf := &config.Conf.Log

	// 如果配置中指定了日誌級別，則使用配置中的
	if conf.Level != "" {
		logConfig.Level = conf.Level
	}

	// 輸出格式
	if conf.Format != "" {
		format := strings.ToLower(conf.Format)
		if format == "json" || format == "console" {
			logConfig.Format = format
		}
	}

	// 輸出目標
	if conf.Outputs != "" {
		outputs := strings.Split(conf.Outputs, ",")
		validOutputs := make([]string, 0)
		for _, output := range outputs {
			output = strings.TrimSpace(output)
			if output == "console" || output == "file" {
				validOutputs = append(validOutputs, output)
			}
		}
		if len(validOutputs) > 0 {
			logConfig.Outputs = validOutputs
		}
	}

	// 日誌路徑
	if config.Conf.Log.LogPath != "" {
		logConfig.LogPath = config.Conf.Log.LogPath
	}

	// 日誌文件名
	if config.Conf.Log.LogFile != "" {
		logConfig.FileName = config.Conf.Log.LogFile
	}

	// 其他配置
	if conf.MaxSize > 0 {
		logConfig.MaxSize = conf.MaxSize
	}

	if conf.MaxAge > 0 {
		logConfig.MaxAge = conf.MaxAge
	}

	if conf.MaxBackups > 0 {
		logConfig.MaxBackups = conf.MaxBackups
	}

	logConfig.Compress = conf.Compress

	// 是否添加調用者資訊
	logConfig.AddCaller = conf.AddCaller

	// 是否添加堆疊追蹤
	logConfig.AddStacktrace = conf.AddStacktrace

	return logConfig
}

// Debug 快速訪問 Debug 級別日誌
func Debug(msg, category string, fields ...Field) {
	if Log != nil {
		// 對 SQL 查詢做特殊處理，擴大匹配範圍，捕獲所有 SQL 相關訊息
		if strings.Contains(strings.ToLower(msg), "sql") {
			for i, field := range fields {
				if field.Key == "sql" && field.Type == zapcore.StringType {
					// 替換 SQL 中的反斜線
					fields[i].String = processSQLString(field.String)
				}
			}
		}

		// 移除反斜線
		cleanedMsg := strings.ReplaceAll(msg, "\\", "")
		// 處理 fields 中的字符串值
		cleanedFields := cleanFields(fields)
		Log.Debug(cleanedMsg, append([]Field{String("category", category)}, cleanedFields...)...)
	}
}

// Info 快速訪問 Info 級別日誌
func Info(msg, category string, fields ...Field) {
	if Log != nil {
		// 對 SQL 查詢做特殊處理，擴大匹配範圍，捕獲所有 SQL 相關訊息
		if strings.Contains(strings.ToLower(msg), "sql") {
			for i, field := range fields {
				if field.Key == "sql" && field.Type == zapcore.StringType {
					// 替換 SQL 中的反斜線
					fields[i].String = processSQLString(field.String)
				}
			}
		}

		// 移除反斜線
		cleanedMsg := strings.ReplaceAll(msg, "\\", "")
		// 處理 fields 中的字符串值
		cleanedFields := cleanFields(fields)
		Log.Info(cleanedMsg, append([]Field{String("category", category)}, cleanedFields...)...)
	}
}

// Warn 快速訪問 Warn 級別日誌
func Warn(msg, category string, fields ...Field) {
	if Log != nil {
		// 對 SQL 查詢做特殊處理，擴大匹配範圍，捕獲所有 SQL 相關訊息
		if strings.Contains(strings.ToLower(msg), "sql") {
			for i, field := range fields {
				if field.Key == "sql" && field.Type == zapcore.StringType {
					// 替換 SQL 中的反斜線
					fields[i].String = processSQLString(field.String)
				}
			}
		}

		// 移除反斜線
		cleanedMsg := strings.ReplaceAll(msg, "\\", "")
		// 處理 fields 中的字符串值
		cleanedFields := cleanFields(fields)
		Log.Warn(cleanedMsg, append([]Field{String("category", category)}, cleanedFields...)...)
	}
}

// Error 快速訪問 Error 級別日誌
func Error(msg, category string, fields ...Field) {
	if Log != nil {
		// 對 SQL 查詢做特殊處理，擴大匹配範圍，捕獲所有 SQL 相關訊息
		if strings.Contains(strings.ToLower(msg), "sql") {
			for i, field := range fields {
				if field.Key == "sql" && field.Type == zapcore.StringType {
					// 替換 SQL 中的反斜線
					fields[i].String = processSQLString(field.String)
				}
			}
		}

		// 移除反斜線
		cleanedMsg := strings.ReplaceAll(msg, "\\", "")
		// 處理 fields 中的字符串值
		cleanedFields := cleanFields(fields)
		Log.Error(cleanedMsg, append([]Field{String("category", category)}, cleanedFields...)...)
	}
}

// Fatal 快速訪問 Fatal 級別日誌
func Fatal(msg, category string, fields ...Field) {
	if Log != nil {
		// 對 SQL 查詢做特殊處理，擴大匹配範圍，捕獲所有 SQL 相關訊息
		if strings.Contains(strings.ToLower(msg), "sql") {
			for i, field := range fields {
				if field.Key == "sql" && field.Type == zapcore.StringType {
					// 替換 SQL 中的反斜線
					fields[i].String = processSQLString(field.String)
				}
			}
		}

		// 移除反斜線
		cleanedMsg := strings.ReplaceAll(msg, "\\", "")
		// 處理 fields 中的字符串值
		cleanedFields := cleanFields(fields)
		Log.Fatal(cleanedMsg, append([]Field{String("category", category)}, cleanedFields...)...)
	}
}

// String 創建字符串類型的字段
func String(key, value string) Field {
	return zap.String(key, value)
}

// Int 創建整數類型的字段
func Int(key string, value int) Field {
	return zap.Int(key, value)
}

// Int64 創建 int64 類型的字段
func Int64(key string, value int64) Field {
	return zap.Int64(key, value)
}

// Float64 創建浮點類型的字段
func Float64(key string, value float64) Field {
	return zap.Float64(key, value)
}

// Bool 創建布爾類型的字段
func Bool(key string, value bool) Field {
	return zap.Bool(key, value)
}

// Err 創建錯誤類型的字段
func Err(err error) Field {
	return zap.Error(err)
}

// Any 創建任意類型的字段
func Any(key string, value interface{}) Field {
	return zap.Any(key, value)
}

// Duration 創建時間間隔類型的字段
func Duration(key string, value time.Duration) Field {
	return zap.Duration(key, value)
}

// Time 創建時間類型的字段
func Time(key string, value time.Time) Field {
	return zap.Time(key, value)
}

// SetLevel 動態設置日誌級別
func SetLevel(level string) {
	var zapLevel zapcore.Level
	switch strings.ToLower(level) {
	case "debug":
		zapLevel = zap.DebugLevel
	case "info":
		zapLevel = zap.InfoLevel
	case "warn":
		zapLevel = zap.WarnLevel
	case "error":
		zapLevel = zap.ErrorLevel
	case "fatal":
		zapLevel = zap.FatalLevel
	default:
		zapLevel = zap.InfoLevel
	}
	zapGlobalLevel.SetLevel(zapLevel)
	Info("日誌級別已變更", "system", String("level", level))
}

// cleanFields 函數中的處理邏輯
func cleanFields(fields []Field) []Field {
	cleanedFields := make([]Field, len(fields))
	for i, field := range fields {
		cleanedFields[i] = field // 先完整複製

		if field.Type == zapcore.StringType {
			// 特別處理 SQL 查詢
			if field.Key == "sql" {
				// 使用新的專門函數處理 SQL 字串
				cleanedFields[i].String = processSQLString(field.String)
			} else {
				// 一般字符串，移除反斜線但保持一些特殊字符的轉義
				cleanedFields[i].String = strings.ReplaceAll(field.String, "\\\\", "\\")
			}
		}
	}
	return cleanedFields
}

// 新增一個專門處理 SQL 字串的函數
func processSQLString(sql string) string {
	// 移除所有雙重反斜線
	sql = strings.ReplaceAll(sql, "\\\\", "\\")

	// 移除引號轉義的反斜線
	sql = strings.ReplaceAll(sql, "\\\"", "\"")

	// 移除單引號轉義的反斜線
	sql = strings.ReplaceAll(sql, "\\'", "'")

	return sql
}
