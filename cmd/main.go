package main

import (
	"alert-webhooks/config"
	"alert-webhooks/pkg/logger"
	"alert-webhooks/pkg/service"
	"alert-webhooks/pkg/watcher"
	"alert-webhooks/routes"
	"context"
	"net/http"
	"os"
	"os/signal"

	//"strings"
	"syscall"
	"time"
	//"github.com/gin-gonic/gin"
)

const (
	shutdownTimeout = 5 * time.Second
)

var mainString = "main"

// waitForShutdownSignal 等待關閉信號並優雅地關閉服務
func waitForShutdownSignal(server *http.Server) {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	logger.Info("Service is shutting down...", mainString)

	// 設置關閉超時
	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Fatal("Service terminated forcefully", mainString, logger.Err(err))
	}

	logger.Info("Service exited", mainString)
}


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
	
	// 初始化服務
	serviceManager := service.GetServiceManager()
	if err := serviceManager.InitServices(); err != nil {
		logger.Warn("Failed to initialize services, some features may not be available", mainString, logger.Err(err))
	}
	
	// 啟動配置檔案監控器
	configWatcher := watcher.NewConfigWatcher()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	
	if err := configWatcher.Start(ctx); err != nil {
		logger.Warn("Failed to start config watcher", mainString, logger.Err(err))
	} else {
		defer configWatcher.Stop()
	}
	
	// 記錄應用啟動信息
	logger.Info("Starting application...", mainString,
		logger.String("mode", config.App.Mode),
		logger.String("version", config.App.Version),
	)

	// 啟動HTTP服務
	startHTTPServer()
}



// startHTTPServer 啟動HTTP服務並處理優雅關閉
func startHTTPServer() {
	port := config.App.Port
	serverAddr := ":" + port
	router := routes.DefaultRoute()

	server := &http.Server{
		Addr:    serverAddr,
		Handler: router,
	}

	// 在goroutine中啟動服務
	go func() {
		logger.Info("Service started", mainString, logger.String("Address", serverAddr))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Error occurred while starting service listener", mainString, logger.Err(err))
		}
	}()

	// 等待中斷信號以優雅關閉服務
	waitForShutdownSignal(server)
}