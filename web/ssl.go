package web

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gdy666/lucky/config"
	"github.com/gin-gonic/gin"
)

func addSSL(c *gin.Context) {
	var requestObj config.SSLCertficate
	err := c.BindJSON(&requestObj)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"ret": 1, "msg": "请求解析出错"})
		return
	}

	err = requestObj.Init()
	if err != nil {
		log.Printf("addSSL requestObj.Init() error:%s", err.Error())
		c.JSON(http.StatusOK, gin.H{"ret": 2, "msg": "证书和Key有误!"})
		return
	}

	//fmt.Printf("CertsInfo:%v\n", *requestObj.CertsInfo)
	err = config.SSLCertficateListAdd(&requestObj)
	if err != nil {
		log.Printf("config.SSLCertficateListAdd error:%s", err.Error())
		c.JSON(http.StatusOK, gin.H{"ret": 2, "msg": fmt.Sprintf("添加SSL证书出错!:%s", err.Error())})
		return
	}

	c.JSON(http.StatusOK, gin.H{"ret": 0})

}

type sslResInfo struct {
	Key       string             `json:"Key"`
	Remark    string             `json:"Remark"`
	Enable    bool               `json:"Enable"`
	AddTime   string             `json:"AddTime"`
	CertsInfo *[]config.CertInfo `json:"CertsInfo"`
}

func getSSLCertficateList(c *gin.Context) {
	rawList := config.GetSSLCertficateList()
	var res []sslResInfo
	for i := range rawList {
		info := sslResInfo{
			Key:       rawList[i].Key,
			Remark:    rawList[i].Remark,
			Enable:    rawList[i].Enable,
			AddTime:   rawList[i].AddTime,
			CertsInfo: rawList[i].CertsInfo,
		}

		res = append(res, info)
	}
	c.JSON(http.StatusOK, gin.H{"ret": 0, "list": res})
}

func alterSSLCertficate(c *gin.Context) {
	key := c.Query("key")
	field := c.Query("field")
	value := c.Query("value")
	var err error
	switch field {
	case "enable":
		{
			enable := false

			if value == "true" {
				enable = true
			}
			err = config.SSLCertficateEnable(key, enable)
		}
	case "remark":
		{
			err = config.SSLCertficateAlterRemark(key, value)
		}
	default:
		err = fmt.Errorf("不支持修改的字段:%s", field)
	}

	if err != nil {
		c.JSON(http.StatusOK, gin.H{"ret": 1, "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ret": 0})
}

func deleteSSLCertficate(c *gin.Context) {
	key := c.Query("key")
	err := config.SSLCertficateListDelete(key)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"ret": 1, "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ret": 0})
}
