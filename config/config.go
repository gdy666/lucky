//Copyright 2022 gdy, 272288813@qq.com
package config

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"strings"
	"sync"

	"github.com/gdy666/lucky/base"
	"github.com/gdy666/lucky/thirdlib/gdylib/fileutils"
	"github.com/gdy666/lucky/thirdlib/gdylib/stringsp"
)

const defaultAdminAccount = "666"
const defaultAdminPassword = "666"
const defaultAdminListenPort = 16601

var runMode = "prod"
var version = "0.0.0"

func SetVersion(v string) {
	version = v
}

func GetVersion() string {
	return version
}

var loginRandomkey = ""

func GetLoginRandomKey() string {
	return loginRandomkey
}

func FlushLoginRandomKey() {
	loginRandomkey = stringsp.GetRandomString(16)
}

type ConfigureRelayRule struct {
	Name         string                `json:"Name"`
	Configurestr string                `json:"Configurestr"`
	Enable       bool                  `json:"Enable"`
	Options      base.RelayRuleOptions `json:"Options"`
}

type BaseConfigure struct {
	AdminWebListenPort   int    `json:"AdminWebListenPort"`   //管理后台端口
	ProxyCountLimit      int64  `json:"ProxyCountLimit"`      //全局代理数量限制
	AdminAccount         string `json:"AdminAccount"`         //登录账号
	AdminPassword        string `json:"AdminPassword"`        //登录密码
	AllowInternetaccess  bool   `json:"AllowInternetaccess"`  //允许外网访问
	GlobalMaxConnections int64  `json:"GlobalMaxConnections"` //全局最大连接数
}

type ProgramConfigure struct {
	BaseConfigure      BaseConfigure        `json:"BaseConfigure"`
	RelayRuleList      []ConfigureRelayRule `json:"RelayRuleList"`
	WhiteListConfigure WhiteListConfigure   `json:"WhiteListConfigure"`
	BlackListConfigure BlackListConfigure   `json:"BlackListConfigure"`
	DDNSConfigure      DDNSConfigure        `json:"DDNSConfigure"` //DDNS 参数设置
	DDNSTaskList       []DDNSTask           `json:"DDNSTaskList"`
}

var programConfigureMutex sync.RWMutex
var programConfigure *ProgramConfigure
var configurePath string

//var readConfigureFileOnce sync.Once
var checkConfigureFileOnce sync.Once
var configureFileSign int8 = -1

// func GetConfigMutex() *sync.RWMutex {
// 	return &programConfigureMutex
// }

func GetAuthAccount() map[string]string {
	programConfigureMutex.RLock()
	defer programConfigureMutex.RUnlock()
	accountInfo := make(map[string]string)
	accountInfo[programConfigure.BaseConfigure.AdminAccount] = programConfigure.BaseConfigure.AdminPassword
	return accountInfo
}

func SetConfigRuleList(ruleList *[]ConfigureRelayRule) {
	programConfigureMutex.Lock()
	defer programConfigureMutex.Unlock()
	programConfigure.RelayRuleList = *ruleList
}

func GetRunMode() string {
	return runMode
}

func SetRunMode(mode string) {
	runMode = mode
}

func GetConfig() *ProgramConfigure {
	return programConfigure
}

func GetBaseConfigure() BaseConfigure {
	programConfigureMutex.RLock()
	defer programConfigureMutex.RUnlock()
	baseConf := programConfigure.BaseConfigure
	return baseConf
}

func GetDDNSConfigure() DDNSConfigure {
	programConfigureMutex.RLock()
	defer programConfigureMutex.RUnlock()
	conf := programConfigure.DDNSConfigure
	return conf
}

//保存基础配置
func SetBaseConfigure(conf *BaseConfigure) error {
	programConfigureMutex.Lock()
	defer programConfigureMutex.Unlock()
	programConfigure.BaseConfigure = *conf

	base.SetGlobalMaxConnections(conf.GlobalMaxConnections)
	base.SetGlobalMaxProxyCount(conf.ProxyCountLimit)

	return Save()
}

func SetDDNSConfigure(conf *DDNSConfigure) error {
	programConfigureMutex.Lock()
	defer programConfigureMutex.Unlock()

	if conf.Intervals < 30 {
		conf.Intervals = 30
	}

	if conf.Intervals > 3600 {
		conf.Intervals = 3600
	}

	if conf.FirstCheckDelay < 0 {
		conf.FirstCheckDelay = 0
	}

	if conf.FirstCheckDelay > 3600 {
		conf.FirstCheckDelay = 3600
	}

	programConfigure.DDNSConfigure = *conf
	return Save()
}

func Read(filePath string) (err error) {

	pc, err := readProgramConfigure(filePath)
	if err != nil {
		return err
	}

	if pc.BaseConfigure.GlobalMaxConnections <= 0 {
		pc.BaseConfigure.GlobalMaxConnections = base.DEFAULT_GLOBAL_MAX_CONNECTIONS
	}

	if pc.BaseConfigure.ProxyCountLimit <= 0 {
		pc.BaseConfigure.ProxyCountLimit = base.DEFAULT_MAX_PROXY_COUNT
	}

	if pc.BaseConfigure.AdminWebListenPort <= 0 {
		pc.BaseConfigure.AdminWebListenPort = 16601
	}

	programConfigure = pc

	return nil
}

func LoadDefault(proxyCountLimit int64,
	adminWebListenPort int,
	globalMaxConnections int64) {

	programConfigure = loadDefaultConfigure(proxyCountLimit, adminWebListenPort, globalMaxConnections)
}

func Save() (err error) {
	//log.Printf("Save配置\n")
	if configureFileSign == 0 {
		return fmt.Errorf("配置文件不存在,不作保存")
	}
	defer func() {
		checkConfigureFileOnce.Do(func() {
			if err == nil {
				configureFileSign = 1
			} else {
				configureFileSign = 0
			}
		})

	}()

	err = saveProgramConfig(programConfigure, configurePath)
	return
}

//------------------------------------------------------------------------------------

func readProgramConfigure(filePath string) (conf *ProgramConfigure, err error) {
	if filePath == "" {
		return nil, fmt.Errorf("未指定配置文件路径")
	}

	if !strings.HasPrefix(filePath, "/") {
		filePath = fmt.Sprintf("%s/%s", fileutils.GetCurrentDirectory(), filePath)
	}

	configurePath = filePath

	//fmt.Printf("filePath:%s\n", configurePath)

	fileContent, err := fileutils.ReadTextFromFile(configurePath)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件出错:%s", err.Error())
	}
	var pc ProgramConfigure

	err = json.Unmarshal([]byte(fileContent), &pc)
	if err != nil {
		log.Fatalf("解析配置文件出错:%s", err.Error())
		return nil, fmt.Errorf("解析配置文件出错:%s", err.Error())
	}

	return &pc, nil
}

func loadDefaultConfigure(
	proxyCountLimit int64,
	adminWebListenPort int,
	globalMaxConnections int64) *ProgramConfigure {

	baseConfigure := BaseConfigure{AdminWebListenPort: adminWebListenPort,
		GlobalMaxConnections: globalMaxConnections,
		AdminAccount:         defaultAdminAccount,
		AdminPassword:        defaultAdminPassword,
		ProxyCountLimit:      proxyCountLimit,
		AllowInternetaccess:  false}

	whiteListConfigure := WhiteListConfigure{BaseConfigure: WhiteListBaseConfigure{ActivelifeDuration: 36, BasicAccount: defaultAdminAccount, BasicPassword: defaultAdminPassword}}

	var pc ProgramConfigure
	pc.BaseConfigure = baseConfigure
	pc.WhiteListConfigure = whiteListConfigure

	if pc.BaseConfigure.GlobalMaxConnections <= 0 {
		pc.BaseConfigure.GlobalMaxConnections = base.DEFAULT_GLOBAL_MAX_CONNECTIONS
	}

	if pc.BaseConfigure.ProxyCountLimit <= 0 {
		pc.BaseConfigure.ProxyCountLimit = base.DEFAULT_MAX_PROXY_COUNT
	}

	if pc.BaseConfigure.AdminWebListenPort <= 0 {
		pc.BaseConfigure.AdminWebListenPort = defaultAdminListenPort
	}

	if pc.DDNSConfigure.Intervals < 30 {
		pc.DDNSConfigure.Intervals = 30
	}

	if pc.DDNSConfigure.FirstCheckDelay <= 0 {
		pc.DDNSConfigure.FirstCheckDelay = 0
	}

	return &pc
}

func saveProgramConfig(programConfigure *ProgramConfigure, filePath string) error {
	resBytes, err := json.MarshalIndent(*programConfigure, "", "\t")
	if err != nil {
		return fmt.Errorf("json.Marshal:%s", err.Error())
	}
	return fileutils.SaveTextToFile(string(resBytes), filePath)
}

func CheckTCPPortAvalid(port int) bool {
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return false
	}
	ln.Close()
	return true
}
