package weblog

import (
	"time"

	"github.com/gdy666/lucky/thirdlib/gdylib/logsbuffer"
)

type LogItem struct {
	ProxyKey   string
	ClientIP   string
	LogContent string
	LogTime    string
}

// 2006-01-02 15:04:05
func WebLogConvert(lg *logsbuffer.LogItem) any {
	l := LogItem{
		LogContent: lg.Content,
		LogTime:    time.Unix(lg.Timestamp/int64(time.Second), 0).Format("2006-01-02 15:04:05")}
	return l
}
