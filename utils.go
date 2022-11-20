package confserver

import (
	"context"
	"fmt"
	jsoniter "github.com/json-iterator/go"
	"go.opentelemetry.io/otel/trace"
	"gopkg.in/yaml.v2"
	"time"
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
	return traceIdFromContext(ctx), spanIdFromContext(ctx)
}

func traceIdFromContext(ctx context.Context) string {

	spanCtx := trace.SpanContextFromContext(ctx)
	if spanCtx.HasTraceID() {
		return spanCtx.TraceID().String()
	}
	return ""
}

func spanIdFromContext(ctx context.Context) string {
	spanCtx := trace.SpanContextFromContext(ctx)
	if spanCtx.HasSpanID() {
		return spanCtx.SpanID().String()
	}
	return ""
}

type Password string

func (p Password) String() string {
	return string(p)
}

func (p Password) SecurityString() string {
	var r []rune
	for range []rune(string(p)) {
		r = append(r, []rune("-")...)
	}
	return string(r)
}
