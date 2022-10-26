// Copyright 2022 gdy, 272288813@qq.com
package config

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"runtime"
	"strings"
	"sync"

	"github.com/gdy666/lucky/socketproxy"
	"github.com/gdy666/lucky/thirdlib/gdylib/fileutils"
	"github.com/gdy666/lucky/thirdlib/gdylib/stringsp"
)

const defaultAdminAccount = "666"
const defaultAdminPassword = "666"
const defaultAdminListenPort = 16601
const defaultLogSize = 2048
const minLogSize = 1024
const maxLogSize = 40960

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

type BaseConfigure struct {
	AdminWebListenPort      int  `json:"AdminWebListenPort"`      //管理后台端口
	AdminWebListenTLS       bool `json:"AdminWebListenTLS"`       //启用TLS监听端口
	AdminWebListenHttpsPort int  `json:"AdminWebListenHttpsPort"` //管理后台Https端口

	//ProxyCountLimit      int64  `json:"ProxyCountLimit"`      //全局代理数量限制
	AdminAccount        string `json:"AdminAccount"`        //登录账号
	AdminPassword       string `json:"AdminPassword"`       //登录密码
	AllowInternetaccess bool   `json:"AllowInternetaccess"` //允许外网访问
	//GlobalMaxConnections int64  `json:"GlobalMaxConnections"` //全局最大连接数
	LogMaxSize int `json:"LogMaxSize"` //日志记录最大条数
}

type ProgramConfigure struct {
	BaseConfigure         BaseConfigure         `json:"BaseConfigure"`
	WhiteListConfigure    WhiteListConfigure    `json:"WhiteListConfigure"`
	BlackListConfigure    BlackListConfigure    `json:"BlackListConfigure"`
	DDNSConfigure         DDNSConfigure         `json:"DDNSConfigure"`         //DDNS 参数设置
	DDNSTaskList          []DDNSTask            `json:"DDNSTaskList"`          //DDNS任务列表
	ReverseProxyRuleList  []ReverseProxyRule    `json:"ReverseProxyRuleList"`  //反向代理规则列表
	SSLCertficateList     []SSLCertficate       `json:"SSLCertficateList"`     //SSL证书列表
	PortForwardsRuleList  []PortForwardsRule    `json:"PortForwardsRuleList"`  //端口转发规则列表
	PortForwardsConfigure PortForwardsConfigure `json:"PortForwardsConfigure"` //端口转发设置
	WOLDeviceList         []WOLDevice           `json:"WOLDeviceList"`         //网络唤醒设备列表
}

var programConfigureMutex sync.RWMutex
var programConfigure *ProgramConfigure
var configurePath string

// var readConfigureFileOnce sync.Once
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

func GetRunMode() string {
	return runMode
}

func SetRunMode(mode string) {
	runMode = mode
}

func SetConfig(p *ProgramConfigure) error {
	programConfigureMutex.Lock()
	defer programConfigureMutex.Unlock()
	programConfigure = p
	return Save()
}

func GetConfig() *ProgramConfigure {
	programConfigureMutex.RLock()
	defer programConfigureMutex.RUnlock()
	conf := *programConfigure
	return &conf
}

func GetConfigureBytes() []byte {
	programConfigureMutex.RLock()
	defer programConfigureMutex.RUnlock()
	if programConfigure == nil {
		return []byte("{}")
	}
	//JSON.Pars
	res, err := json.MarshalIndent(*programConfigure, "", "\t")
	if err != nil {
		return []byte("{}")
	}
	return res
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

func GetPortForwardsConfigure() PortForwardsConfigure {
	programConfigureMutex.RLock()
	defer programConfigureMutex.RUnlock()
	conf := programConfigure.PortForwardsConfigure
	return conf
}

func SetPortForwardsConfigure(conf *PortForwardsConfigure) error {
	programConfigureMutex.Lock()
	defer programConfigureMutex.Unlock()

	if conf.PortForwardsLimit < 0 {
		conf.PortForwardsLimit = 0
	} else if conf.PortForwardsLimit > 1024 {
		conf.PortForwardsLimit = 1024
	}

	if conf.TCPPortforwardMaxConnections < 0 {
		conf.TCPPortforwardMaxConnections = 0
	} else if conf.TCPPortforwardMaxConnections > 4096 {
		conf.TCPPortforwardMaxConnections = 4096
	}

	if conf.UDPReadTargetDataMaxgoroutineCount < 0 {
		conf.UDPReadTargetDataMaxgoroutineCount = 0
	} else if conf.UDPReadTargetDataMaxgoroutineCount > 4096 {
		conf.UDPReadTargetDataMaxgoroutineCount = 4096
	}

	programConfigure.PortForwardsConfigure = *conf

	socketproxy.SetGlobalMaxPortForwardsCountLimit(conf.PortForwardsLimit)
	socketproxy.SetGlobalTCPPortforwardMaxConnections(conf.TCPPortforwardMaxConnections)
	socketproxy.SetGlobalUDPReadTargetDataMaxgoroutineCountLimit(conf.UDPReadTargetDataMaxgoroutineCount)
	return Save()
}

// 保存基础配置
func SetBaseConfigure(conf *BaseConfigure) error {
	programConfigureMutex.Lock()
	defer programConfigureMutex.Unlock()
	programConfigure.BaseConfigure = *conf

	//socketproxy.SetGlobalMaxConnections(conf.GlobalMaxConnections)
	//socketproxy.SetGlobalMaxPortForwardsCount(conf.ProxyCountLimit)

	if conf.LogMaxSize < minLogSize {
		conf.LogMaxSize = minLogSize
	} else if conf.LogMaxSize > maxLogSize {
		conf.LogMaxSize = maxLogSize
	}

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

	if runtime.GOOS == "windows" && filePath == "" {
		filePath = "lucky.conf"
		log.Printf("未指定配置文件路径,使用默认路径lucky所在位置,默认配置文件名lucky.conf")
	}

	pc, err := readProgramConfigure(filePath)
	if err != nil {
		return err
	}

	if pc.PortForwardsConfigure.PortForwardsLimit <= 0 {
		pc.PortForwardsConfigure.PortForwardsLimit = socketproxy.DEFAULT_MAX_PORTFORWARDS_LIMIT
	}
	if pc.PortForwardsConfigure.TCPPortforwardMaxConnections <= 0 {
		pc.PortForwardsConfigure.TCPPortforwardMaxConnections = socketproxy.DEFAULT_GLOBAL_MAX_CONNECTIONS
	}

	if pc.PortForwardsConfigure.UDPReadTargetDataMaxgoroutineCount <= 0 {
		pc.PortForwardsConfigure.UDPReadTargetDataMaxgoroutineCount = socketproxy.DEFAULT_GLOBAL_UDPReadTargetDataMaxgoroutineCount
	}

	socketproxy.SetGlobalMaxPortForwardsCountLimit(pc.PortForwardsConfigure.PortForwardsLimit)
	socketproxy.SetGlobalTCPPortforwardMaxConnections(pc.PortForwardsConfigure.TCPPortforwardMaxConnections)
	socketproxy.SetGlobalUDPReadTargetDataMaxgoroutineCountLimit(pc.PortForwardsConfigure.UDPReadTargetDataMaxgoroutineCount)

	if pc.BaseConfigure.AdminWebListenPort <= 0 {
		pc.BaseConfigure.AdminWebListenPort = 16601
	}

	if pc.BaseConfigure.AdminWebListenHttpsPort <= 0 {
		pc.BaseConfigure.AdminWebListenHttpsPort = 16626
	}

	if pc.BaseConfigure.LogMaxSize < minLogSize {
		pc.BaseConfigure.LogMaxSize = minLogSize
	} else if pc.BaseConfigure.LogMaxSize > maxLogSize {
		pc.BaseConfigure.LogMaxSize = maxLogSize
	}

	programConfigure = pc

	return nil
}

func LoadDefault(adminWebListenPort int) {
	programConfigure = loadDefaultConfigure(adminWebListenPort)
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
	adminWebListenPort int) *ProgramConfigure {

	baseConfigure := BaseConfigure{AdminWebListenPort: adminWebListenPort,
		AdminAccount:        defaultAdminAccount,
		AdminPassword:       defaultAdminPassword,
		AllowInternetaccess: false,
		LogMaxSize:          defaultLogSize}

	whiteListConfigure := WhiteListConfigure{BaseConfigure: WhiteListBaseConfigure{ActivelifeDuration: 36, BasicAccount: defaultAdminAccount, BasicPassword: defaultAdminPassword}}

	var pc ProgramConfigure
	pc.BaseConfigure = baseConfigure
	pc.WhiteListConfigure = whiteListConfigure

	if pc.PortForwardsConfigure.PortForwardsLimit <= 0 {
		pc.PortForwardsConfigure.PortForwardsLimit = socketproxy.DEFAULT_MAX_PORTFORWARDS_LIMIT
	}
	socketproxy.SetGlobalMaxPortForwardsCountLimit(pc.PortForwardsConfigure.PortForwardsLimit)

	if pc.PortForwardsConfigure.TCPPortforwardMaxConnections <= 0 {
		pc.PortForwardsConfigure.TCPPortforwardMaxConnections = socketproxy.TCPUDP_DEFAULT_SINGLE_PROXY_MAX_CONNECTIONS
	}
	socketproxy.SetGlobalTCPPortforwardMaxConnections(pc.PortForwardsConfigure.TCPPortforwardMaxConnections)

	if pc.PortForwardsConfigure.UDPReadTargetDataMaxgoroutineCount <= 0 {
		pc.PortForwardsConfigure.UDPReadTargetDataMaxgoroutineCount = socketproxy.DEFAULT_GLOBAL_UDPReadTargetDataMaxgoroutineCount
	}

	socketproxy.SetGlobalUDPReadTargetDataMaxgoroutineCountLimit(pc.PortForwardsConfigure.UDPReadTargetDataMaxgoroutineCount)

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
