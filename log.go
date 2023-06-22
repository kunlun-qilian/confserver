package confserver

import (
	"github.com/sirupsen/logrus"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func LoggerHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := time.Now()
		traceID, spanID := traceAndSpanIDFromContext(c.Request.Context())
		c.Next()

		endTime := time.Now()

		// status
		statusCode := c.Writer.Status()

		entry := logrus.WithFields(logrus.Fields{
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
