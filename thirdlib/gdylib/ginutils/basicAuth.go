package ginutils

import (
	"crypto/subtle"
	"encoding/base64"
	"unsafe"

	"github.com/gin-gonic/gin"
)

type BasicAuthPair struct {
	value string
	user  string
}

type BasicAuthPairs []BasicAuthPair

func ProcessAccounts(accounts gin.Accounts) BasicAuthPairs {
	length := len(accounts)
	assert1(length > 0, "Empty list of authorized credentials")
	pairs := make(BasicAuthPairs, 0, length)
	for user, password := range accounts {
		assert1(user != "", "User can not be empty")
		value := authorizationHeader(user, password)
		pairs = append(pairs, BasicAuthPair{
			value: value,
			user:  user,
		})
	}
	return pairs
}

func authorizationHeader(user, password string) string {
	base := user + ":" + password
	return "Basic " + base64.StdEncoding.EncodeToString(StringToBytes(base))
}

func assert1(guard bool, text string) {
	if !guard {
		panic(text)
	}
}

func StringToBytes(s string) []byte {
	return *(*[]byte)(unsafe.Pointer(
		&struct {
			string
			Cap int
		}{s, len(s)},
	))
}

func (a BasicAuthPairs) SearchCredential(authValue string) (string, bool) {
	if authValue == "" {
		return "", false
	}
	for _, pair := range a {
		if subtle.ConstantTimeCompare(StringToBytes(pair.value), StringToBytes(authValue)) == 1 {
			return pair.user, true
		}
	}
	return "", false
}
