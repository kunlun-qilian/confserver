package confserver

import (
	"context"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	jsoniter "github.com/json-iterator/go"
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

func RequestContextFromGinContext(c *gin.Context) context.Context {
	return c.Request.Context()
}
