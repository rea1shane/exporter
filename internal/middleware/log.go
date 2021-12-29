package middleware

import (
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"time"
)

func ToStdout(logger *log.Logger) gin.HandlerFunc {

	return func(c *gin.Context) {
		startTime := time.Now()
		c.Next()
		endTime := time.Now()
		latencyTime := endTime.Sub(startTime)

		reqMethod := c.Request.Method
		reqUri := c.Request.RequestURI
		statusCode := c.Writer.Status()
		clientIP := c.ClientIP()

		entry := logger.
			WithField("status_code", statusCode)

		if statusCode == 200 {
			entry.Infof("%15v | %15v | %7v %v", latencyTime, clientIP, reqMethod, reqUri)
		} else if statusCode == 404 {
			entry.Warnf("%15v | %15v | %7v %v", latencyTime, clientIP, reqMethod, reqUri)
		} else {
			entry.Errorf("%15v | %15v | %7v %v", latencyTime, clientIP, reqMethod, reqUri)
		}
	}
}
