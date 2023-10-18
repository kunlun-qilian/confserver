package confserver

import (
	"context"
	"fmt"
	"github.com/gin-contrib/gzip"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/gin-contrib/pprof"

	"github.com/gin-gonic/gin"
	"github.com/kunlun-qilian/confx"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.opentelemetry.io/otel"
)

type Server struct {
	Port            int    `env:",opt,expose"`
	Mode            string `env:""`
	HealthCheckPath string `env:",opt,healthCheck"`
	OpenAPISpec     string `env:",opt,copy"`
	UseH2C          bool   `env:""`
	// 跨域
	CorsCheck bool
	// 流式返回 取消压缩
	Stream bool
	r      *gin.Engine
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
	s.SetDefaults()

	tracer := otel.Tracer("Server")
	s.r = gin.New()
	// enable http2
	s.r.UseH2C = s.UseH2C
	// gzip
	// 流式返回 取消压缩
	if !s.Stream {
		s.r.Use(gzip.Gzip(gzip.DefaultCompression))
	}
	// cors
	if s.CorsCheck {
		s.r.Use(AllHeaderCORS())
	} else {
		s.r.Use(DefaultCORS())

	}
	// trace
	s.r.Use(otelgin.Middleware(confx.Config.ServiceName(), otelgin.WithTracerProvider(otel.GetTracerProvider())))
	// log
	s.r.Use(LoggerHandler(tracer))
	// root
	s.r.GET("/", s.RootHandler)
	// openapi
	s.r.GET(fmt.Sprintf("/%s", strings.TrimPrefix(confx.Config.ServiceName(), "srv-")), s.OpenapiHandler)
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
	return s.r.Group(confx.RootPath)
}

func (s *Server) RootHandler(ctx *gin.Context) {
	ctx.Data(200, "text/plain; charset=utf-8", []byte(confx.Config.ServiceName()))
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

	contentByte, err := io.ReadAll(file)
	if err != nil {
		ctx.Data(http.StatusOK, "text/plain; charset=utf-8", nil)
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

	ctx.Data(http.StatusOK, "text/plain; charset=utf-8", openapiByte)
}
