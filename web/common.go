package web

import (
	"encoding/base64"
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
)

func getFileBase64(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"ret": 1, "msg": fmt.Sprintf("c.FormFile err:%s", err.Error())})
		return
	}

	src, err := file.Open()
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"ret": 1, "msg": fmt.Sprintf("file.Open err:%s", err.Error())})
		return
	}
	defer src.Close()

	fileBytes, err := io.ReadAll(src)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"ret": 1, "msg": fmt.Sprintf("ioutil.ReadAll err:%s", err.Error())})
		return
	}

	fileBytesBase64Str := base64.StdEncoding.EncodeToString(fileBytes)

	c.JSON(http.StatusOK, gin.H{"ret": 0, "file": file.Filename, "base64": fileBytesBase64Str})

}
