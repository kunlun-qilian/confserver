package confserver

import (
	"github.com/gin-gonic/gin"
	"github.com/go-courier/httptransport"
	"github.com/go-courier/httptransport/httpx"
	"github.com/go-courier/httptransport/transformers"
	"github.com/go-courier/x/context"
	"github.com/julienschmidt/httprouter"
	"net/http"
	"reflect"
)

var rtMgr = httptransport.NewRequestTransformerMgr(nil, nil)

func init() {
	rtMgr.SetDefaults()
}

func Bind(ctx *gin.Context, obj interface{}) error {

	req, err := rtMgr.NewRequestTransformer(ctx, reflect.TypeOf(obj))
	if err != nil {
		return err
	}
	err = req.DecodeAndValidate(ctx, httpx.NewRequestInfo(contextWithPathParams(ctx)), obj)
	if err != nil {
		return err
	}
	return nil
}

func contextWithPathParams(ctx *gin.Context) *http.Request {
	params, _ := transformers.NewPathnamePattern(ctx.FullPath()).Parse(ctx.Request.URL.Path)
	return ctx.Request.WithContext(context.WithValue(ctx.Request.Context(), httprouter.ParamsKey, params))
}
