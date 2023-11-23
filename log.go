package confserver

import (
	"context"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-courier/logr"
	"github.com/kunlun-qilian/conflogger"
	trace2 "github.com/kunlun-qilian/trace"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/contrib/propagators/b3"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
	"net/http"
	"net/url"
	"time"
)

func TraceHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.FullPath() == "/" || c.FullPath() == "/swagger/*any" {
			return
		}

		startTime := time.Now()
		ctx := c.Request.Context()
		b3Ctx := b3.New().Extract(ctx, propagation.HeaderCarrier(c.Request.Header))
		span := trace2.Start(b3Ctx, c.FullPath(), trace.WithTimestamp(startTime), trace.WithSpanKind(trace.SpanKindServer))
		defer func() {
			span.End()
		}()

		log := conflogger.SpanLogger(span.TraceSpan())
		ctx = logr.WithLogger(ctx, log)

		// inject trace span
		c.Request = c.Request.WithContext(context.WithValue(ctx, trace2.ContextTraceSpanKey, span))
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
			log.WithValues(convertLogFields(entry.Data)...).Error(fmt.Errorf("error status code:%d", statusCode))
		} else if statusCode >= http.StatusBadRequest && statusCode < http.StatusInternalServerError {
			log.WithValues(convertLogFields(entry.Data)...).Warn(fmt.Errorf("warn status code:%d", statusCode))
		} else {
			log.WithValues(convertLogFields(entry.Data)...).Info("")
		}
	}
}

func LoggerHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := time.Now()

		c.Next()

		endTime := time.Now()
		traceID, spanID := traceAndSpanIDFromContext(c.Request.Context())

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

func convertLogFields(fields logrus.Fields) []interface{} {

	var logFields []interface{}
	for k, v := range fields {
		logFields = append(logFields, k)
		logFields = append(logFields, v)
	}
	return logFields
}

func newLoggerResponseWriter(rw http.ResponseWriter) *loggerResponseWriter {
	h, hok := rw.(http.Hijacker)
	if !hok {
		h = nil
	}

	f, fok := rw.(http.Flusher)
	if !fok {
		f = nil
	}

	return &loggerResponseWriter{
		ResponseWriter: rw,
		Hijacker:       h,
		Flusher:        f,
	}
}

type loggerResponseWriter struct {
	http.ResponseWriter
	http.Hijacker
	http.Flusher

	headerWritten bool
	statusCode    int
	err           error
}

func (rw *loggerResponseWriter) WriteError(err error) {
	rw.err = err
}

func (rw *loggerResponseWriter) Header() http.Header {
	return rw.ResponseWriter.Header()
}

func (rw *loggerResponseWriter) WriteHeader(statusCode int) {
	rw.writeHeader(statusCode)
}

func (rw *loggerResponseWriter) Write(data []byte) (int, error) {
	if rw.err == nil && rw.statusCode >= http.StatusBadRequest {
		rw.err = errors.New(string(data))
	}
	return rw.ResponseWriter.Write(data)
}

func (rw *loggerResponseWriter) writeHeader(statusCode int) {
	if !rw.headerWritten {
		rw.ResponseWriter.WriteHeader(statusCode)
		rw.statusCode = statusCode
		rw.headerWritten = true
	}
}

func omitAuthorization(u *url.URL) string {
	query := u.Query()
	query.Del("authorization")
	u.RawQuery = query.Encode()
	return u.String()
}
