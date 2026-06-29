package confserver

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-courier/logr"
	"github.com/kunlun-qilian/conflogger"
	trace2 "github.com/kunlun-qilian/confserver/pkg/trace"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

func TraceHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.FullPath() == "/swagger/*any" || c.FullPath() == "/healthz" || c.FullPath() == "/favicon.ico" {
			c.Next()
			return
		}

		startTime := time.Now()
		ctx := c.Request.Context()

		propagator := otel.GetTextMapPropagator()
		traceCtx := propagator.Extract(ctx, propagation.HeaderCarrier(c.Request.Header))
		span := trace2.StartHTTPServerSpan(traceCtx, c.Request, c.FullPath(), trace.WithTimestamp(startTime), trace.WithSpanKind(trace.SpanKindServer))
		defer func() {
			span.End()
		}()

		ctx = logr.WithLogger(ctx, conflogger.StdLogger())

		// inject trace span
		c.Request = c.Request.WithContext(context.WithValue(ctx, trace2.ContextTraceSpan{}, span))
		c.Next()
		span.SetHTTPResponseStatus(c.Writer.Status())

	}
}

func LoggerHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := time.Now()

		c.Next()

		endTime := time.Now()
		traceID, spanID := trace2.TraceAndSpanIDFromContext(c.Request.Context())

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
			"trace_id":    traceID,
			"span_id":     spanID,
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
