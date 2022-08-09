package confserver

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	b3prop "go.opentelemetry.io/contrib/propagators/b3"
	"go.opentelemetry.io/otel"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

type Server struct {
	Port int `env:",opt,expose"`
	LogOption
	Mode            string `env:""`
	HealthCheckPath string `env:",opt,healthCheck"`
	OpenAPISpec     string `env:",opt,copy"`

	r *gin.Engine

	// healthCheckUpdated
	healthCheckUpdated bool
}

func (s *Server) SetDefaults() {
	if s.Port == 0 {
		s.Port = 80
	}

	if s.Mode == "" {
		s.Mode = "debug"
	}

	s.LogFormatter = "json"
	s.LogLevel = "debug"

	if s.LogFormatter == "json" {
		logrus.SetFormatter(&logrus.JSONFormatter{})
	} else {
		logrus.SetFormatter(&logrus.TextFormatter{})
	}
	logrus.SetReportCaller(true)
	logLevel, err := logrus.ParseLevel(s.LogLevel)
	if err != nil {
		panic(err)
	}
	logrus.SetLevel(logLevel)

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

	s.r = gin.New()

}

func (s *Server) Init() {

	otp := otel.GetTracerProvider()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSampler(sdktrace.AlwaysSample()))
	otel.SetTextMapPropagator(b3prop.New())
	otel.SetTracerProvider(tp)

	s.SetDefaults()
	// gzip
	s.r.Use(gzip.Gzip(gzip.DefaultCompression))
	// cors
	s.r.Use(DefaultCORS())
	// trace
	s.r.Use(otelgin.Middleware(config.ServiceName(), otelgin.WithTracerProvider(otp)))
	// log
	s.r.Use(SetLogger())
	// root
	s.r.GET("/", s.RootHandler)
	// openapi
	s.r.GET(fmt.Sprintf("/%s", strings.TrimPrefix(config.ServiceName(), "srv-")), s.OpenapiHandler)
	if strings.ToLower(s.Mode) == "debug" {
		s.r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	}
}

func (s *Server) Engine() *gin.Engine {
	return s.r
}

func (s *Server) Serve() {
	err := s.r.Run(fmt.Sprintf(":%d", s.Port))
	if err != nil {
		panic(err)
	}
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
