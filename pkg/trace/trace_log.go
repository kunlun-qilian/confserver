package trace

import (
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	oteltrace "go.opentelemetry.io/otel/trace"
)

func (c *Span) Info(msg interface{}, format ...string) {
	message := formatSpanMessage(msg, format...)
	c.logrusEntry().Info(message)
	c.span.AddEvent("@info",
		oteltrace.WithTimestamp(time.Now()),
		oteltrace.WithAttributes(
			attribute.String("msg", message),
		))
}

func (c *Span) Warn(msg interface{}, format ...string) {
	message := formatSpanMessage(msg, format...)
	c.logrusEntry().Warn(message)
	c.span.AddEvent("@warn",
		oteltrace.WithTimestamp(time.Now()),
		oteltrace.WithAttributes(
			attribute.String("msg", message),
		),
	)
}

func (c *Span) Error(msg interface{}, format ...string) {
	message := formatSpanMessage(msg, format...)
	c.logrusEntry().Error(message)

	c.span.SetStatus(codes.Error, message)
	c.span.AddEvent("@error",
		oteltrace.WithTimestamp(time.Now()),
		oteltrace.WithAttributes(
			attribute.String("msg", message),
		),
	)
}

func (c *Span) Debug(msg interface{}, format ...string) {
	message := formatSpanMessage(msg, format...)
	c.logrusEntry().Debug(message)
	c.span.AddEvent("@debug",
		oteltrace.WithTimestamp(time.Now()),
		oteltrace.WithAttributes(
			attribute.String("msg", message)),
	)
}

func (c *Span) logrusEntry() *logrus.Entry {
	fields := logrus.Fields{}
	if traceID := TraceIDFromContext(c.ctx); traceID != "" {
		fields["trace_id"] = traceID
	}
	if spanID := SpanIDFromContext(c.ctx); spanID != "" {
		fields["span_id"] = spanID
	}
	if len(fields) == 0 {
		return logrus.NewEntry(logrus.StandardLogger())
	}
	return logrus.WithFields(fields)
}

func formatSpanMessage(msg interface{}, format ...string) string {
	if len(format) > 0 {
		return fmt.Sprintf(format[0], msg)
	}
	return fmt.Sprintf("%v", msg)
}
