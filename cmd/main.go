package main

import (
	"alert-webhooks/config"
	"alert-webhooks/pkg/logger"
	"alert-webhooks/pkg/service"
	"alert-webhooks/pkg/trace"
	"alert-webhooks/pkg/watcher"
	"alert-webhooks/routes"
	"context"
	"net/http"
	"time"

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
	config.Init()

	// 初始化日誌系統
	logger.InitLogger(config.Log.Level, config.IsDevelopment())

	// 初始化 OpenTelemetry TracerProvider
	traceShutdown, err := trace.InitTracerProvider(
		context.Background(),
		config.Trace,
		config.App.AppName,
	)
	if err != nil {
		logger.Fatal("Failed to initialize TracerProvider", mainString, logger.Err(err))
	}

	// 初始化服務
	serviceManager := service.GetServiceManager()
	if err := serviceManager.InitServices(); err != nil {
		logger.Warn("Failed to initialize services, some features may not be available", mainString, logger.Err(err))
	}

	startHTTPServer(traceShutdown)
}

// startHTTPServer 啟動HTTP服務並處理優雅關閉
func startHTTPServer(traceShutdown func(context.Context) error) {
	port := config.App.Port
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
				logger.String("mode", config.App.Mode),
				logger.String("version", config.App.Version),
			)
			logger.Info("Service started", mainString, logger.String("Address", serverAddr))

			return graceful.HTTPTask(server)(ctx)
		},
		graceful.WithTimeout(shutdownTimeout),
		graceful.WithCleanup(traceShutdown),
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
