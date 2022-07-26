package recoverutil

import (
	"fmt"
	"os"
	"runtime/debug"
	"strings"
	"time"

	"github.com/gdy666/lucky/thirdlib/gdylib/fileutils"
	"github.com/gdy666/lucky/thirdlib/gdylib/stderrredirect"
)

//RecoverHandler 恢复处理
func RecoverHandler(recoverErr interface{}, exit, reboot bool, panicFileURL string) {
	if recoverErr == nil {
		return
	}

	outputPanicV2(panicFileURL, recoverErr)

	if reboot {

		var argvsBuilder strings.Builder

		for i := range os.Args {
			if i == 0 {
				continue
			}
			if argvsBuilder.Len() == 0 {
				argvsBuilder.WriteString(os.Args[i])
			} else {
				argvsBuilder.WriteString(" ")
				argvsBuilder.WriteString(os.Args[i])
			}
		}

		fileutils.OpenProgramOrFile(os.Args) //重启程序
		//fileutil.OpenProgramOrFile(restartURIBuilder.String())
	}
	if exit {
		os.Exit(1)
	}
}

func outputPanic(panicFileURL string, recoverErr interface{}) {
	exeName := os.Args[0] //获取程序名称
	now := time.Now()     //获取当前时间
	pid := os.Getpid()    //获取进程ID

	if !strings.Contains(panicFileURL, ":") && !strings.HasPrefix(panicFileURL, "/") { //相对路径
		panicFileURL = fmt.Sprintf("%s%s%s", fileutils.GetCurrentDirectory(), string(os.PathSeparator), panicFileURL)
	}

	fileDir := ""
	lastIndex := strings.LastIndex(panicFileURL, string(os.PathSeparator))
	if lastIndex > 0 {
		fileDir = panicFileURL[:lastIndex]
	}

	if err := os.MkdirAll(fileDir, 0755); err != nil {
		panic(fmt.Sprintf("创建错误重定向文件夹路径出错:%s", err.Error()))
	}

	file, err := os.OpenFile(panicFileURL, os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		panic(fmt.Sprintf("panic重定向出错:%s", err.Error()))
	}

	defer file.Close()

	timeStr := now.Format("2006-01-02 15:04:05") //设定时间格式
	file.WriteString("\n\n\n↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓\r\n")
	file.WriteString(fmt.Sprintf("%s-%d-%s dump LOG\r\n", exeName, pid, timeStr))
	file.WriteString(fmt.Sprintf("%v\r\n", err)) //输出panic信息
	file.WriteString(string(debug.Stack()))      //输出堆栈信息
	file.WriteString("↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑\r\n")

	// timeStr := now.Format("20060102150405")                          //设定时间格式
	// fname := fmt.Sprintf("%s-%d-%s-dump.log", exeName, pid, timeStr) //保存错误信息文件名:程序名-进程ID-当前时间（年月日时分秒）
	// fmt.Println("dump to file ", fname)

	// f, err := os.Create(fname)
	// if err != nil {
	// 	return
	// }
	// defer f.Close()

	// if recoverErr != nil {
	// 	f.WriteString(fmt.Sprintf("%v\r\n", err)) //输出panic信息
	// 	f.WriteString("========\r\n")
	// }

	// f.WriteString(string(debug.Stack())) //输出堆栈信息
}

func outputPanicV2(panicFileURL string, recoverErr interface{}) {
	exeName := os.Args[0] //获取程序名称
	now := time.Now()     //获取当前时间
	pid := os.Getpid()    //获取进程ID
	setPanicRedirect := true
	if panicFileURL == "" { //空路径不设置
		setPanicRedirect = false
	}
	if !strings.Contains(panicFileURL, ":") && !strings.HasPrefix(panicFileURL, "/") { //相对路径
		panicFileURL = fmt.Sprintf("%s%s%s", fileutils.GetCurrentDirectory(), string(os.PathSeparator), panicFileURL)
	}

	// fileDir := ""
	// lastIndex := strings.LastIndex(panicFileURL, string(os.PathSeparator))
	// if lastIndex > 0 {
	// 	fileDir = panicFileURL[:lastIndex]
	// }

	// if err := os.MkdirAll(fileDir, 0755); err != nil {
	// 	panic(fmt.Sprintf("创建错误重定向文件夹路径出错:%s", err.Error()))
	// }

	if setPanicRedirect {
		stderrredirect.PanicRedirect(panicFileURL)
	}

	timeStr := now.Format("2006-01-02 15:04:05") //设定时间格式
	stderrredirect.PanicFile.WriteString("\n\n\n↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓\r\n")
	stderrredirect.PanicFile.WriteString(fmt.Sprintf("%s-%d-%s dump LOG\r\n", exeName, pid, timeStr))
	//	stderrredirect.PanicFile.WriteString(fmt.Sprintf("%v\r\n", err)) //输出panic信息
	stderrredirect.PanicFile.WriteString(string(debug.Stack())) //输出堆栈信息
	stderrredirect.PanicFile.WriteString("↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑\r\n")

	// file, err := os.OpenFile(panicFileURL, os.O_CREATE|os.O_APPEND, 0666)
	// if err != nil {
	// 	panic(fmt.Sprintf("panic重定向出错:%s", err.Error()))
	// }

	// defer file.Close()
}
