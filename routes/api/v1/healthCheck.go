// Package v1 提供 API V1 路由處理功能，包括健康檢查和 API 路由註冊
package v1

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/vincent119/commons/httpx/resp"
)

type ErrorResponse struct {
	Code    int    `json:"code" example:"401"`
	Message string `json:"message" example:"unauthorized"`
}

// HealthCheck API
// @Summary Check health status
// @Description Returns OK if the service is healthy
// @Tags health
// @Produce json
// @Success 200 {object} resp.Health "Health check response"
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security BasicAuth
// @ID Healthz
// @Router /healthz [get]
func HealthCheck(c *gin.Context) {
	c.Header("Content-Type", "application/json")
	response := resp.Health{
		Status: "OK",
	}
	c.JSON(http.StatusOK, response)
}
