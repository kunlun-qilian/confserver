package confserver

import (
	"context"
	"fmt"
	"github.com/gin-contrib/pprof"
	"io/ioutil"
	"os"
	"strings"
	"sync"

	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	b3prop "go.opentelemetry.io/contrib/propagators/b3"
	"go.opentelemetry.io/otel"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

type Server struct {
	Port            int        `env:",opt,expose"`
	LogOption       *LogOption `env:""`
	Mode            string     `env:""`
	HealthCheckPath string     `env:",opt,healthCheck"`
	OpenAPISpec     string     `env:",opt,copy"`

	r *gin.Engine

	// healthCheckUpdated
	healthCheckUpdated bool
}

func (s *Server) SetDefaults() {
	if s.Port == 0 {
		s.Port = 80
	}

	if s.Mode == "" {
		s.Mode = "release"
	}

	if s.LogOption.LogFormatter == "" {
		s.LogOption.LogFormatter = "json"
	}
	if s.LogOption.LogLevel == "" {
		s.LogOption.LogLevel = "debug"
	}

	gin.SetMode(s.Mode)

	if s.OpenAPISpec == "" {
		s.OpenAPISpec = "./openapi.json"
	}

	if !s.healthCheckUpdated {
		if s.HealthCheckPath == "" {
			s.HealthCheckPath = fmt.Sprintf("http://:%d/", s.Port)
		} else {
			s.HealthCheckPath = fmt.Sprintf("http://:%d%s", s.Port, s.HealthCheckPath)
		}
		s.healthCheckUpdated = true
	}
}

func (s *Server) Init() {

	otp := otel.GetTracerProvider()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSampler(sdktrace.AlwaysSample()))
	otel.SetTextMapPropagator(b3prop.New())
	otel.SetTracerProvider(tp)

	s.SetDefaults()
	s.SetLogger()

	s.r = gin.New()
	// gzip
	s.r.Use(gzip.Gzip(gzip.DefaultCompression))
	// cors
	s.r.Use(DefaultCORS())
	// trace
	s.r.Use(otelgin.Middleware(config.ServiceName(), otelgin.WithTracerProvider(otp)))
	// log
	s.r.Use(LoggerHandler())
	// root
	s.r.GET("/", s.RootHandler)
	// openapi
	s.r.GET(fmt.Sprintf("/%s", strings.TrimPrefix(config.ServiceName(), "srv-")), s.OpenapiHandler)
	if strings.ToLower(s.Mode) == "debug" {
		s.r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
		pprof.Register(s.r)
	}
}

func (s *Server) Engine() *gin.Engine {
	return s.r
}

func (s *Server) serve(ctx context.Context) error {
	return s.r.Run(fmt.Sprintf(":%d", s.Port))

}

func (s *Server) Serve(ctx context.Context, fn ...func(ctx context.Context) error) (err error) {
	wg := &sync.WaitGroup{}
	serverList := []func(ctx context.Context) error{
		s.serve,
	}

	if len(fn) != 0 {
		serverList = append(serverList, fn...)
	}

	for i := range serverList {
		wg.Add(1)
		go func(s func(ctx context.Context) error) {
			defer wg.Done()

			if e := s(ctx); e != nil {
				err = e
			}
		}(serverList[i])
	}
	wg.Wait()
	return
}

func (s *Server) SvcRootRouter() *gin.RouterGroup {
	return s.r.Group(svcRouterPath)
}

func (s *Server) RootHandler(ctx *gin.Context) {
	ctx.Data(200, "text/plain; charset=utf-8", []byte(config.ServiceName()))
}

func (s *Server) OpenapiHandler(ctx *gin.Context) {

	file, err := os.Open(s.OpenAPISpec)
	defer func() {
		_ = file.Close()
	}()
	if err != nil {
		ctx.Data(200, "text/plain; charset=utf-8", nil)
		return
	}

	var openapiByte []byte

	contentByte, err := ioutil.ReadAll(file)
	if err != nil {
		ctx.Data(200, "text/plain; charset=utf-8", nil)
		return
	}
	openapiByte = contentByte

	q, exist := ctx.Get("format")
	// yaml format
	if exist && q.(string) == "yaml" {
		yamlByte, err := JSONToYAML(contentByte)
		if err != nil {
			ctx.Data(200, "text/plain; charset=utf-8", nil)
			return
		}
		openapiByte = yamlByte
	}

	ctx.Data(200, "text/plain; charset=utf-8", openapiByte)
}
