package jwt

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"qzkvm/src/pkg/setting"
	"qzkvm/src/qzcloud"
	"strings"
)

func JWT() gin.HandlerFunc {
	err:=qzcloud.ConnError
	return func(c *gin.Context) {
		var data interface{}
		if(strings.Trim(c.GetHeader("apikey"),"")!=""&&strings.Trim(c.GetHeader("apikey"),"")==strings.Trim(setting.APIKEY,"")){
			if err!=nil{
				c.JSON(http.StatusUnauthorized, gin.H{
					"code" : 0,
					"msg" :err.Error(),
					"data" :"",
				})

				c.Abort()
				return
			}
			c.Next()
		}else{
			c.JSON(http.StatusUnauthorized, gin.H{
				"code" : 0,
				"msg" :"apikey error",
				"data" : data,
			})
			c.Abort()
			return
		}

	}
}