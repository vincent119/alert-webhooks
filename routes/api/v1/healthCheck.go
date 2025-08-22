package v1

import (
	//"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
)

type ResponseHealthCheck struct {
	Status      string `json:"Status"`
	RecvTime    string `json:"recv_time"`
	RecvTimeUTC string `json:"recv_time_utc"`
}

// HealthCheck API
// @Summary Check health status
// @Description Returns OK if the service is healthy
// @Tags health
// @Produce json
// @Success 200 {object} ResponseHealthCheck "Health check response"
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security BasicAuth
// @ID Healthz
// @Router /api/v1/healthz [get]
func HealthCheck(c *gin.Context) {
	c.Header("Content-Type", "application/json")
	response := ResponseHealthCheck{
		Status:      "OK",
		RecvTime:    time.Now().Format("2006-01-02T15:04:05"),
		RecvTimeUTC: time.Now().UTC().Format(time.RFC3339),
	}
	c.JSON(http.StatusOK, response)
}
