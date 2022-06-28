package confserver

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

const (
	corsAcceptHeader                string = "Accept"
	corsAcceptContentLanguageHeader string = "Accept-Language"
	corsContentLanguageHeader       string = "Content-Language"
	corsOriginHeader                string = "Origin"
	corsVaryHeader                  string = "Vary"

	corsAllowOriginHeader      string = "Access-Control-Allow-Origin"
	corsExposeHeadersHeader    string = "Access-Control-Expose-Headers"
	corsMaxAgeHeader           string = "Access-Control-Max-Age"
	corsAllowMethodsHeader     string = "Access-Control-Allow-Methods"
	corsAllowHeadersHeader     string = "Access-Control-Allow-Headers"
	corsAllowCredentialsHeader string = "Access-Control-Allow-Credentials"
	corsRequestMethodHeader    string = "Access-Control-Request-Method"
	corsRequestHeadersHeader   string = "Access-Control-Request-Headers"
)

var (
	defaultCorsMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"}
	defaultCorsHeaders = []string{
		corsAcceptHeader,
		corsAcceptContentLanguageHeader,
		corsContentLanguageHeader,
		corsOriginHeader,
		corsVaryHeader,
		corsAllowOriginHeader,
		corsExposeHeadersHeader,
		corsMaxAgeHeader,
		corsAllowMethodsHeader,
		corsAllowHeadersHeader,
		corsAllowCredentialsHeader,
		corsRequestMethodHeader,
		corsRequestHeadersHeader,
	}
)

func DefaultConfig() cors.Config {
	return cors.Config{
		AllowAllOrigins:  true,
		AllowMethods:     defaultCorsMethods,
		AllowHeaders:     defaultCorsHeaders,
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}
}

func DefaultCORS() gin.HandlerFunc {
	return cors.New(
		DefaultConfig(),
	)
}
