package common

import (
	"encoding/json"
	"github.com/biggestfatboy/gorder-v2/common/handler/errors"
	"github.com/biggestfatboy/gorder-v2/common/tracing"
	"github.com/gin-gonic/gin"
	"net/http"
)

type BaseResponse struct {
}

type response struct {
	Errno   int    `json:"errno"`
	Message string `json:"message"`
	Data    any    `json:"data"`
	TraceID string `json:"trace_id"`
}

func (base *BaseResponse) Response(c *gin.Context, err error, data interface{}) {
	if err != nil {
		base.error(c, err)
	} else {
		base.success(c, data)
	}
}

func (base *BaseResponse) success(c *gin.Context, data interface{}) {
	errno, errmsg := errors.OutPut(nil)
	r := response{
		Errno:   errno,
		Message: errmsg,
		Data:    data,
		TraceID: tracing.TraceID(c.Request.Context()),
	}
	c.JSON(http.StatusOK, r)
	resp, _ := json.Marshal(r)
	c.Set("response", string(resp))
}

func (base *BaseResponse) error(c *gin.Context, err error) {
	errno, errmsg := errors.OutPut(err)
	r := response{
		Errno:   errno,
		Message: errmsg,
		Data:    nil,
		TraceID: tracing.TraceID(c.Request.Context()),
	}
	c.JSON(http.StatusOK, r)
	resp, _ := json.Marshal(r)
	c.Set("response", string(resp))
}
