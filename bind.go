package confserver

import (
    "github.com/gin-gonic/gin"
    "github.com/go-courier/httptransport"
    "github.com/go-courier/httptransport/httpx"
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
    err = req.DecodeAndValidate(ctx, httpx.NewRequestInfo(ctx.Request), obj)
    if err != nil {
        return err
    }
    return nil
}
