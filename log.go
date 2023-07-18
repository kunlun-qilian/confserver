package confserver

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/go-courier/logr"
	"github.com/kunlun-qilian/conflogger"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/contrib/propagators/b3"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
	"net/http"
	"net/url"
	"time"
)

func LoggerHandler(tracer trace.Tracer) gin.HandlerFunc {
	return func(c *gin.Context) {

		ctx := c.Request.Context()
		ctx = b3.New().Extract(ctx, propagation.HeaderCarrier(c.Request.Header))

		startTime := time.Now()

		ctx, span := tracer.Start(ctx, c.FullPath(), trace.WithTimestamp(startTime))
		defer func() {
			span.End(trace.WithTimestamp(time.Now()))
		}()

		log := conflogger.SpanLogger(span)
		ctx = logr.WithLogger(ctx, log)
		c.Request = c.Request.WithContext(ctx)

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
