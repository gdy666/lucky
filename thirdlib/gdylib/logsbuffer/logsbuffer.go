package logsbuffer

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

type LogsBuffer struct {
	bufferSize     int
	preTimeStamp   int64
	timeStampIndex int64
	logsStore      []LogItem
	mu             sync.RWMutex
	fireCallback   func(entry *logrus.Entry) error
}

type LogItem struct {
	Timestamp int64
	Content   string
	Data      map[string]any
}

func (l *LogsBuffer) AddLog(t time.Time, msg string, data map[string]any) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.preTimeStamp == t.UnixNano() {
		l.timeStampIndex++
	} else {
		l.timeStampIndex = 0
	}

	li := LogItem{Timestamp: t.UnixNano() + l.timeStampIndex, Content: strings.TrimSpace(msg), Data: data}
	l.logsStore = append(l.logsStore, li)
	l.preTimeStamp = t.UnixNano()
	if len(l.logsStore) > l.bufferSize+16 {
		l.logsStore = l.logsStore[len(l.logsStore)-l.bufferSize:]
	}
}

func (l *LogsBuffer) Fire(entry *logrus.Entry) error {

	entryStr, err := entry.String()
	if err != nil {
		return fmt.Errorf("entry.String() err:%s", err.Error())
	}

	if l.fireCallback != nil {
		return l.fireCallback(entry)
	} else {
		l.AddLog(entry.Time, entryStr, entry.Data)
	}

	return nil
}

func (l *LogsBuffer) SetFireCallback(f func(entry *logrus.Entry) error) {
	l.fireCallback = f
}

func (l *LogsBuffer) Levels() []logrus.Level {
	return logrus.AllLevels
}

func (l *LogsBuffer) Write(p []byte) (n int, err error) {

	l.AddLog(time.Now(), string(p), nil)

	return len(p), nil
}

func (l *LogsBuffer) GetLogs(logItemConvertFunc func(*LogItem) any, fTimestamp int64) []any {
	l.mu.RLock()
	defer l.mu.RUnlock()
	var logs []any
	for i := range l.logsStore {
		if l.logsStore[i].Timestamp <= fTimestamp {
			continue
		}
		lg := l.getLogItem(logItemConvertFunc, l.logsStore[i])

		logs = append(logs, lg)
	}
	return logs
}

func (l *LogsBuffer) GetLastLogs(logItemConvertFunc func(*LogItem) any, maxCount int) []any {
	l.mu.RLock()
	defer l.mu.RUnlock()
	logCount := len(l.logsStore)
	var resRaw []LogItem
	if maxCount >= logCount {
		resRaw = l.logsStore[0:]
	} else {
		resRaw = l.logsStore[logCount-maxCount:]
	}
	var logs []any
	for i := range resRaw {
		lg := l.getLogItem(logItemConvertFunc, resRaw[i])
		logs = append(logs, lg)
	}
	return logs
}

func (l *LogsBuffer) GetLogsByLimit(logItemConvertFunc func(*LogItem) any, pageSize, page int) (int, []any) {
	l.mu.RLock()
	defer l.mu.RUnlock()
	logCount := len(l.logsStore)
	var resRaw []LogItem

	firstIndex := (page - 1) * pageSize
	endIndex := firstIndex + pageSize

	if firstIndex < logCount {
		if endIndex >= logCount {
			resRaw = l.logsStore[firstIndex:]
		} else {
			resRaw = l.logsStore[firstIndex:endIndex]
		}
	}

	var logs []any
	for i := range resRaw {
		lg := l.getLogItem(logItemConvertFunc, resRaw[i])
		logs = append(logs, lg)
	}
	return logCount, logs
}

func (l *LogsBuffer) getLogItem(logItemConvertFunc func(*LogItem) any, li LogItem) any {
	var lg any
	if logItemConvertFunc == nil {
		lg = li
	} else {
		lg = logItemConvertFunc(&li)
	}
	return lg
}

func (l *LogsBuffer) ClearLog() {
	l.logsStore = l.logsStore[:0]
}

func Create(size int) *LogsBuffer {
	lb := &LogsBuffer{bufferSize: size}
	return lb
}

func (l *LogsBuffer) SetBufferSize(size int) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.bufferSize = size
	if len(l.logsStore) > l.bufferSize {
		l.logsStore = l.logsStore[len(l.logsStore)-l.bufferSize:]
	}
}

func (l *LogsBuffer) GetBufferSize() int {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.bufferSize
}

func (l *LogsBuffer) GetLogCount() int {
	l.mu.Lock()
	defer l.mu.Unlock()
	return len(l.logsStore)
}

//---------------------------

var LogsBufferStore map[string]*LogsBuffer
var LogsBufferStoreMu sync.Mutex

func init() {
	LogsBufferStore = make(map[string]*LogsBuffer)
}

func CreateLogbuffer(key string, buffSize int) *LogsBuffer {
	if strings.TrimSpace(key) == "" {
		return nil
	}
	LogsBufferStoreMu.Lock()
	defer LogsBufferStoreMu.Unlock()
	var buf *LogsBuffer
	var ok bool
	if buf, ok = LogsBufferStore[key]; !ok {
		buf = &LogsBuffer{}
		buf.SetBufferSize(buffSize)
		LogsBufferStore[key] = buf
	} else if buf.GetBufferSize() != buffSize {
		buf.SetBufferSize(buffSize)
	}
	return buf
}
