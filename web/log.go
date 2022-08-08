package web

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

var maxWebLogsNum = 100

var webLogsStore []WebLog
var webLogs WebLogs
var startTimeStamp string

type WebLogs struct {
	preTimeStamp   int64
	timeStampIndex int64
}

// WebLogs
type WebLog struct {
	Timestamp int64
	Log       string
}

func (wlog *WebLogs) Write(p []byte) (n int, err error) {
	nowTime := time.Now()
	// tripContent := strings.TrimSpace(string(p)[20:])
	// content := fmt.Sprintf("%s %s", nowTime.Format("2006-01-02 15:04:05"), tripContent)

	if webLogs.preTimeStamp == nowTime.UnixNano() {
		webLogs.timeStampIndex++
	} else {
		webLogs.timeStampIndex = 0
	}

	l := WebLog{Timestamp: nowTime.UnixNano() + webLogs.timeStampIndex, Log: strings.TrimSpace(string(p))}

	webLogsStore = append(webLogsStore, l)
	webLogs.preTimeStamp = nowTime.UnixNano()

	if len(webLogsStore) > maxWebLogsNum {
		webLogsStore = webLogsStore[len(webLogsStore)-maxWebLogsNum:]
	}
	return len(p), nil
}

// 初始化日志
func init() {
	log.SetOutput(io.MultiWriter(&webLogs, os.Stdout))
	// log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	var cstZone = time.FixedZone("CST", 8*3600) // 东八
	time.Local = cstZone

	startTimeStamp = fmt.Sprintf("%d", time.Now().UnixNano())

}

// Logs web
func Logs(c *gin.Context) {

	preTimeStampStr := c.Query("pre")
	preTimeStamp, err := strconv.ParseInt(preTimeStampStr, 10, 64)
	if err != nil {
		preTimeStamp = 0
	}

	var logs []interface{}
	for _, l := range webLogsStore {
		if l.Timestamp <= preTimeStamp {
			continue
		}
		m := make(map[string]interface{})
		m["timestamp"] = fmt.Sprintf("%d", l.Timestamp)
		m["log"] = l.Log
		logs = append(logs, m)
	}

	c.JSON(http.StatusOK, gin.H{
		"ret":       0,
		"starttime": startTimeStamp,
		"logs":      logs,
	})

}

// ClearLog
func ClearLog(writer http.ResponseWriter, request *http.Request) {
	webLogsStore = webLogsStore[:0]
}
