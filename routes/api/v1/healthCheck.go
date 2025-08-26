package v1

import (
	//"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/vincent119/commons/modelx"
	//"time"
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
// @Failure 401 {object} modelx.ErrorResponse
// @Failure 500 {object} modelx.ErrorResponse
// @Security BasicAuth
// @ID Healthz
// @Router /healthz [get]
func HealthCheck(c *gin.Context) {
	c.Header("Content-Type", "application/json")
	response := modelx.ResponseHealthCheck{
		Status:      "OK",
		// RecvTime:    time.Now().Format("2006-01-02T15:04:05"),
		// RecvTimeUTC: time.Now().UTC().Format(time.RFC3339),
	}
	c.JSON(http.StatusOK, response)
}
