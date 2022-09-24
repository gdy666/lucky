package web

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gdy666/lucky/config"
	"github.com/gdy666/lucky/ddnscore.go"
	"github.com/gdy666/lucky/thirdlib/gdylib/service"
	"github.com/gin-gonic/gin"
)

func addDDNS(c *gin.Context) {
	var requestObj config.DDNSTask

	err := c.BindJSON(&requestObj)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"ret": 1, "msg": "请求解析出错"})
		return
	}
	//fmt.Printf("addDDNS requestObj:%v\n", requestObj)
	err = config.CheckDDNSTaskAvalid(&requestObj)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"ret": 1, "msg": err.Error()})
		return
	}

	dealRequestDDNSTask(&requestObj)

	err = config.DDNSTaskListAdd(&requestObj)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"ret": 1, "msg": "DDNS任务添加出错"})
		return
	}

	if requestObj.Enable {
		service.Message("ddns", "syncDDNSTask", requestObj.TaskKey)
	}

	c.JSON(http.StatusOK, gin.H{"ret": 0, "msg": ""})
}

func deleteDDNSTask(c *gin.Context) {
	taskKey := c.Query("key")
	err := config.DDNSTaskListDelete(taskKey)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"ret": 1, "msg": fmt.Errorf("删除DDNS任务出错:%s", err.Error())})
		return
	}

	ddnscore.DDNSTaskInfoMapDelete(taskKey)

	c.JSON(http.StatusOK, gin.H{"ret": 0, "msg": ""})
}

func enableddns(c *gin.Context) {
	enable := c.Query("enable")
	key := c.Query("key")

	var err error

	if enable == "true" {
		err = ddnscore.EnableDDNSTaskByKey(key, true)
		if err == nil {
			service.Message("ddns", "syncDDNSTask", key)
		}
	} else {
		err = ddnscore.EnableDDNSTaskByKey(key, false)
	}

	if err != nil {
		c.JSON(http.StatusOK, gin.H{"ret": 1, "msg": fmt.Sprintf("开关DDNS任务出错:%s", err.Error())})
		return
	}

	c.JSON(http.StatusOK, gin.H{"ret": 0, "msg": ""})
}

func ddnsconfigure(c *gin.Context) {
	conf := config.GetDDNSConfigure()

	c.JSON(http.StatusOK, gin.H{"ret": 0, "ddnsconfigure": conf})
}

func alterDDNSTask(c *gin.Context) {
	taskKey := c.Query("key")
	var requestObj config.DDNSTask
	err := c.BindJSON(&requestObj)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"ret": 1, "msg": "请求解析出错"})
		return
	}

	err = config.CheckDDNSTaskAvalid(&requestObj)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"ret": 1, "msg": err.Error()})
		return
	}
	dealRequestDDNSTask(&requestObj)

	err = config.UpdateTaskToDDNSTaskList(taskKey, requestObj)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"ret": 1, "msg": fmt.Sprintf("更新DDNS任务出错:%s", err.Error())})
		return
	}

	ddnscore.DDNSTaskInfoMapDelete(taskKey)

	if requestObj.Enable {
		service.Message("ddns", "syncDDNSTask", taskKey)
	}

	c.JSON(http.StatusOK, gin.H{"ret": 0, "msg": ""})

}

func ddnsTaskList(c *gin.Context) {

	conf := config.GetDDNSConfigure()

	if !conf.Enable {
		c.JSON(http.StatusOK, gin.H{"ret": 6, "msg": "请先在设置页面启用DDNS动态域名服务"})
		return
	}

	taskList := ddnscore.GetDDNSTaskInfoList()
	ddnscore.FLushWebLastAccessDDNSTaskListLastTime()

	c.JSON(http.StatusOK, gin.H{"ret": 0, "data": taskList})
}

func dealRequestDDNSTask(t *config.DDNSTask) {

	if t.DNS.Name == "callback" {
		t.DNS.ID = ""
		t.DNS.Secret = ""
		t.DNS.Callback.URL = strings.TrimSpace(t.DNS.Callback.URL)
		//requestObj.DNS.Callback.CallbackSuccessContent = strings.TrimSpace(requestObj.DNS.Callback.CallbackSuccessContent)
		t.DNS.Callback.RequestBody = strings.TrimSpace(t.DNS.Callback.RequestBody)
	} else {
		t.DNS.Callback = config.DNSCallback{}
	}

	if !t.DNS.ResolverDoaminCheck && len(t.DNS.DNSServerList) > 0 {
		t.DNS.DNSServerList = []string{}
	}

	if t.DNS.ResolverDoaminCheck && (len(t.DNS.DNSServerList) == 0 || (len(t.DNS.DNSServerList) == 1 && t.DNS.DNSServerList[0] == "")) {
		if t.TaskType == "IPv6" {
			t.DNS.DNSServerList = config.DefaultIPv6DNSServerList
		} else {
			t.DNS.DNSServerList = config.DefaultIPv4DNSServerList
		}
	}

	if t.DNS.HttpClientProxyType != "" && t.DNS.HttpClientProxyAddr == "" {
		t.DNS.HttpClientProxyType = ""
	}

	if t.DNS.HttpClientProxyType == "" {
		t.DNS.HttpClientProxyAddr = ""
		t.DNS.HttpClientProxyUser = ""
		t.DNS.HttpClientProxyPassword = ""
	}

	if t.GetType == "url" {
		t.NetInterface = ""
		t.IPReg = ""
	}

	if t.GetType == "netInterface" {
		t.URL = []string{}
	}

	if t.DNS.Name == "callback" {
		if t.DNS.Callback.DisableCallbackSuccessContentCheck {
			t.DNS.Callback.CallbackSuccessContent = []string{}
		}
	}

	if !t.WebhookEnable {
		t.WebhookHeaders = []string{}
		t.WebhookMethod = ""
		t.WebhookRequestBody = ""
		t.WebhookURL = ""
		t.WebhookSuccessContent = []string{}
		t.WebhookProxy = ""
		t.WebhookProxyAddr = ""
		t.WebhookProxyUser = ""
		t.WebhookProxyPassword = ""
	}

	if t.WebhookEnable {
		if t.WebhookMethod == "get" {
			t.WebhookRequestBody = ""
		}

		if t.WebhookProxy == "" {
			t.WebhookProxyAddr = ""
			t.WebhookProxyUser = ""
			t.WebhookProxyPassword = ""
		}
	}

	if t.DNS.ForceInterval < 60 {
		t.DNS.ForceInterval = 60
	} else if t.DNS.ForceInterval > 360000 {
		t.DNS.ForceInterval = 360000
	}

	if t.HttpClientTimeout < 3 {
		t.HttpClientTimeout = 3
	} else if t.HttpClientTimeout > 60 {
		t.HttpClientTimeout = 60
	}

}

func alterDDNSConfigure(c *gin.Context) {
	var requestObj config.DDNSConfigure
	err := c.BindJSON(&requestObj)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"ret": 1, "msg": "请求解析出错"})
		return
	}

	preConfigure := config.GetDDNSConfigure()

	if preConfigure.Enable != requestObj.Enable {

		//log.Printf("动态服务服务状态改变:%v", requestObj.Enable)
		if requestObj.Enable {
			service.Start("ddns")
		} else {
			service.Stop("ddns")
		}

	}

	err = config.SetDDNSConfigure(&requestObj)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"ret": 2, "msg": "保存配置过程发生错误,请检测相关启动配置"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ret": 0})
}

func webhookTest(c *gin.Context) {
	key := c.Query("key")
	ddnsTask := ddnscore.GetDDNSTaskInfoByKey(key)

	if ddnsTask == nil {
		c.JSON(http.StatusOK, gin.H{"ret": 1, "msg": fmt.Sprintf("找不到key对应的DDNS任务:%s", key)})
		return
	}

	var request struct {
		WebhookURL            string   `json:"WebhookURL"`
		WebhookMethod         string   `json:"WebhookMethod"`
		WebhookHeaders        []string `json:"WebhookHeaders"`
		WebhookRequestBody    string   `json:"WebhookRequestBody"`
		WebhookSuccessContent []string `json:"WebhookSuccessContent"` //接口调用成功包含的内容
		WebhookProxy          string   `json:"WebhookProxy"`          //使用DNS代理设置  ""表示禁用，"dns"表示使用dns的代理设置
		WebhookProxyAddr      string   `json:"WebhookProxyAddr"`      //代理服务器IP
		WebhookProxyUser      string   `json:"WebhookProxyUser"`      //代理用户
		WebhookProxyPassword  string   `json:"WebhookProxyPassword"`  //代理密码
	}
	err := c.Bind(&request)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"ret": 1, "msg": fmt.Sprintf("请求解析出错:%s", err.Error())})
		return
	}

	responseStr, err := ddnscore.WebhookTest(ddnsTask,
		request.WebhookURL,
		request.WebhookMethod,
		request.WebhookRequestBody,
		request.WebhookProxy,
		request.WebhookProxyAddr,
		request.WebhookProxyUser,
		request.WebhookProxyPassword,
		request.WebhookHeaders,
		request.WebhookSuccessContent)

	msg := "Webhook接口调用成功"

	if err != nil {
		msg = err.Error()
	}

	c.JSON(http.StatusOK, gin.H{"ret": 0, "msg": msg, "Response": responseStr})
}
