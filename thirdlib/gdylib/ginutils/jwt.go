package ginutils

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt"
)

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
// GetJWTTokenString 获取Token字符串
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
