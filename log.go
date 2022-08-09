package confserver

import (
	"github.com/sirupsen/logrus"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
)

func defaultLogger() *logrus.Logger {
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})
	logger.SetLevel(logrus.DebugLevel)
	logger.Out = os.Stdout
	gin.DefaultWriter = logger.Out
	return logger
}

type LogOption struct {
	LogLevel     string `env:""`
	LogFormatter string `env:""`
}

func SetLogger(opts ...LogOption) gin.HandlerFunc {

	logger := defaultLogger()
	if len(opts) != 0 {
		if opts[0].LogFormatter == "json" {
			logger.SetFormatter(&logrus.JSONFormatter{})
		} else {
			logger.SetFormatter(&logrus.TextFormatter{})
		}
		logLevel, err := logrus.ParseLevel(opts[0].LogLevel)
		if err != nil {
			panic(err)
		}
		logger.SetLevel(logLevel)
	}

	return func(c *gin.Context) {
		startTime := time.Now()
		traceID, spanID := traceAndSpanIDFromContext(c.Request.Context())
		c.Next()

		endTime := time.Now()

		// status
		statusCode := c.Writer.Status()

		entry := logger.WithFields(logrus.Fields{
			"tag":         "access",
			"status":      statusCode,
			"cost":        ReprOfDuration(endTime.Sub(startTime)),
			"remote_ip":   c.ClientIP(),
			"method":      c.Request.Method,
			"request_url": c.Request.URL.String(),
			"user_agent":  c.Request.UserAgent(),
			"refer":       c.Request.Referer(),
			"traceID":     traceID,
			"spanID":      spanID,
		})

		if statusCode >= http.StatusInternalServerError {
			entry.Error()
		} else if statusCode >= http.StatusBadRequest && statusCode < http.StatusInternalServerError {
			entry.Warn()
		} else {
			entry.Info()
		}
	}
}
