package web

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gdy666/lucky/config"
	"github.com/gdy666/lucky/thirdlib/gdylib/ginutils"
	"github.com/gin-gonic/gin"
)

func whitelistConfigure(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"ret": 0, "data": config.GetWhiteListBaseConfigure()})
}

func alterWhitelistConfigure(c *gin.Context) {
	var requestObj config.WhiteListBaseConfigure
	err := c.BindJSON(&requestObj)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"ret": 1, "msg": "修改请求解析出错"})
		return
	}

	requestObj.BasicAccount = strings.TrimSpace(requestObj.BasicAccount)
	if len(requestObj.BasicAccount) == 0 || len(requestObj.BasicPassword) == 0 {
		c.JSON(http.StatusOK, gin.H{"ret": 1, "msg": "账号或密码不能为空"})
		return
	}

	err = config.SetWhiteListBaseConfigure(requestObj.ActivelifeDuration, requestObj.URL, requestObj.BasicAccount, requestObj.BasicPassword)

	if err != nil {
		c.JSON(http.StatusOK, gin.H{"ret": 1, "msg": "保存白名单配置出错"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ret": 0})
}

func querywhitelist(c *gin.Context) {
	resList := config.GetWhiteList()
	c.JSON(http.StatusOK, gin.H{"ret": 0, "data": resList})
}

func deleteblacklist(c *gin.Context) {
	ip := c.Query("ip")
	err := config.BlackListDelete(ip)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"ret": 1, "msg": "删除黑名单指定IP出错"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ret": 0})
}

func deletewhitelist(c *gin.Context) {
	ip := c.Query("ip")
	err := config.WhiteListDelete(ip)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"ret": 1, "msg": "删除白名单指定IP出错"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ret": 0})
}

func flushblacklist(c *gin.Context) {
	ip := c.Query("ip")
	activelifeDurationStr := c.Query("life")
	life, _ := strconv.Atoi(activelifeDurationStr)

	newTime, err := config.BlackListAdd(ip, int32(life))
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"ret": 1, "msg": fmt.Sprintf("刷新IP有效期出错:%s", err.Error())})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ret": 0, "data": newTime})
}

func flushwhitelist(c *gin.Context) {
	ip := c.Query("ip")
	activelifeDurationStr := c.Query("life")
	life, _ := strconv.Atoi(activelifeDurationStr)

	newTime, err := config.WhiteListAdd(ip, int32(life))
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"ret": 1, "msg": fmt.Sprintf("刷新IP有效期出错:%s", err.Error())})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ret": 0, "data": newTime})
}

func queryblacklist(c *gin.Context) {
	resList := config.GetBlackList()
	c.JSON(http.StatusOK, gin.H{"ret": 0, "data": resList})
}

func whitelistBasicAuth(c *gin.Context) {
	bc := config.GetWhiteListBaseConfigure()
	whilelistURL := c.Param("url")
	if (c.Request.RequestURI == "/wl" && bc.URL != "") || whilelistURL != bc.URL {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	realm := "Basic realm=" + strconv.Quote("Authorization Required")
	pairs := ginutils.ProcessAccounts(gin.Accounts{bc.BasicAccount: bc.BasicPassword})
	user, found := pairs.SearchCredential(c.GetHeader("Authorization"))
	if !found {
		// Credentials doesn't match, we return 401 and abort handlers chain.
		c.Header("WWW-Authenticate", realm)
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	c.Set("user", user)
}

func whilelistAdd(c *gin.Context) {

	lifeTime, err := config.WhiteListAdd(c.ClientIP(), 0)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"ret": 1, "msg": "记录白名单IP出错"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"ret": 0, "msg": "IP已记录进白名单", "ip": c.ClientIP(), " effective_time": lifeTime})
}
