package web

import (
	"crypto/subtle"
	"embed"
	"encoding/base64"
	"fmt"
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
	"unsafe"

	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"

	"github.com/golang-jwt/jwt"
	"github.com/ljymc/goports/base"
	"github.com/ljymc/goports/config"
	"github.com/ljymc/goports/rule"
	"github.com/ljymc/goports/thirdlib/gdylib/fileutils"
	"github.com/ljymc/goports/thirdlib/gdylib/ginutils"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/process"
)

//go:embed goports-adminviews/dist
var staticFs embed.FS
var stafs fs.FS
var loginErrorCount = int32(0)
var rebootOnce sync.Once

//store := cookie.NewStore([]byte("secret11111"))
//var fileServer http.Handler
//var cookieStore cookie.Store

func init() {
	stafs, _ = fs.Sub(staticFs, "goports-adminviews/dist")
	//cookieStore = cookie.NewStore([]byte("goports2022"))
}

func RunAdminWeb(listen string) {

	//gin.Default()

	gin.SetMode(gin.ReleaseMode)
	r := gin.New()

	if gin.Mode() != gin.ReleaseMode {
		r.Use(gin.Logger(), gin.Recovery())
	} else {
		r.Use(gin.Recovery())
	}

	r.Use(checkLocalIP)

	//r.Use(sessions.Sessions("goportssession", cookieStore))

	r.Use(gzip.Gzip(gzip.DefaultCompression))

	// if config.GetRunMode() == "dev" {
	// 	r.Use(CrosHandler())
	// }
	r.Use(ginutils.Cors())

	r.Use(HandlerStaticFiles())

	//r.Use(sessionCheck())
	//r.StaticFS("/", http.FS(stafs))

	authorized := r.Group("/")
	authorized.Use(tokenCheck())
	{
		authorized.GET("/api/logs", Logs)
		authorized.GET("/api/status", status)
		authorized.GET("/api/test", test)
		authorized.GET("/api/rulelist", rulelist)
		authorized.POST("/api/rule", addrule)
		authorized.DELETE("/api/rule", deleterule)
		authorized.PUT("/api/rule", alterrule)
		authorized.GET("/api/rule/enable", enablerule)
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
		r.PUT("/api/logout", logout)
	}
	r.POST("/api/login", login)

	r.GET("/wl", whitelistBasicAuth, whilelistAdd)
	r.GET("/wl/:url", whitelistBasicAuth, whilelistAdd)
	r.GET("/version", queryVersion)

	//r.Use(func() *gin.Context {})

	err := r.Run(listen)

	if err != nil {
		log.Printf("http.ListenAndServe error:%s", err.Error())
		time.Sleep(time.Minute)
		os.Exit(1)
	}
}

func logout(c *gin.Context) {
	config.FlushLoginRandomKey()
	c.JSON(http.StatusOK, gin.H{"ret": 0, "msg": "已注销登录"})
}

func queryVersion(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"ret": 0, "version": config.GetVersion()})
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
		c.JSON(http.StatusOK, gin.H{"ret": 1, "msg": "刷新IP有效期出错"})
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
		c.JSON(http.StatusOK, gin.H{"ret": 1, "msg": "刷新IP有效期出错"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ret": 0, "data": newTime})
}

func queryblacklist(c *gin.Context) {
	resList := config.GetBlackList()
	c.JSON(http.StatusOK, gin.H{"ret": 0, "data": resList})
}

func querywhitelist(c *gin.Context) {
	resList := config.GetWhiteList()
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
	pairs := processAccounts(gin.Accounts{bc.BasicAccount: bc.BasicPassword})
	user, found := pairs.searchCredential(c.GetHeader("Authorization"))
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

	c.JSON(http.StatusOK, gin.H{"ret": 0, "msg": fmt.Sprintf("IP已记录进白名单"), "ip": c.ClientIP(), " effective_time": lifeTime})
}

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
	return
}

func checkLocalIP(c *gin.Context) {
	clientIP := c.ClientIP()
	//fmt.Printf("clientIP:%s\n", clientIP)
	bc := config.GetBaseConfigure()

	if !isLocalIP(clientIP) && !bc.AllowInternetaccess {
		c.JSON(http.StatusForbidden, gin.H{"ret": 1, "msg": "Forbidden Internetaccess "})
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
	c.JSON(http.StatusOK, gin.H{"ret": 0})
}

func baseconfigure(c *gin.Context) {
	conf := config.GetBaseConfigure()
	c.JSON(http.StatusOK, gin.H{"ret": 0, "baseconfigure": conf})
}

func enablerule(c *gin.Context) {

	enable := c.Query("enable")
	key := c.Query("key")

	var err error
	var r *rule.RelayRule
	var syncSuccess bool

	if enable == "true" {
		r, syncSuccess, err = rule.EnableRelayRuleByKey(key)
	} else {
		r, syncSuccess, err = rule.DisableRelayRuleByKey(key)
	}

	if err != nil {
		c.JSON(http.StatusOK, gin.H{"ret": 1, "msg": fmt.Sprintf("开关规则出错:%s", err.Error())})
		return
	}

	log.Printf("[%s] relayRule[%s][%s]", enable, r.Name, r.MainConfigure)
	syncRes := ""
	if !syncSuccess {
		syncRes = "同步规则状态到配置文件出错"
	}

	c.JSON(http.StatusOK, gin.H{"ret": 0, "msg": "", "syncres": syncRes})
}

func alterrule(c *gin.Context) {

	var requestRule rule.RelayRule
	err := c.BindJSON(&requestRule)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"ret": 1, "msg": fmt.Sprintf("修改请求解析出错:%s", err.Error())})
		return
	}

	//fmt.Printf("balance:%v\n", requestRule.BalanceTargetAddressList)

	preConfigureStr := requestRule.MainConfigure
	configureStr := requestRule.CreateMainConfigure()
	// configureStr := fmt.Sprintf("%s@%s:%sto%s:%s",
	// 	requestRule.RelayType,
	// 	requestRule.ListenIP, requestRule.ListenPorts,
	// 	requestRule.TargetIP, requestRule.TargetPorts)

	r, err := rule.CreateRuleByConfigureAndOptions(requestRule.Name, configureStr, requestRule.Options)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"ret": 1, "msg": fmt.Sprintf("修改转发规则[%s]时出错:%s", preConfigureStr, err.Error())})
		return
	}

	syncSuccess, err := rule.AlterRuleInGlobalRuleListByKey(preConfigureStr, r)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"ret": 1, "msg": fmt.Sprintf("修改转发规则[%s]时出错:%s", preConfigureStr, err.Error())})
		return
	}

	r, _, err = rule.EnableRelayRuleByKey(r.MainConfigure)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"ret": 0, "msg": fmt.Sprintf("修改转发规则成功,但启用规则时出错:%s", err.Error())})
		return
	}
	log.Printf("修改转发规则[%s][%s]成功", r.Name, r.MainConfigure)

	synsRes := ""

	if !syncSuccess {
		synsRes = "同步修改规则数据到配置文件出错"
	}

	c.JSON(http.StatusOK, gin.H{"ret": 0, "msg": "修改转发规则成功", "syncres": synsRes})
}

func deleterule(c *gin.Context) {
	ruleKey := c.Query("rule")

	rule.DisableRelayRuleByKey(ruleKey)

	syncSuccess, err := rule.DeleteGlobalRuleByKey(ruleKey)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"ret": 1, "msg": fmt.Sprintf("删除转发规则出错:%s", err.Error())})
		return
	}

	syncRes := ""
	if !syncSuccess {
		syncRes = "同步规则信息到配置文件出错"
	}

	c.JSON(http.StatusOK, gin.H{"ret": 0, "msg": "删除成功", "syncres": syncRes})
}

func addrule(c *gin.Context) {
	var requestRule rule.RelayRule
	err := c.BindJSON(&requestRule)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"ret": 1, "msg": fmt.Sprintf("请求解析出错:%s", err.Error())})
		return
	}

	// configureStr := fmt.Sprintf("%s@%s:%sto%s:%s",
	// 	requestRule.RelayType,
	// 	requestRule.ListenIP, requestRule.ListenPorts,
	// 	requestRule.TargetIP, requestRule.TargetPorts)
	configureStr := requestRule.CreateMainConfigure()

	r, err := rule.CreateRuleByConfigureAndOptions(requestRule.Name, configureStr, requestRule.Options)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"ret": 1, "msg": fmt.Sprintf("创建转发规则出错:%s", err.Error())})
		return
	}

	synsRes, err := rule.AddRuleToGlobalRuleList(true, *r)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"ret": 1, "msg": fmt.Sprintf("添加转发规则出错:%s", err.Error())})
		return
	}

	r, _, err = rule.EnableRelayRuleByKey(r.MainConfigure)

	if err != nil {
		c.JSON(http.StatusOK, gin.H{"ret": 0, "msg": fmt.Sprintf("启用规则出错:%s", err.Error())})
		return
	}
	log.Printf("添加转发规则[%s][%s]成功", r.Name, r.MainConfigure)

	if synsRes != "" {
		synsRes = "保存配置文件出错,请检查配置文件设置"
	}

	c.JSON(http.StatusOK, gin.H{"ret": 0, "msg": "添加规则并启用成功", "syncres": synsRes})
}

func rulelist(c *gin.Context) {
	ruleList, proxyListInfoMap := rule.GetRelayRuleList()
	type ruleItem struct {
		Name                     string                    `json:"Name"`
		MainConfigure            string                    `json:"Mainconfigure"`
		RelayType                string                    `json:"RelayType"`
		ListenIP                 string                    `json:"ListenIP"`
		ListenPorts              string                    `json:"ListenPorts"`
		TargetIP                 string                    `json:"TargetIP"`
		TargetPorts              string                    `json:"TargetPorts"`
		BalanceTargetAddressList []string                  `json:"BalanceTargetAddressList"`
		Options                  base.RelayRuleOptions     `json:"Options"`
		SubRuleList              []rule.SubRelayRule       `json:"SubRuleList"`
		From                     string                    `json:"From"`
		IsEnable                 bool                      `json:"Enable"`
		ProxyList                []rule.RelayRuleProxyInfo `json:"ProxyList"`
	}

	//proxyListInfoMap[(*ruleList)[i].MainConfigure]
	var data []ruleItem

	for i := range *ruleList {
		item := ruleItem{
			Name:                     (*ruleList)[i].Name,
			MainConfigure:            (*ruleList)[i].MainConfigure,
			RelayType:                (*ruleList)[i].RelayType,
			ListenIP:                 (*ruleList)[i].ListenIP,
			ListenPorts:              (*ruleList)[i].ListenPorts,
			TargetIP:                 (*ruleList)[i].TargetIP,
			TargetPorts:              (*ruleList)[i].TargetPorts,
			Options:                  (*ruleList)[i].Options,
			SubRuleList:              (*ruleList)[i].SubRuleList,
			From:                     (*ruleList)[i].From,
			IsEnable:                 (*ruleList)[i].IsEnable,
			ProxyList:                proxyListInfoMap[(*ruleList)[i].MainConfigure],
			BalanceTargetAddressList: (*ruleList)[i].BalanceTargetAddressList,
		}
		data = append(data, item)
	}
	c.JSON(http.StatusOK, gin.H{"ret": 0, "data": data})

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

	respMap := make(map[string]interface{})
	respMap["totleMem"] = formatFileSize(v.Total)
	respMap["usedMem"] = formatFileSize(v.Used)
	respMap["unusedMem"] = formatFileSize(v.Free)
	respMap["currentProcessUsedCPU"] = fmt.Sprintf("%.2f%%", GetCurrentProcessCPUPrecent())
	respMap["goroutine"] = fmt.Sprintf("%d", runtime.NumGoroutine())
	respMap["processUsedMem"] = formatFileSize(currentProcessMem)
	respMap["currentConnections"] = fmt.Sprintf("%d", base.GetGlobalConnections())
	respMap["maxConnections"] = fmt.Sprintf("%d", base.GetGlobalMaxConnections())
	respMap["usedCPU"] = fmt.Sprintf("%.2f%%", GetCpuPercent())
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

//跨域访问：cross  origin resource share
func CrosHandler() gin.HandlerFunc {
	return func(context *gin.Context) {
		method := context.Request.Method
		//context.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		origin := context.Request.Header.Get("Origin")
		context.Header("Access-Control-Allow-Origin", origin) // 设置允许访问所有域
		context.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE,UPDATE")
		context.Header("Access-Control-Allow-Headers", "Authorization, Content-Length, X-CSRF-Token, Token,session,X_Requested_With,Accept, Origin, Host, Connection, Accept-Encoding, Accept-Language,DNT, X-CustomHeader, Keep-Alive, User-Agent, X-Requested-With, If-Modified-Since, Cache-Control, Content-Type, Pragma,token,openid,opentoken")
		//context.Header("Access-Control-Expose-Headers", "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers,Cache-Control,Content-Language,Content-Type,Expires,Last-Modified,Pragma,FooBar")
		//context.Header("Access-Control-Allow-Methods", "*")
		//context.Header("Access-Control-Allow-Headers", "*")
		context.Header("Access-Control-Expose-Headers", "*")
		context.Header("Access-Control-Allow-Credentials", "true")

		context.Header("Access-Control-Max-Age", "172800")
		//context.Header("Access-Control-Allow-Credentials", "false")
		//context.Set("content-type", "application/json")

		if method == "OPTIONS" {
			context.JSON(http.StatusOK, gin.H{
				"ret": 0,
			})
		}
		//处理请求
		context.Next()
	}
}

//------------------------------------------------------------------------------------------------------------------

func BasicAuth() gin.HandlerFunc {
	return gin.BasicAuthForRealm(config.GetAuthAccount(), "")
}

func formatFileSize(fileSize uint64) (size string) {
	switch {
	case fileSize < 1024:
		return fmt.Sprintf("%.2fB", float64(fileSize)/float64(1))
	case fileSize < (1024 * 1024):
		return fmt.Sprintf("%.2fKB", float64(fileSize)/float64(1024))
	case fileSize < (1024 * 1024 * 1024):
		return fmt.Sprintf("%.2fMB", float64(fileSize)/float64(1024*1024))
	case fileSize < (1024 * 1024 * 1024 * 1024):
		return fmt.Sprintf("%.2fGB", float64(fileSize)/float64(1024*1024*1024))
	case fileSize < (1024 * 1024 * 1024 * 1024 * 1024):
		return fmt.Sprintf("%.2fTB", float64(fileSize)/float64(1024*1024*1024*1024))
	default:
		return fmt.Sprintf("%.2fEB", float64(fileSize)/float64(1024*1024*1024*1024*1024))
	}

}

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

//***********************
//basicAuth

type authPair struct {
	value string
	user  string
}

type authPairs []authPair

func processAccounts(accounts gin.Accounts) authPairs {
	length := len(accounts)
	assert1(length > 0, "Empty list of authorized credentials")
	pairs := make(authPairs, 0, length)
	for user, password := range accounts {
		assert1(user != "", "User can not be empty")
		value := authorizationHeader(user, password)
		pairs = append(pairs, authPair{
			value: value,
			user:  user,
		})
	}
	return pairs
}

func authorizationHeader(user, password string) string {
	base := user + ":" + password
	return "Basic " + base64.StdEncoding.EncodeToString(StringToBytes(base))
}
func assert1(guard bool, text string) {
	if !guard {
		panic(text)
	}
}

func StringToBytes(s string) []byte {
	return *(*[]byte)(unsafe.Pointer(
		&struct {
			string
			Cap int
		}{s, len(s)},
	))
}

func (a authPairs) searchCredential(authValue string) (string, bool) {
	if authValue == "" {
		return "", false
	}
	for _, pair := range a {
		if subtle.ConstantTimeCompare(StringToBytes(pair.value), StringToBytes(authValue)) == 1 {
			return pair.user, true
		}
	}
	return "", false
}
