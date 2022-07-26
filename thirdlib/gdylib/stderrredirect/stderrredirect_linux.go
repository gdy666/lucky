package stderrredirect

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"syscall"

	"github.com/gdy666/lucky/thirdlib/gdylib/fileutils"
)

var PanicFile *os.File

var doOnce sync.Once

//PanicRedirect panic重定向
func PanicRedirect(fileURL string) {

	doOnce.Do(func() {
		if !strings.HasPrefix(fileURL, "/") { //相对路径
			fileURL = fmt.Sprintf("%s%s%s", fileutils.GetCurrentDirectory(), string(os.PathSeparator), fileURL)
		}

		lastIndex := strings.LastIndex(fileURL, string(os.PathSeparator))

		fileDir := ""

		if lastIndex > 0 {
			fileDir = fileURL[:lastIndex]
		}

		if err := os.MkdirAll(fileDir, 0755); err != nil {
			panic(fmt.Sprintf("创建错误重定向文件夹路径出错:%s", err.Error()))
		}

		file, err := os.OpenFile(fileURL, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		PanicFile = file
		if err != nil {
			return
		}
		if err = syscall.Dup2(int(file.Fd()), int(os.Stderr.Fd())); err != nil {
			return
		}
		return
	})

}
