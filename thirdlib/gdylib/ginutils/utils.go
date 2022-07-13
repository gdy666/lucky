package ginutils

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
)

//Cors 处理跨域请求,支持options访问
func Cors(params ...interface{}) gin.HandlerFunc {
	return func(c *gin.Context) {
		method := c.Request.Method

		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Headers", "Content-Type,AccessToken,X-CSRF-Token, Authorization, Token,access_token")
		c.Header("Access-Control-Allow-Methods", "POST, GET, PUT,OPTIONS")
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

func GetJWTToken(tokenString, tokenKey string) (t *jwt.Token, e error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(tokenKey), nil
	})
	if err != nil {
		//beego.Error("Parse token:", err)
		if ve, ok := err.(*jwt.ValidationError); ok {
			if ve.Errors&jwt.ValidationErrorMalformed != 0 {
				// That's not even a token
				return nil, errors.New("errInputData")
			} else if ve.Errors&(jwt.ValidationErrorExpired|jwt.ValidationErrorNotValidYet) != 0 {
				// Token is either expired or not active yet
				return nil, errors.New("errExpired")
			} else {
				// Couldn't handle this token
				return nil, errors.New("errInputData")
			}
		} else {
			// Couldn't handle this token
			return nil, errors.New("errInputData")
		}
	}
	if !token.Valid {
		//beego.Error("Token invalid:", tokenString)
		return nil, errors.New("errInputData")
	}

	return token, nil
}

// info 存储的信息
// key 加密的key
// exp 有效期
//GetJWTTokenString 获取Token字符串
func GetJWTTokenString(info map[string]interface{}, key string, exp time.Duration) (string, error) {
	claims := make(jwt.MapClaims)
	claims["exp"] = time.Now().Add(exp).Unix() //token 24小时有效期
	for k := range info {
		claims[k] = info[k]
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(key))
	if err != nil {
		return "", fmt.Errorf("生成TokenString出错:%s", err.Error())
	}

	return tokenString, nil
}

//GetChildDomain 获取子域名部分
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
