package app

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

type Gin struct {
	C *gin.Context
}

type Response struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

// Response setting gin.JSON
func (g *Gin) Response(errCode int, msg string,data interface{}) {
	g.C.JSON(http.StatusOK, Response{
		Code: errCode,
		Msg: msg,
		Data: data,
	})
	return
}