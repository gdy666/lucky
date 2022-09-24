package ginutils

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// Cors 处理跨域请求,支持options访问
func Cors(params ...interface{}) gin.HandlerFunc {
	return func(c *gin.Context) {
		method := c.Request.Method

		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Headers", "Content-Type,AccessToken,X-CSRF-Token, Authorization, Token,access_token")
		c.Header("Access-Control-Allow-Methods", "POST, GET, PUT,OPTIONS,DELETE")
		c.Header("Access-Control-Expose-Headers", "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers, Content-Type")
		c.Header("Access-Control-Allow-Credentials", "true")

		//放行所有OPTIONS方法
		if method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
		}
		// 处理请求
		c.Next()
	}
}

// GetChildDomain 获取子域名部分
func GetChildDomain(host string) string {
	hostSplitList := strings.Split(host, ".")
	listLen := len(hostSplitList)
	var resBuilder strings.Builder

	for i := range hostSplitList {
		if i >= listLen-2 {
			break
		}
		if resBuilder.Len() > 0 {
			resBuilder.WriteString(".")
		}
		resBuilder.WriteString(hostSplitList[i])
	}

	return resBuilder.String()
}
