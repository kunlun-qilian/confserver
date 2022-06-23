package confserver

import (
	"fmt"
	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

type Server struct {
	Port int `env:""`
	LogOption
	Mode string `env:""`
	r    *gin.Engine
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
	s.r = gin.New()

}

func (s *Server) Init() {
	s.SetDefaults()

	s.r.Use(gzip.Gzip(gzip.DefaultCompression))

	s.r.Use(SetLogger())

	s.r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
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
