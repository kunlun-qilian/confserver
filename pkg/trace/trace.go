package trace

import (
	"context"
	"net/http"
	"strings"
	"time"

	b3prop "go.opentelemetry.io/contrib/propagators/b3"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	oteltrace "go.opentelemetry.io/otel/trace"
)

const (
	defaultOTLPEndpoint = "otel-collector.observability:4318"
	propagatorB3        = "b3"
	propagatorW3C       = "w3c"
)

var ServiceName string

type Trace struct {
	// OTLPEndpoint OTEL Collector / Gateway 地址，留空默认访问本集群的 OTEL Collector
	OTLPEndpoint string `env:""`
	// AlwaysSample default false
	AlwaysSample bool `env:""`
	// 默认启用证书，关闭证书设置为true
	Insecure bool `env:""`
	// AccessToken 访问 OTEL 网关的 access token, 可以为空
	AccessToken string `env:""`
	// Propagator 支持 b3 / w3c, 默认 w3c
	Propagator  string `env:""`
	ServiceName string `env:""`
}

func (c *Trace) SetDefaults() {
	if c.OTLPEndpoint == "" {
		c.OTLPEndpoint = defaultOTLPEndpoint
	}

	if c.Propagator == "" {
		c.Propagator = "w3c"
	}
}

func (c *Trace) Init() {
	c.SetDefaults()

	ServiceName = c.ServiceName

	exporter, err := c.newOTLPExporter()
	if err != nil {
		panic(err)
	}

	var opts []sdktrace.TracerProviderOption
	if c.AlwaysSample {
		opts = append(opts, sdktrace.WithBatcher(exporter), sdktrace.WithSampler(sdktrace.AlwaysSample()),
			sdktrace.WithResource(newResource(ServiceName)))
	} else {
		opts = append(opts, sdktrace.WithBatcher(exporter),
			sdktrace.WithResource(newResource(ServiceName)))
	}

	tp := sdktrace.NewTracerProvider(opts...)
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(c.newPropagator())
}

func (c *Trace) newPropagator() propagation.TextMapPropagator {
	switch strings.ToLower(c.Propagator) {
	case propagatorW3C, "tracecontext":
		return propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{})
	case propagatorB3:
		return b3prop.New()
	default:
		return propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{})
	}
}

func NewSpan(ctx context.Context, span oteltrace.Span) *Span {
	return newSpan(ctx, span)
}

type Span struct {
	ctx  context.Context
	span oteltrace.Span
}

func newSpan(ctx context.Context, span oteltrace.Span) *Span {
	traceSpan := &Span{span: span}
	traceSpan.ctx = context.WithValue(ctx, ContextTraceSpan{}, traceSpan)
	return traceSpan
}

func (c *Span) TraceSpan() oteltrace.Span {
	return c.span
}

func (c *Span) Context() context.Context {
	return c.ctx
}

func (c *Span) End() {
	defer c.span.End(oteltrace.WithTimestamp(time.Now()))
}

func (c *Span) SetHTTPResponseStatus(statusCode int) {
	c.span.SetAttributes(attribute.Int("http.response.status_code", statusCode))
	if statusCode >= http.StatusBadRequest {
		statusText := http.StatusText(statusCode)
		c.span.SetStatus(codes.Error, statusText)
		c.span.AddEvent("@error",
			oteltrace.WithTimestamp(time.Now()),
			oteltrace.WithAttributes(
				attribute.Int("http.response.status_code", statusCode),
			),
		)
	}
}

func Start(ctx context.Context, spanName string, opts ...oteltrace.SpanStartOption) *Span {
	traceCtx, span := otel.Tracer(ServiceName).Start(ctx, spanName, opts...)
	return newSpan(traceCtx, span)
}

func StartHTTPServerSpan(ctx context.Context, req *http.Request, spanName string, opts ...oteltrace.SpanStartOption) *Span {
	if req == nil {
		return Start(ctx, spanName, opts...)
	}

	if spanName == "" {
		spanName = req.URL.Path
		if spanName == "" {
			spanName = req.Method
		}
	}

	traceCtx := ExtractHTTPContext(ctx, req)
	startOpts := []oteltrace.SpanStartOption{
		oteltrace.WithAttributes(httpRequestAttributes(req)...),
	}
	startOpts = append(startOpts, opts...)
	return Start(traceCtx, spanName, startOpts...)
}

func WithTrace(ctx context.Context, req *http.Request) error {
	if req == nil {
		return nil
	}
	otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(req.Header))
	return nil
}

func ExtractHTTPContext(ctx context.Context, req *http.Request) context.Context {
	if req == nil {
		return ctx
	}
	return otel.GetTextMapPropagator().Extract(ctx, propagation.HeaderCarrier(req.Header))
}

func httpRequestAttributes(req *http.Request) []attribute.KeyValue {
	if req == nil {
		return nil
	}
	return []attribute.KeyValue{
		attribute.String("http.request.method", req.Method),
		attribute.String("http.request.path", req.URL.Path),
		attribute.String("http.request.query", req.URL.RawQuery),
		attribute.String("http.request.user_agent", req.UserAgent()),
		attribute.String("http.request.remote_ip", httpRequestRemoteIP(req)),
	}
}

type ContextTraceSpan struct {
}

// var ContextTraceSpanKey = reflect.TypeOf(ContextTraceSpan{}).String()

func GetTraceSpanFromContext(ctx context.Context) *Span {
	span := ctx.Value(ContextTraceSpan{})
	if span == nil {
		return nil
	}
	return span.(*Span)
}
