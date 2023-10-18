package confserver

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
)

const (
	corsAcceptHeader                string = "Accept"
	corsAcceptContentLanguageHeader string = "Accept-Language"
	corsContentLanguageHeader       string = "Content-Language"
	corsContentLengthHeader         string = "Content-Length"
	corsOriginHeader                string = "Origin"
	corsVaryHeader                  string = "Vary"
	corsAuthorizationHeader         string = "Authorization"

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
		corsContentLengthHeader,
		corsAcceptContentLanguageHeader,
		corsContentLanguageHeader,
		corsOriginHeader,
		corsVaryHeader,
		corsAuthorizationHeader,
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

func DefaultCORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", strings.Join(defaultCorsMethods, ","))
		c.Header("Access-Control-Allow-Headers", strings.Join(defaultCorsHeaders, ","))
		c.Header("Access-Control-Expose-Headers", "Content-Length")
		c.Header("Access-Control-Allow-Credentials", "true")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}

func AllowAllCors() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "*")
		c.Header("Access-Control-Allow-Headers", "*")
		c.Header("Access-Control-Expose-Headers", "*")
		c.Header("Access-Control-Allow-Credentials", "true")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}
