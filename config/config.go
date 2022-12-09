// Copyright 2022 gdy, 272288813@qq.com
package config

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"runtime"
	"strings"
	"sync"

	ddnsconf "github.com/gdy666/lucky/module/ddns/conf"
	portforwardconf "github.com/gdy666/lucky/module/portforward/conf"
	"github.com/gdy666/lucky/module/portforward/socketproxy"
	reverseproxyconf "github.com/gdy666/lucky/module/reverseproxy/conf"
	safeconf "github.com/gdy666/lucky/module/safe/conf"
	sslconf "github.com/gdy666/lucky/module/sslcertficate/conf"
	wolconf "github.com/gdy666/lucky/module/wol/conf"

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

	HttpClientSecureVerify bool `json:"HttpClientSecureVerify"`
	HttpClientTimeout      int  `json:"HttpClientTimeout"`
}

type ProgramConfigure struct {
	BaseConfigure         BaseConfigure                         `json:"BaseConfigure"`
	WhiteListConfigure    safeconf.WhiteListConfigure           `json:"WhiteListConfigure"`
	BlackListConfigure    safeconf.BlackListConfigure           `json:"BlackListConfigure"`
	DDNSConfigure         ddnsconf.DDNSConfigure                `json:"DDNSConfigure"`         //DDNS 参数设置
	DDNSTaskList          []ddnsconf.DDNSTask                   `json:"DDNSTaskList"`          //DDNS任务列表
	ReverseProxyRuleList  []reverseproxyconf.ReverseProxyRule   `json:"ReverseProxyRuleList"`  //反向代理规则列表
	SSLCertficateList     []sslconf.SSLCertficate               `json:"SSLCertficateList"`     //SSL证书列表
	PortForwardsRuleList  []portforwardconf.PortForwardsRule    `json:"PortForwardsRuleList"`  //端口转发规则列表
	PortForwardsConfigure portforwardconf.PortForwardsConfigure `json:"PortForwardsConfigure"` //端口转发设置
	WOLDeviceList         []wolconf.WOLDevice                   `json:"WOLDeviceList"`         //网络唤醒设备列表
	WOLServiceConfigure   wolconf.WOLServiceConfigure           `json:"WOLServiceConfigure"`   //网络唤醒客户端设置
}

var ConfigureMutex sync.RWMutex
var Configure *ProgramConfigure
var configurePath string

// var readConfigureFileOnce sync.Once
var checkConfigureFileOnce sync.Once
var configureFileSign int8 = -1

// func GetConfigMutex() *sync.RWMutex {
// 	return &programConfigureMutex
// }

func GetAuthAccount() map[string]string {
	ConfigureMutex.RLock()
	defer ConfigureMutex.RUnlock()
	accountInfo := make(map[string]string)
	accountInfo[Configure.BaseConfigure.AdminAccount] = Configure.BaseConfigure.AdminPassword
	return accountInfo
}

func GetRunMode() string {
	return runMode
}

func SetRunMode(mode string) {
	runMode = mode
}

func SetConfig(p *ProgramConfigure) error {
	ConfigureMutex.Lock()
	defer ConfigureMutex.Unlock()
	Configure = p
	return Save()
}

func GetConfig() *ProgramConfigure {
	ConfigureMutex.RLock()
	defer ConfigureMutex.RUnlock()
	conf := *Configure
	return &conf
}

func GetConfigureBytes() []byte {
	ConfigureMutex.RLock()
	defer ConfigureMutex.RUnlock()
	if Configure == nil {
		return []byte("{}")
	}
	//JSON.Pars
	res, err := json.MarshalIndent(*Configure, "", "\t")
	if err != nil {
		return []byte("{}")
	}
	return res
}

func GetBaseConfigure() BaseConfigure {
	ConfigureMutex.RLock()
	defer ConfigureMutex.RUnlock()
	baseConf := Configure.BaseConfigure
	return baseConf
}

// 保存基础配置
func SetBaseConfigure(conf *BaseConfigure) error {
	ConfigureMutex.Lock()
	defer ConfigureMutex.Unlock()
	Configure.BaseConfigure = *conf

	//socketproxy.SetGlobalMaxConnections(conf.GlobalMaxConnections)
	//socketproxy.SetGlobalMaxPortForwardsCount(conf.ProxyCountLimit)

	if conf.LogMaxSize < minLogSize {
		conf.LogMaxSize = minLogSize
	} else if conf.LogMaxSize > maxLogSize {
		conf.LogMaxSize = maxLogSize
	}

	if conf.HttpClientTimeout <= 0 {
		conf.HttpClientTimeout = 1
	} else if conf.HttpClientTimeout > 60 {
		conf.HttpClientTimeout = 60
	}

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

	checkConfigue(pc)
	Configure = pc

	return nil
}

func checkConfigue(pc *ProgramConfigure) {
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

	if pc.BaseConfigure.HttpClientTimeout <= 0 {
		pc.BaseConfigure.HttpClientTimeout = 20
	} else if pc.BaseConfigure.HttpClientTimeout > 60 {
		pc.BaseConfigure.HttpClientTimeout = 60
	}

	if pc.WOLServiceConfigure.Client.Port <= 0 {
		pc.WOLServiceConfigure.Client.Port = 9
	}

	if pc.WOLServiceConfigure.Client.Repeat <= 0 {
		pc.WOLServiceConfigure.Client.Repeat = 5
	}

	if pc.WOLServiceConfigure.Client.DeviceName == "" {
		hostname, _ := os.Hostname()
		pc.WOLServiceConfigure.Client.DeviceName = hostname
	}

	if pc.WOLServiceConfigure.Client.PowerOffCMD == "" {
		switch runtime.GOOS {
		case "linux":
			pc.WOLServiceConfigure.Client.PowerOffCMD = "poweroff"
		case "windows":
			pc.WOLServiceConfigure.Client.PowerOffCMD = "Shutdown /s /t 0"
		default:
			pc.WOLServiceConfigure.Client.PowerOffCMD = ""
		}
	}

	if pc.WOLServiceConfigure.Server.Token == "" {
		pc.WOLServiceConfigure.Server.Token = "666666"
	}
}

func LoadDefault(adminWebListenPort int) {
	Configure = loadDefaultConfigure(adminWebListenPort)
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

	err = saveProgramConfig(Configure, configurePath)
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

	whiteListConfigure := safeconf.WhiteListConfigure{BaseConfigure: safeconf.WhiteListBaseConfigure{ActivelifeDuration: 36, BasicAccount: defaultAdminAccount, BasicPassword: defaultAdminPassword}}

	var pc ProgramConfigure
	pc.BaseConfigure = baseConfigure
	pc.WhiteListConfigure = whiteListConfigure

	checkConfigue(&pc)

	// if pc.PortForwardsConfigure.PortForwardsLimit <= 0 {
	// 	pc.PortForwardsConfigure.PortForwardsLimit = socketproxy.DEFAULT_MAX_PORTFORWARDS_LIMIT
	// }
	// socketproxy.SetGlobalMaxPortForwardsCountLimit(pc.PortForwardsConfigure.PortForwardsLimit)

	// if pc.PortForwardsConfigure.TCPPortforwardMaxConnections <= 0 {
	// 	pc.PortForwardsConfigure.TCPPortforwardMaxConnections = socketproxy.TCPUDP_DEFAULT_SINGLE_PROXY_MAX_CONNECTIONS
	// }
	// socketproxy.SetGlobalTCPPortforwardMaxConnections(pc.PortForwardsConfigure.TCPPortforwardMaxConnections)

	// if pc.PortForwardsConfigure.UDPReadTargetDataMaxgoroutineCount <= 0 {
	// 	pc.PortForwardsConfigure.UDPReadTargetDataMaxgoroutineCount = socketproxy.DEFAULT_GLOBAL_UDPReadTargetDataMaxgoroutineCount
	// }

	// socketproxy.SetGlobalUDPReadTargetDataMaxgoroutineCountLimit(pc.PortForwardsConfigure.UDPReadTargetDataMaxgoroutineCount)

	// if pc.BaseConfigure.AdminWebListenPort <= 0 {
	// 	pc.BaseConfigure.AdminWebListenPort = defaultAdminListenPort
	// }

	// if pc.DDNSConfigure.Intervals < 30 {
	// 	pc.DDNSConfigure.Intervals = 30
	// }

	// if pc.DDNSConfigure.FirstCheckDelay <= 0 {
	// 	pc.DDNSConfigure.FirstCheckDelay = 0
	// }

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
