package confserver

import (
	"context"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	jsoniter "github.com/json-iterator/go"
	ktrace "github.com/kunlun-qilian/trace"
	"go.opentelemetry.io/otel/trace"
	"gopkg.in/yaml.v2"
)

var j = jsoniter.ConfigCompatibleWithStandardLibrary

func ReprOfDuration(duration time.Duration) string {
	return fmt.Sprintf("%.2fms", float32(duration)/float32(time.Microsecond)/1000)
}

func JSONToYAML(j []byte) ([]byte, error) {
	var jsonObj interface{}
	err := yaml.Unmarshal(j, &jsonObj)
	if err != nil {
		return nil, err
	}
	return yaml.Marshal(jsonObj)
}

func traceAndSpanIDFromContext(ctx context.Context) (string, string) {
	return traceIDFromContext(ctx), spanIDFromContext(ctx)
}

func traceIDFromContext(ctx context.Context) string {
	span := ktrace.GetTraceSpanFromContext(ctx)
	if span == nil {
		return ""
	}
	spanCtx := trace.SpanContextFromContext(span.Context())
	if spanCtx.HasTraceID() {
		return spanCtx.TraceID().String()
	}
	return ""
}

func spanIDFromContext(ctx context.Context) string {
	span := ktrace.GetTraceSpanFromContext(ctx)
	if span == nil {
		return ""
	}
	spanCtx := trace.SpanContextFromContext(span.Context())
	if spanCtx.HasSpanID() {
		return spanCtx.SpanID().String()
	}
	return ""
}

func RequestContextFromGinContext(c *gin.Context) context.Context {
	return c.Request.Context()
}
