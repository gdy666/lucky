package web

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gdy666/lucky/config"
	"github.com/gin-gonic/gin"
)

func addWOLDevice(c *gin.Context) {
	var requestObj config.WOLDevice
	err := c.BindJSON(&requestObj)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"ret": 1, "msg": "请求解析出错"})
		return
	}

	err = checkWolDevice(&requestObj)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"ret": 2, "msg": fmt.Sprintf("添加网络唤醒设备出错:%s", err.Error())})
		return
	}

	err = config.WOLDeviceListAdd(&requestObj)
	if err != nil {
		log.Printf("config.WOLDeviceListAdd error:%s", err.Error())
		c.JSON(http.StatusOK, gin.H{"ret": 2, "msg": fmt.Sprintf("添加网络唤醒设备出错!:%s", err.Error())})
		return
	}

	c.JSON(http.StatusOK, gin.H{"ret": 0})
}

func getWOLDeviceList(c *gin.Context) {
	list := config.GetWOLDeviceList()
	c.JSON(http.StatusOK, gin.H{"ret": 0, "list": list})
}

func alterWOLDevice(c *gin.Context) {
	var requestObj config.WOLDevice
	err := c.BindJSON(&requestObj)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"ret": 1, "msg": "请求解析出错"})
		return
	}

	err = checkWolDevice(&requestObj)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"ret": 2, "msg": fmt.Sprintf("修改网络唤醒设备出错:%s", err.Error())})
		return
	}

	err = config.WOLDeviceListAlter(&requestObj)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"ret": 1, "msg": fmt.Sprintf("修改网络唤醒设备配置失败:%s", err.Error())})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ret": 0})
}

func deleteWOLDevice(c *gin.Context) {
	key := c.Query("key")
	err := config.WOLDeviceListDelete(key)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"ret": 1, "msg": fmt.Sprintf("删除网络唤醒设备失败:%s", err.Error())})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ret": 0})
}

func WOLDeviceWakeUp(c *gin.Context) {
	key := c.Query("key")

	device := config.GetWOLDeviceByKey(key)
	if device == nil {
		c.JSON(http.StatusOK, gin.H{"ret": 1, "msg": fmt.Sprintf("找不到Key对应的设备,唤醒失败")})
		return
	}
	err := device.WakeUp()
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"ret": 2, "msg": "唤醒失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ret": 0})
}

func checkWolDevice(d *config.WOLDevice) error {

	// if strings.TrimSpace(d.DeviceName) == "" {
	// 	return fmt.Errorf("设备名称不能为空")
	// }

	if d.Port <= 0 || d.Port > 065535 {
		d.Port = 9
	}

	if d.Repeat <= 0 || d.Repeat > 10 {
		d.Repeat = 5
	}
	return nil
}
