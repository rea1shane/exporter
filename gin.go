package exporter

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func newGinHandler(logger logrus.FieldLogger, notLogged ...string) *gin.Engine {
	engine := gin.New()
	engine.Use(handlerLogger(logger, notLogged...), gin.Recovery())
	return engine
}

// see gin.LoggerWithConfig
func handlerLogger(logger logrus.FieldLogger, notLogged ...string) gin.HandlerFunc {
	var skip map[string]struct{}

	if length := len(notLogged); length > 0 {
		skip = make(map[string]struct{}, length)

		for _, path := range notLogged {
			skip[path] = struct{}{}
		}
	}

	return func(c *gin.Context) {
		// Start timer
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// Process request
		c.Next()

		// Log only when path is not being skipped
		if _, ok := skip[path]; !ok {
			param := gin.LogFormatterParams{
				Request: c.Request,
				Keys:    c.Keys,
			}

			// Stop timer
			param.TimeStamp = time.Now()
			param.Latency = param.TimeStamp.Sub(start)

			param.ClientIP = c.ClientIP()
			param.Method = c.Request.Method
			param.StatusCode = c.Writer.Status()
			param.ErrorMessage = c.Errors.ByType(gin.ErrorTypePrivate).String()

			param.BodySize = c.Writer.Size()

			if raw != "" {
				path = path + "?" + raw
			}

			param.Path = path

			log(logger, param)
		}
	}
}

func log(logger logrus.FieldLogger, param gin.LogFormatterParams) {
	if param.Latency > time.Minute {
		param.Latency = param.Latency.Truncate(time.Second)
	}

	entry := logger.WithFields(logrus.Fields{
		"StatusCode": fmt.Sprintf("%3d", param.StatusCode),
		"Latency":    fmt.Sprintf("%13v", param.Latency),
	})

	msg := fmt.Sprintf("%15s | %-7s %#v",
		param.ClientIP,
		param.Method,
		param.Path,
	)
	if len(param.ErrorMessage) > 0 {
		msg = fmt.Sprintf("%s\n%s",
			msg,
			param.ErrorMessage[:len(param.ErrorMessage)-1],
		)
	}

	if param.StatusCode >= http.StatusInternalServerError {
		entry.Error(msg)
	} else if param.StatusCode >= http.StatusBadRequest {
		entry.Warn(msg)
	} else {
		entry.Info(msg)
	}
}
