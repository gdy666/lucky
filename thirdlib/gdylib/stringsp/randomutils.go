package stringsp

import (
	"math/rand"
	"strings"
	"sync"
	"time"
)

var strModel = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
var stringsBytes = []byte(strModel)
var strModelLength = len(strModel)

var numStrModel = "0123456789"
var numStrBytes = []byte(numStrModel)
var numStrLength = len(numStrModel)

var randVar = rand.New(rand.NewSource(time.Now().UnixNano()))

//GetRandomString 生成随机字符串
func GetRandomString(len int) string {
	var resBuf strings.Builder
	for i := 0; i < len; i++ {
		resBuf.WriteByte(stringsBytes[randVar.Intn(strModelLength)])
	}
	return resBuf.String()
}

//GetRandomStringNum 生成随机数字字符串
func GetRandomStringNum(len int) string {
	var resBuf strings.Builder
	for i := 0; i < len; i++ {
		resBuf.WriteByte(numStrBytes[randVar.Intn(numStrLength)])
	}
	return resBuf.String()

}

var timeStampIDMutex sync.Mutex
var pretimeStampID int64 = 0

//GetTimeStampID 获取时间戳ID
func GetTimeStampID() int64 {
	timeStampIDMutex.Lock()
	defer timeStampIDMutex.Unlock()
	id := time.Now().UnixNano()

CHECK:
	if id == pretimeStampID || id < pretimeStampID {
		if id < pretimeStampID {
			id = pretimeStampID + 1
		} else {
			id++
		}

		goto CHECK
	}

	pretimeStampID = id
	return id
}
