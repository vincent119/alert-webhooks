package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"alert-webhooks/config"
	"alert-webhooks/pkg/logger"
	"alert-webhooks/pkg/service"
	"alert-webhooks/pkg/trace"
	"alert-webhooks/pkg/watcher"
	"alert-webhooks/routes"

	"github.com/vincent119/commons/graceful"
)

const (
	shutdownTimeout   = 30 * time.Second
	readHeaderTimeout = 5 * time.Second
	readTimeout       = 15 * time.Second
	writeTimeout      = 30 * time.Second
	idleTimeout       = 60 * time.Second
	maxHeaderBytes    = 1 << 20
)

var mainString = "main"

// @title           Alert Webhooks API
// @version         1.0
// @description     這是 Alert Webhooks 的 Swagger 文件（由 swag 自動產生）。
// @schemes         https http
// @BasePath        /api/v1
// @produce         json
// @consume         json
// @contact.name    API Support
// @contact.email   vincent119@gmail.com
// @license.name    MIT
// @license.url     https://opensource.org/licenses/MIT
// @securityDefinitions.basic  BasicAuth
// @security  BasicAuth
func main() {
	// 初始化配置
	if err := config.Init(); err != nil {
		fmt.Fprintf(os.Stderr, "初始化設定失敗：%v\n", err)
		return
	}
	cfg := config.Current()

	// 初始化日誌系統
	loggerCleanup, err := logger.Configure(cfg.Log, cfg.App.Mode == "development")
	if err != nil {
		fmt.Fprintf(os.Stderr, "初始化日誌系統失敗：%v\n", err)
		return
	}

	// 初始化 OpenTelemetry TracerProvider
	traceShutdown, err := trace.InitTracerProvider(
		context.Background(),
		cfg.Trace,
		cfg.App.AppName,
	)
	if err != nil {
		logger.Error("初始化 TracerProvider 失敗", mainString, logger.Err(err))
		if cleanupErr := loggerCleanup(); cleanupErr != nil {
			fmt.Fprintf(os.Stderr, "關閉日誌系統失敗：%v\n", cleanupErr)
		}
		return
	}

	// 初始化服務
	serviceManager := service.GetServiceManager()
	if err := serviceManager.InitServices(); err != nil {
		logger.Warn("Failed to initialize services, some features may not be available", mainString, logger.Err(err))
	}

	startHTTPServer(traceShutdown, loggerCleanup)
}

// startHTTPServer 啟動HTTP服務並處理優雅關閉
func startHTTPServer(traceShutdown func(context.Context) error, loggerCleanup func() error) {
	cfg := config.Current()
	port := cfg.App.Port
	serverAddr := ":" + port
	router := routes.DefaultRoute()
	server := newHTTPServer(serverAddr, router)
	configWatcher := watcher.NewConfigWatcher()

	if err := graceful.Run(
		func(ctx context.Context) error {
			if err := configWatcher.Start(ctx); err != nil {
				logger.Warn("Failed to start config watcher", mainString, logger.Err(err))
			}

			logger.Info("Starting application...", mainString,
				logger.String("mode", cfg.App.Mode),
				logger.String("version", cfg.App.Version),
			)
			logger.Info("Service started", mainString, logger.String("Address", serverAddr))

			return graceful.HTTPTask(server)(ctx)
		},
		graceful.WithTimeout(shutdownTimeout),
		// graceful 以後進先出執行清理，日誌必須最後關閉以記錄其他資源的關閉結果。
		graceful.WithCleanup(func(context.Context) error {
			return loggerCleanup()
		}),
		graceful.WithCleanup(traceShutdown),
		graceful.WithCleanup(service.GetServiceManager().Close),
		graceful.WithCleanup(func(context.Context) error {
			configWatcher.Stop()
			return nil
		}),
		graceful.WithCleanup(server.Shutdown),
	); err != nil {
		logger.Error("HTTP service exited with error", mainString, logger.Err(err))
	}
}

func newHTTPServer(addr string, handler http.Handler) *http.Server {
	return &http.Server{
		Addr:              addr,
		Handler:           handler,
		ReadHeaderTimeout: readHeaderTimeout,
		ReadTimeout:       readTimeout,
		WriteTimeout:      writeTimeout,
		IdleTimeout:       idleTimeout,
		MaxHeaderBytes:    maxHeaderBytes,
	}
}
