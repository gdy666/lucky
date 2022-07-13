package fileutils

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

//获取当前路径
func GetCurrentDirectory() string {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatal(err)
	}
	return dir
}

//保存base64为文件，一般用于保存图片
func SaveBase64AsFile(base64Str *string, fileURL string) (err error) {
	decodeStr, _ := base64.StdEncoding.DecodeString(*base64Str) //把base64写入缓存
	err = ioutil.WriteFile(fileURL, decodeStr, 0666)            //buffer输出到jpg文件中（不做处理，直接写到文件）
	return
}

//判断文件或文件夹是否存在
func FileExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

//保存Text到文本
func SaveTextToFile(text, fileURL string) error {
	dstFile, err := os.Create(fileURL)
	if err != nil {
		fmt.Println(err.Error())
		return err
	}
	defer dstFile.Close()
	dstFile.WriteString(text + "\n")
	return nil
}

//ReadTextFromFile 从文本读取内容
func ReadTextFromFile(path string) (string, error) {
	fi, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer fi.Close()
	fd, err := ioutil.ReadAll(fi)
	return string(fd), nil
}
