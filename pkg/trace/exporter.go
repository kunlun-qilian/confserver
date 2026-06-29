package trace

import (
	"context"

	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/trace"
)

func (c *Trace) newOTLPExporter() (trace.SpanExporter, error) {
	if c.Insecure {
		return c.newInsecureOTLPExporter()
	}
	return c.newSecureOTLPExporter()
}

func (c *Trace) newInsecureOTLPExporter() (trace.SpanExporter, error) {
	if c.AccessToken == "" {
		return otlptracehttp.New(context.Background(),
			otlptracehttp.WithEndpoint(c.OTLPEndpoint),
			otlptracehttp.WithInsecure(),
		)
	}

	return otlptracehttp.New(context.Background(),
		otlptracehttp.WithEndpoint(c.OTLPEndpoint),
		otlptracehttp.WithInsecure(),
		otlptracehttp.WithHeaders(map[string]string{
			"Authorization": c.AccessToken,
		}),
	)
}

func (c *Trace) newSecureOTLPExporter() (trace.SpanExporter, error) {
	if c.AccessToken == "" {
		return otlptracehttp.New(context.Background(),
			otlptracehttp.WithEndpoint(c.OTLPEndpoint),
		)
	}

	return otlptracehttp.New(context.Background(),
		otlptracehttp.WithEndpoint(c.OTLPEndpoint),
		otlptracehttp.WithHeaders(map[string]string{
			"Authorization": c.AccessToken,
		}),
	)
}
