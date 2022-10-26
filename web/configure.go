package web

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/gdy666/lucky/config"
	"github.com/gdy666/lucky/thirdlib/gdylib/fileutils"
	"github.com/gdy666/lucky/thirdlib/gdylib/stringsp"
	"github.com/gin-gonic/gin"
)

func configure(c *gin.Context) {
	//c config.GetConfig()
	configureBytes := config.GetConfigureBytes()
	//c.Header("Content-Type", "application/json")

	//c.Data(http.StatusOK, "application/json", configureBytes)
	c.JSON(http.StatusOK,
		gin.H{
			"ret":       0,
			"time":      time.Now().Format("060102150405"),
			"configure": string(configureBytes)},
	)
}

var restoreConfigureVar *config.ProgramConfigure
var restoreConfigureKey string
var restoreConfigureMutex sync.Mutex

func restoreConfigureConfirm(c *gin.Context) {
	restoreConfigureMutex.Lock()
	defer restoreConfigureMutex.Unlock()
	key := c.Query("key")
	if key != restoreConfigureKey {
		c.JSON(http.StatusOK, gin.H{"ret": 1, "msg": "key不一致"})
		return
	}

	err := config.SetConfig(restoreConfigureVar)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"ret": 1, "msg": "还原配置出错"})
		return
	}

	rebootOnce.Do(func() {
		go func() {
			fileutils.OpenProgramOrFile(os.Args)
			os.Exit(0)
		}()
	})

	c.JSON(http.StatusOK, gin.H{"ret": 0, "msg": "还原配置成功", "port": restoreConfigureVar.BaseConfigure.AdminWebListenPort})

}

func restoreConfigure(c *gin.Context) {
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
	//log.Printf("file:%s\n", string(fileBytes))

	var conf config.ProgramConfigure

	err = json.Unmarshal(fileBytes, &conf)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"ret": 1, "msg": fmt.Sprintf("配置文件[%s]有误", file.Filename)})
		return
	}

	if conf.BaseConfigure.AdminAccount == "" ||
		conf.BaseConfigure.AdminPassword == "" ||
		conf.BaseConfigure.AdminWebListenPort <= 0 ||
		conf.BaseConfigure.AdminWebListenPort >= 65536 {
		c.JSON(http.StatusOK, gin.H{"ret": 1, "msg": fmt.Sprintf("配置文件[%s]参数有误", file.Filename)})
		return
	}

	restoreConfigureMutex.Lock()
	defer restoreConfigureMutex.Unlock()
	restoreConfigureVar = &conf
	restoreConfigureKey = stringsp.GetRandomStringNum(16)

	c.JSON(http.StatusOK, gin.H{"ret": 0, "file": file.Filename, "restoreConfigureKey": restoreConfigureKey})

}
