package stderrredirect

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"syscall"

	"github.com/gdy666/lucky/thirdlib/gdylib/fileutils"
)

//错误输出重定向,用于捕获闪退信息

//PanicFile 记录闪退的文件
var PanicFile *os.File

var doOnce sync.Once

//PanicRedirect panic重定向
func PanicRedirect(fileURL string) {

	doOnce.Do(func() {
		if !strings.Contains(fileURL, ":") { //相对路径
			fileURL = fmt.Sprintf("%s%s%s", fileutils.GetCurrentDirectory(), string(os.PathSeparator), fileURL)
		}

		//fmt.Printf("FileURL:%s\n", fileURL)

		lastIndex := strings.LastIndex(fileURL, string(os.PathSeparator))

		fileDir := ""

		if lastIndex > 0 {
			fileDir = fileURL[:lastIndex]
		}

		if err := os.MkdirAll(fileDir, 0755); err != nil {
			panic(fmt.Sprintf("创建错误重定向文件夹路径出错:%s", err.Error()))
		}

		//fileDir := strings.LastIndex(fileURL, string(os.PathSeparator))

		file, err := os.OpenFile(fileURL, os.O_CREATE|os.O_APPEND, 0666)
		PanicFile = file
		if err != nil {
			panic(fmt.Sprintf("panic重定向出错:%s", err.Error()))
		}
		kernel32 := syscall.NewLazyDLL("kernel32.dll")
		setStdHandle := kernel32.NewProc("SetStdHandle")
		sh := syscall.STD_ERROR_HANDLE
		v, _, err := setStdHandle.Call(uintptr(sh), uintptr(file.Fd()))
		if v == 0 {
			return
		}
	})

}
