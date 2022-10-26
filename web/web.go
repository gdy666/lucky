package web

import (
	"crypto/tls"
	"embed"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"

	"github.com/gdy666/lucky/config"
	"github.com/gdy666/lucky/socketproxy"
	"github.com/gdy666/lucky/thirdlib/gdylib/fileutils"
	"github.com/gdy666/lucky/thirdlib/gdylib/ginutils"
	"github.com/gdy666/lucky/thirdlib/gdylib/logsbuffer"
	"github.com/gdy666/lucky/thirdlib/gdylib/netinterfaces"
	"github.com/gdy666/lucky/thirdlib/gdylib/stringsp"
	"github.com/golang-jwt/jwt"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/process"
)

//go:embed adminviews/dist
var staticFs embed.FS
var stafs fs.FS
var loginErrorCount = int32(0)
var rebootOnce sync.Once
var logBuffer *logsbuffer.LogsBuffer

type LogItem struct {
	Timestamp string `json:"timestamp"`
	Content   string `json:"log"`
}

func logConvert(lg *logsbuffer.LogItem) any {
	l := LogItem{Content: lg.Content, Timestamp: fmt.Sprintf("%d", lg.Timestamp)}
	return l
}

func init() {
	stafs, _ = fs.Sub(staticFs, "adminviews/dist")
	logBuffer = logsbuffer.Create(1024)
	//logBuffer.SetLogItemConverFunc(logConvert)
	log.SetOutput(io.MultiWriter(logBuffer, os.Stdout))

}

func RunAdminWeb(conf *config.BaseConfigure) {

	//gin.Default()
	logBuffer.SetBufferSize(conf.LogMaxSize)

	gin.SetMode(gin.ReleaseMode)
	r := gin.New()

	if gin.Mode() != gin.ReleaseMode {
		r.Use(gin.Logger(), gin.Recovery())
	} else {
		r.Use(gin.Recovery())
	}

	r.Use(checkLocalIP)
	r.Use(gzip.Gzip(gzip.DefaultCompression))
	r.Use(ginutils.Cors())

	r.Use(ginutils.HandlerStaticFiles(stafs))

	//r.Use(sessionCheck())
	//r.StaticFS("/", http.FS(stafs))

	authorized := r.Group("/")
	authorized.Use(tokenCheck())
	{
		authorized.GET("/api/logs", Logs)
		authorized.GET("/api/status", status)
		authorized.GET("/api/test", test)

		authorized.GET("/api/portforwards", PortForwardsRuleList)
		authorized.POST("/api/portforward", PortForwardsRuleAdd)
		authorized.DELETE("/api/portforward", PortForwardsRuleDelete)
		authorized.PUT("/api/portforward", PortForwardsRuleAlter)
		authorized.GET("/api/portforward/enable", PortForwardsRuleEnable)
		authorized.GET("/api/portforward/configure", portforwardConfigure)
		authorized.PUT("/api/portforward/configure", alterPortForwardConfigure)
		authorized.GET("/api/portforward/logs", getPortwardRuleLogs)

		authorized.GET("/api/baseconfigure", baseconfigure)
		authorized.PUT("/api/baseconfigure", alterBaseConfigure)
		authorized.GET("/api/reboot_program", rebootProgram)

		authorized.GET("/api/whitelist/configure", whitelistConfigure)
		authorized.PUT("/api/whitelist/configure", alterWhitelistConfigure)
		authorized.GET("/api/whitelist", querywhitelist)
		authorized.PUT("/api/whitelist/flush", flushwhitelist)
		authorized.DELETE("/api/whitelist", deletewhitelist)
		authorized.GET("/api/blacklist", queryblacklist)
		authorized.PUT("/api/blacklist/flush", flushblacklist)
		authorized.DELETE("/api/blacklist", deleteblacklist)

		authorized.POST("/api/ddns", addDDNS)
		authorized.PUT("/api/ddns", alterDDNSTask)
		authorized.GET("/api/ddnstasklist", ddnsTaskList)
		authorized.DELETE("/api/ddns", deleteDDNSTask)
		authorized.GET("/api/ddns/enable", enableddns)
		authorized.GET("/api/ddns/configure", ddnsconfigure)
		authorized.PUT("/api/ddns/configure", alterDDNSConfigure)

		authorized.GET("/api/netinterfaces", getNetinterfaces)
		authorized.GET("/api/ipregtest", IPRegTest)
		authorized.POST("/api/webhooktest", webhookTest)

		authorized.GET("/api/reverseproxyrules", reverseProxys)
		authorized.POST("/api/reverseproxyrule", addReverseProxyRule)
		authorized.PUT("/api/reverseproxyrule", alterReverseProxyRule)
		authorized.DELETE("/api/reverseproxyrule", deleteReverseProxyRule)
		authorized.GET("/api/reverseproxyrule/enable", enableReverseProxyRule)
		authorized.GET("/api/reverseproxyrule/logs", getReverseProxyLog)

		authorized.POST("/api/ssl", addSSL)
		authorized.GET("/api/ssl", getSSLCertficateList)
		authorized.PUT("/api/ssl", alterSSLCertficate)
		authorized.DELETE("/api/ssl", deleteSSLCertficate)

		authorized.POST("/api/wol/device", addWOLDevice)
		authorized.GET("/api/wol/device/wakeup", WOLDeviceWakeUp)
		authorized.GET("/api/wol/devices", getWOLDeviceList)
		authorized.PUT("/api/wol/device", alterWOLDevice)
		authorized.DELETE("/api/wol/device", deleteWOLDevice)

		authorized.GET("/api/info", info)
		authorized.GET("/api/configure", configure)
		authorized.POST("/api/configure", restoreConfigure)
		authorized.POST("/api/getfilebase64", getFileBase64)

		authorized.GET("/api/restoreconfigureconfirm", restoreConfigureConfirm)
		r.PUT("/api/logout", logout)
	}
	r.POST("/api/login", login)
	//r.GET("/FreeOSMemory", FreeOSMemory)

	r.GET("/wl", whitelistBasicAuth, whilelistAdd)
	r.GET("/wl/:url", whitelistBasicAuth, whilelistAdd)
	r.GET("/version", queryVersion)

	//r.Use(func() *gin.Context {})

	go func() {
		httpListen := fmt.Sprintf(":%d", conf.AdminWebListenPort)
		log.Printf("AdminWeb(Http) listen on %s", httpListen)
		err := r.Run(httpListen)
		if err != nil {
			log.Printf("Admin Http Listen error:%s", err.Error())
			os.Exit(1)
		}
	}()

	if conf.AdminWebListenTLS {
		certlist := config.GetValidSSLCertficateList()
		if len(certlist) <= 0 {
			log.Printf("可用SSL证书列表为空,AdminWeb(Https) 监听服务中止运行")
			return
		}
		httpsListen := fmt.Sprintf(":%d", conf.AdminWebListenHttpsPort)

		server := &http.Server{
			Addr:    httpsListen,
			Handler: r,
		}
		server.TLSConfig = &tls.Config{}
		server.TLSConfig.Certificates = certlist
		ln, err := net.Listen("tcp", httpsListen)
		if err != nil {
			log.Fatalf("Admin Https Listen error:%s", err.Error())
		}
		log.Printf("AdminWeb(Https) listen on %s", httpsListen)
		err = server.ServeTLS(ln, "", "")
		if err != nil {
			log.Printf("AdminWeb(Https) Server error:%s", err.Error())
		}
	}

}

// func FreeOSMemory(c *gin.Context) {
// 	debug.FreeOSMemory()
// 	c.JSON(http.StatusOK, gin.H{"ret": 0})
// }

// Logs web
func Logs(c *gin.Context) {
	preTimeStampStr := c.Query("pre")
	preTimeStamp, _ := strconv.ParseInt(preTimeStampStr, 10, 64)
	logs := logBuffer.GetLogs(logConvert, preTimeStamp)

	logCount := logBuffer.GetLogCount()

	c.JSON(http.StatusOK, gin.H{
		"ret":       0,
		"logs":      logs,
		"logsCount": logCount,
	})

}

func info(c *gin.Context) {
	info := config.GetAppInfo()
	c.JSON(http.StatusOK, gin.H{"ret": 0, "info": *info})

}

func logout(c *gin.Context) {
	config.FlushLoginRandomKey()
	c.JSON(http.StatusOK, gin.H{"ret": 0, "msg": "已注销登录"})
}

func queryVersion(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"ret": 0, "version": config.GetVersion()})
}

func checkLocalIP(c *gin.Context) {
	clientIP := c.ClientIP()
	//fmt.Printf("clientIP:%s\n", clientIP)
	bc := config.GetBaseConfigure()

	if !isLocalIP(clientIP) && !bc.AllowInternetaccess {
		c.JSON(http.StatusForbidden, gin.H{
			"ret": 1,
			"msg": "外网禁止访问,如果需要允许外网访问请在后台设置中打开相应开关.",
			"ip":  clientIP})
		c.Abort()
		return
	}

}

func tokenCheck() gin.HandlerFunc {
	return func(c *gin.Context) {

		// if config.GetRunMode() == "dev" {
		// 	c.Next()
		// 	return
		// }

		tokenString, _ := c.GetQuery("Authorization")
		if tokenString == "" {
			tokenString = c.GetHeader("Authorization")
		}

		token, err := ginutils.GetJWTToken(tokenString, "strings")
		if err != nil {
			c.AbortWithStatusJSON(http.StatusOK, gin.H{"ret": -1, "msg": "登录失效"})
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.AbortWithStatusJSON(http.StatusOK, gin.H{"ret": -1, "msg": "登录失效"})
			return
		}

		account := claims["account"].(string)
		password := claims["password"].(string)
		loginKey := claims["loginkey"].(string)

		if account == "" || password == "" {
			c.AbortWithStatusJSON(http.StatusOK, gin.H{"ret": -1, "msg": "登录失效"})
			return
		}

		bc := config.GetBaseConfigure()

		// //fmt.Printf("session中的account:%s password:%s\n", account, password)
		if bc.AdminAccount != account || bc.AdminPassword != password || loginKey != config.GetLoginRandomKey() {
			c.AbortWithStatusJSON(http.StatusOK, gin.H{"ret": -1, "msg": "登录失效"})
			return
		}
		c.Next()
	}
}

func rebootProgram(c *gin.Context) {
	rebootOnce.Do(func() {
		go func() {
			fileutils.OpenProgramOrFile(os.Args)
			os.Exit(0)
		}()
	})

	c.JSON(http.StatusOK, gin.H{"ret": 0, "msg": ""})
}

func login(c *gin.Context) {
	var requestObj struct {
		Account  string `json:"Account"`
		Password string `json:"Password"`
	}
	err := c.BindJSON(&requestObj)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"ret": 1, "msg": "登录失败,登录请求解析出错"})
		return
	}

	if atomic.LoadInt32(&loginErrorCount) >= 99 {
		c.JSON(http.StatusOK, gin.H{"ret": 1, "msg": "登录错误次数太多,后台登录功能已禁用,请重启程序."})
		return
	}

	bc := config.GetBaseConfigure()

	if bc.AdminAccount != requestObj.Account || bc.AdminPassword != requestObj.Password {
		c.JSON(http.StatusOK, gin.H{"ret": 1, "msg": "登录失败,账号或密码有误"})
		atomic.AddInt32(&loginErrorCount, 1)
		return
	}

	config.FlushLoginRandomKey()
	tokenInfo := make(map[string]interface{})
	tokenInfo["account"] = requestObj.Account //用户名
	tokenInfo["password"] = requestObj.Password
	tokenInfo["loginkey"] = config.GetLoginRandomKey()
	tokenString, err := ginutils.GetJWTTokenString(tokenInfo, "strings", time.Hour*24)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"ret": 1, "msg": "登录失败,token生成出错"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"ret": 0, "msg": "登录成功", "token": tokenString})
}

func alterBaseConfigure(c *gin.Context) {
	var requestObj config.BaseConfigure
	err := c.BindJSON(&requestObj)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"ret": 1, "msg": "请求解析出错"})
		return
	}
	requestObj.AdminAccount = strings.TrimSpace(requestObj.AdminAccount)

	if len(requestObj.AdminAccount) == 0 || len(requestObj.AdminPassword) == 0 {
		c.JSON(http.StatusOK, gin.H{"ret": 1, "msg": "账号或密码不能为空"})
		return
	}

	preBaseConfigure := config.GetBaseConfigure()
	if preBaseConfigure.AdminWebListenPort != requestObj.AdminWebListenPort && !config.CheckTCPPortAvalid(requestObj.AdminWebListenPort) { //检测新端口
		c.JSON(http.StatusOK, gin.H{"ret": 1, "msg": fmt.Sprintf("新的后端管理监听端口[%d]已被占用,修改设置失败", requestObj.AdminWebListenPort)})
		return
	}

	err = config.SetBaseConfigure(&requestObj)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"ret": 2, "msg": "保存配置过程发生错误,请检测相关启动配置"})
		return
	}

	if preBaseConfigure.LogMaxSize != requestObj.LogMaxSize {
		logBuffer.SetBufferSize(requestObj.LogMaxSize)
	}

	c.JSON(http.StatusOK, gin.H{"ret": 0})
}

func baseconfigure(c *gin.Context) {
	conf := config.GetBaseConfigure()
	c.JSON(http.StatusOK, gin.H{"ret": 0, "baseconfigure": conf})
}

func getNetinterfaces(c *gin.Context) {
	ipv4NetInterfaces, ipv6Netinterfaces, err := netinterfaces.GetNetInterface()
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"ret": 1, "msg": fmt.Sprintf("获取网卡列表出错：%s", err.Error())})
		return
	}

	c.JSON(http.StatusOK, gin.H{"ret": 0, "data": gin.H{"IPv6NewInterfaces": ipv6Netinterfaces, "IPv4NewInterfaces": ipv4NetInterfaces}})
}

func IPRegTest(c *gin.Context) {
	iptype := c.Query("iptype")
	netinterface := c.Query("netinterface")
	ipreg := c.Query("ipreg")

	ip := netinterfaces.GetIPFromNetInterface(iptype, netinterface, ipreg)

	c.JSON(http.StatusOK, gin.H{"ret": 0, "ip": ip})
}

func test(c *gin.Context) {

	c.JSON(http.StatusOK, gin.H{"ret": 0})
}

func status(c *gin.Context) {

	v, _ := mem.VirtualMemory()

	currentProcessMem := GetCurrentProcessMem()
	//fmt.Fprintf(w, "当前进程 CPU使用率:%.2f%% 协程数:%d 进程内存使用:%s 系统内存总量:%s 已用:%s 可用:%s \n", GetCurrentProcessCPUPrecent(), runtime.NumGoroutine(), formatFileSize(currentProcessMem), formatFileSize(v.Total), formatFileSize(v.Used), formatFileSize(v.Free))
	//fmt.Fprintf(w, "当前全局TCP 连接数:%d   全局TCP连接数最大限制:%d\n", core.GetGlobalTCPConns(), core.GetGlobalMaxConnections())

	//var proxyStatusList []string

	// for _, p := range *config.GlobalProxy {
	// 	//fmt.Fprintf(w, "%s\n", p.GetStatus())
	// 	proxyStatusList = append(proxyStatusList, p.GetStatus())
	// }

	appInfo := config.GetAppInfo()

	respMap := make(map[string]interface{})
	respMap["totleMem"] = stringsp.BinaryUnitToStr(v.Total)
	respMap["usedMem"] = stringsp.BinaryUnitToStr(v.Used)
	respMap["unusedMem"] = stringsp.BinaryUnitToStr(v.Free)
	respMap["currentProcessUsedCPU"] = fmt.Sprintf("%.2f%%", GetCurrentProcessCPUPrecent())
	respMap["goroutine"] = fmt.Sprintf("%d", runtime.NumGoroutine())
	respMap["processUsedMem"] = stringsp.BinaryUnitToStr(currentProcessMem)
	respMap["currentTCPConnections"] = fmt.Sprintf("%d", socketproxy.GetGlobalTCPPortForwardConnections())
	respMap["currentUDPConnections"] = fmt.Sprintf("%d", socketproxy.GetGlobalUDPPortForwardGroutineCount())
	respMap["maxTCPConnections"] = fmt.Sprintf("%d", socketproxy.GetGlobalTCPPortforwardMaxConnections())
	respMap["usedCPU"] = fmt.Sprintf("%.2f%%", GetCpuPercent())
	respMap["runTime"] = appInfo.RunTime
	//respMap["proxysStatus"] = proxyStatusList

	c.JSON(http.StatusOK, gin.H{
		"ret":  0,
		"data": respMap,
	})
}

func GetCurrentProcessMem() uint64 {
	plist, e := process.Processes()
	if e == nil {
		for _, p := range plist {
			if int(p.Pid) == os.Getpid() {
				mem, err := p.MemoryInfo()
				if err != nil {
					return 0
				}
				return mem.RSS
			}
		}
	}
	return 0
}

func GetCurrentProcessCPUPrecent() float64 {
	plist, e := process.Processes()
	if e == nil {
		for _, p := range plist {
			if int(p.Pid) == os.Getpid() {
				cpuprecent, err := p.CPUPercent()
				if err != nil {
					return 0
				}
				return cpuprecent
			}
		}
	}
	return 0
}

func GetCpuPercent() float64 {
	percent, _ := cpu.Percent(time.Second, false)
	return percent[0]
}

//------------------------------------------------------------------------------------------------------------------

func isLocalIP(ipstr string) bool {

	ip := net.ParseIP(ipstr)

	if ip.IsLoopback() {
		return true
	}

	ip4 := ip.To4()
	if ip4 == nil {
		return false
	}

	return ip4[0] == 10 || // 10.0.0.0/8
		(ip4[0] == 172 && ip4[1] >= 16 && ip4[1] <= 31) || // 172.16.0.0/12
		(ip4[0] == 169 && ip4[1] == 254) || // 169.254.0.0/16
		(ip4[0] == 192 && ip4[1] == 168) // 192.168.0.0/16
}
