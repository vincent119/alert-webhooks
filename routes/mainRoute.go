// Package routes 提供主路由配置和相關功能
package routes

import (
	"alert-webhooks/config"
	"alert-webhooks/pkg/logger"
	"alert-webhooks/pkg/middleware"
	v1 "alert-webhooks/routes/api/v1"
	"io"
	"net/http"
	"os"
	"strings"

	"alert-webhooks/docs"

	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	ginprometheus "github.com/zsais/go-gin-prometheus"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
)

// 需要跳過日誌的 API 路徑
var skipPathList = []string{
	"/healthz",
	"/healthy",
	"/api/v1/metrics",
	"/",
}

var mainRouteString = "main-route"

// setupSwaggerHost 動態設定 Swagger Host
func setupSwaggerHost() {
	cfg := config.Current()
	// 從請求頭或環境變數動態獲取 host
	host := os.Getenv("SWAGGER_HOST")
	if host == "" {
		// 根據環境設定預設值
		switch strings.ToLower(cfg.App.Mode) {
		case "development", "debug":
			host = "localhost:" + cfg.App.Port
		case "production", "release":
			host = "" // 生產環境讓瀏覽器自動決定
		default:
			host = ""
		}
	}

	// 動態設定 SwaggerInfo
	if host != "" {
		docs.SwaggerInfo.Host = host
		logger.Info("Swagger host set", mainRouteString, logger.String("host", host))
	} else {
		docs.SwaggerInfo.Host = ""
		logger.Info("Swagger host set to auto-detect", mainRouteString)
	}
}

// setupGinMode 設置 Gin 運行模式
func setupGinMode() {
	mode := strings.ToLower(config.Current().App.Mode)

	switch mode {
	case "release", "production":
		gin.SetMode(gin.ReleaseMode)
		logger.Info("Gin mode set to release", mainRouteString)
	case "debug", "development":
		gin.SetMode(gin.DebugMode)
		logger.Info("Gin mode set to debug", mainRouteString)
	case "test":
		gin.SetMode(gin.TestMode)
		logger.Info("Gin mode set to test", mainRouteString)
	default:
		gin.SetMode(gin.ReleaseMode)
		logger.Warn("Unknown mode, defaulting to release", mainRouteString, logger.String("mode", mode))
	}
}

// DefaultRoute 建立並註冊應用程式的 HTTP routes 與 middleware。
func DefaultRoute() *gin.Engine {
	cfg := config.Current()
	setupGinMode()

	// 動態設定 Swagger Host
	setupSwaggerHost()

	// 設定 Gin 輸出
	gin.ForceConsoleColor()
	gin.DefaultWriter = io.MultiWriter(os.Stdout)

	// 創建 Gin 實例
	routes := gin.New()

	// 設定中介件
	routes.Use(
		middleware.RequestBodyLimit(middleware.DefaultRequestBodyLimit),
		otelgin.Middleware(cfg.App.AppName,
			otelgin.WithFilter(func(r *http.Request) bool {
				// 回傳 false 表示不追蹤該請求
				switch r.URL.Path {
				case "/healthz", "/healthy", "/":
					return false
				default:
					return true
				}
			}),
		),
		middleware.CORS(),
		//middleware.RequestID(),
		//middleware.Logger(),
		middleware.LoggerWithSkipPaths(skipPathList),
		//middleware.Recovery(),
		gin.Recovery(),
		gin.LoggerWithConfig(gin.LoggerConfig{SkipPaths: skipPathList}),

		gzip.Gzip(gzip.DefaultCompression, gzip.WithExcludedPaths([]string{
			"/healthy",
			"/api/v1/metrics",
			"/swagger/*any",
			"/swagger/doc.json",
		})),

	// logger.GinLogger()
	)

	// 設置信任代理
	if err := setupTrustedProxies(routes); err != nil {
		logger.Fatal("Failed to set trusted proxies", mainRouteString, logger.Err(err))
	}

	// for load balancer health check
	routes.GET("/healthy", HealthCheck)

	// 設置 Prometheus 監控
	setupPrometheus(routes)

	// // 如果是本地存儲，則配置靜態文件路由
	// if config.Upload.StorageType == "local" {
	// 	uploadPath := config.Upload.BasePath
	// 	staticPathPrefix := "/uploads/"
	// 	routes.Static(staticPathPrefix, uploadPath)
	// 	logger.Info("Local file storage configured",
	// 		logger.String("path", uploadPath),
	// 		logger.String("url_prefix", staticPathPrefix))
	// }

	// Swagger API 文檔
	routes.GET("/swagger/*any", middleware.BasicAuth(), ginSwagger.WrapHandler(swaggerFiles.Handler))

	// 獲取並註冊 API 路由組
	apiGroup := routes.Group("/api")
	{
		v1Router := apiGroup.Group("/v1")
		v1.RegisterApiV1Routes(v1Router)
	}
	return routes
}

// setupTrustedProxies 設置信任代理
func setupTrustedProxies(r *gin.Engine) error {
	cfg := config.Current()
	var trustedProxies []string

	// 從配置中讀取信任代理
	if proxyConfig := cfg.App.TrustedProxies; len(proxyConfig) > 0 {
		trustedProxies = strings.Split(proxyConfig, ",")
	} else {
		// 預設值
		if cfg.App.Mode == "production" {
			trustedProxies = []string{"127.0.0.1", "10.0.0.0/8", "172.16.0.0/12", "192.168.0.0/16"}
		} else {
			trustedProxies = []string{"127.0.0.1"}
		}
	}

	logger.Debug("Setting trusted proxies", mainRouteString, logger.Any("proxies", trustedProxies))
	return r.SetTrustedProxies(trustedProxies)
}

// setupPrometheus 設置 Prometheus 監控
func setupPrometheus(r *gin.Engine) {
	cfg := config.Current()
	p := ginprometheus.NewPrometheus("gin")

	// 設定 Prometheus 指標路徑
	p.MetricsPath = "/api/v1/metrics"

	// 如果配置了用戶名和密碼，則使用身份驗證
	metricUser := cfg.Metric.User
	metricPassword := cfg.Metric.Password

	if metricUser != "" && metricPassword != "" {
		logger.Info("Authentication has been enabled for Prometheus metrics", mainRouteString)
		p.UseWithAuth(r, gin.Accounts{metricUser: metricPassword})
	} else {
		logger.Warn("Authentication is not enabled for Prometheus metrics; it is recommended to enable it in production environments", mainRouteString)

		p.Use(r)
	}

}

// HealthCheck 回傳服務存活狀態與版本。
func HealthCheck(c *gin.Context) {
	cfg := config.Current()
	c.JSON(200, gin.H{
		"status":  "ok",
		"version": cfg.App.Version,
	})
}
