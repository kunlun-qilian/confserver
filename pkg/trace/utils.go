package trace

import (
	"context"
	"net"
	"net/http"
	"strings"

	"go.opentelemetry.io/otel/trace"
	oteltrace "go.opentelemetry.io/otel/trace"
)

func SpanCtxFromContext(ctx context.Context) oteltrace.SpanContext {
	return oteltrace.SpanContextFromContext(ctx)
}

func SpanFromContext(ctx context.Context) oteltrace.Span {
	return oteltrace.SpanFromContext(ctx)
}

// func TraceIDFromContext(ctx context.Context) string {
// 	spanCtx := oteltrace.SpanContextFromContext(ctx)
// 	if spanCtx.HasTraceID() {
// 		return spanCtx.TraceID().String()
// 	}
// 	return ""
// }

// func SpanIDFromContext(ctx context.Context) string {
// 	spanCtx := oteltrace.SpanContextFromContext(ctx)
// 	if spanCtx.HasSpanID() {
// 		return spanCtx.SpanID().String()
// 	}
// 	return ""
// }

func httpRequestRemoteIP(req *http.Request) string {
	if req == nil {
		return ""
	}

	for _, header := range []string{"X-Forwarded-For", "X-Real-IP"} {
		value := strings.TrimSpace(req.Header.Get(header))
		if value == "" {
			continue
		}
		if header == "X-Forwarded-For" {
			parts := strings.Split(value, ",")
			if len(parts) > 0 {
				return strings.TrimSpace(parts[0])
			}
		}
		return value
	}

	host, _, err := net.SplitHostPort(strings.TrimSpace(req.RemoteAddr))
	if err == nil {
		return host
	}
	return strings.TrimSpace(req.RemoteAddr)
}

func TraceAndSpanIDFromContext(ctx context.Context) (string, string) {
	return TraceIDFromContext(ctx), SpanIDFromContext(ctx)
}

func TraceIDFromContext(ctx context.Context) string {
	span := GetTraceSpanFromContext(ctx)
	if span == nil {
		return ""
	}
	spanCtx := trace.SpanContextFromContext(span.Context())
	if spanCtx.HasTraceID() {
		return spanCtx.TraceID().String()
	}
	return ""
}

func SpanIDFromContext(ctx context.Context) string {
	span := GetTraceSpanFromContext(ctx)
	if span == nil {
		return ""
	}
	spanCtx := trace.SpanContextFromContext(span.Context())
	if spanCtx.HasSpanID() {
		return spanCtx.SpanID().String()
	}
	return ""
}
