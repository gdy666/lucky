package config

import (
	"fmt"
	"log"

	"github.com/gdy666/lucky/thirdlib/gdylib/stringsp"
)

type DDNSConfigure struct {
	Enable                 bool `json:"Enable"`
	HttpClientSecureVerify bool `json:"HttpClientSecureVerify"`
	FirstCheckDelay        int  `json:"FirstCheckDelay"` //首次检查延迟时间
	Intervals              int  `json:"Intervals"`
}

type DDNSTask struct {
	TaskName string `json:"TaskName"`
	TaskKey  string `json:"TaskKey"` //添加任务时随机生成,方便管理任务(修改删除)
	//规则类型 IPv4/IPv6
	TaskType string `json:"TaskType"`
	Enable   bool
	// 获取IP类型 url/netInterface

	GetType      string    `json:"GetType"`
	URL          []string  `json:"URL"`
	NetInterface string    `json:"NetInterface"`
	IPReg        string    `json:"IPReg"`
	Domains      []string  `json:"Domains"`
	DNS          DNSConfig `json:"DNS"`
	Webhook
	TTL               string `json:"TTL"`
	HttpClientTimeout int    `json:"HttpClientTimeout"`
}

type Webhook struct {
	WebhookEnable                             bool     `json:"WebhookEnable"`          //Webhook开关
	WebhookCallOnGetIPfail                    bool     `json:"WebhookCallOnGetIPfail"` //获取IP失败时触发Webhook 开关
	WebhookURL                                string   `json:"WebhookURL"`
	WebhookMethod                             string   `json:"WebhookMethod"`
	WebhookHeaders                            []string `json:"WebhookHeaders"`
	WebhookRequestBody                        string   `json:"WebhookRequestBody"`
	WebhookDisableCallbackSuccessContentCheck bool     `json:"WebhookDisableCallbackSuccessContentCheck"` //禁用成功调用返回检测
	WebhookSuccessContent                     []string `json:"WebhookSuccessContent"`                     //接口调用成功包含的内容
	WebhookProxy                              string   `json:"WebhookProxy"`                              //使用DNS代理设置  ""表示禁用，"dns"表示使用dns的代理设置
	WebhookProxyAddr                          string   `json:"WebhookProxyAddr"`                          //代理服务器IP
	WebhookProxyUser                          string   `json:"WebhookProxyUser"`                          //代理用户
	WebhookProxyPassword                      string   `json:"WebhookProxyPassword"`                      //代理密码
}

// DNSConfig DNS配置
type DNSConfig struct {
	// 名称。如：alidns,webhook
	Name                    string      `json:"Name"`
	ID                      string      `json:"ID"`
	Secret                  string      `json:"Secret"`
	ForceInterval           int         `json:"ForceInterval"`       //(秒)即使IP没有变化,到一定时间后依然强制更新或先DNS解析比较IP再更新
	ResolverDoaminCheck     bool        `json:"ResolverDoaminCheck"` //调用callback同步前先解析一次域名,如果IP相同就不同步
	DNSServerList           []string    `json:"DNSServerList"`       //DNS服务器列表
	CallAPINetwork          string      `json:"CallAPINetwork"`      //空代理tcp, tcp4,tcp6
	Callback                DNSCallback `json:"Callback"`
	HttpClientProxyType     string      `json:"HttpClientProxyType"`     //http client代理服务器设置
	HttpClientProxyAddr     string      `json:"HttpClientProxyAddr"`     //代理服务器IP
	HttpClientProxyUser     string      `json:"HttpClientProxyUser"`     //代理用户
	HttpClientProxyPassword string      `json:"HttpClientProxyPassword"` //代理密码
}

func (d *DNSConfig) GetCallAPINetwork() string {
	switch d.CallAPINetwork {
	case "tcp4", "tcp6":
		return d.CallAPINetwork
	default:
		return "tcp"
	}
}

type DNSCallback struct {
	URL                                string   `json:"URL"`    //请求地址
	Method                             string   `json:"Method"` //请求方法
	Headers                            []string `json:"Headers"`
	RequestBody                        string   `json:"RequestBody"`
	Server                             string   `json:"Server"`                             //预设服务商
	DisableCallbackSuccessContentCheck bool     `json:"DisableCallbackSuccessContentCheck"` //禁用成功调用返回检测
	CallbackSuccessContent             []string `json:"CallbackSuccessContent"`             //接口调用成功包含内容
}

var checkIPv4URLList = []string{"https://4.ipw.cn", "http://v4.ip.zxinc.org/getip", "https://myip4.ipip.net", "https://www.taobao.com/help/getip.php", "https://ddns.oray.com/checkip", "https://ip.3322.net", "https://v4.myip.la"}
var checkIPv6URLList = []string{"https://6.ipw.cn", "https://ipv6.ddnspod.com", "http://v6.ip.zxinc.org/getip", "https://speed.neu6.edu.cn/getIP.php", "https://v6.ident.me", "https://v6.myip.la"}

var DefaultIPv6DNSServerList = []string{
	"[2001:4860:4860::8888]:53", //谷歌
	"[2001:4860:4860::8844]:53", //谷歌
	"[2606:4700:4700::64]:53",   //cloudflare
	"[2606:4700:4700::6400]:53", //cloudflare
	"[240C::6666]:53",           //下一代互联网北京研究中心
	"[240C::6644]:53",           //下一代互联网北京研究中心
	"[2402:4e00::]:53",          //dnspod
	//"[2400:3200::1]:53",         //阿里
	//		"[2400:3200:baba::1]:53",    //阿里
	"[240e:4c:4008::1]:53",  //中国电信
	"[240e:4c:4808::1]:53",  //中国电信
	"[2408:8899::8]:53",     //中国联通
	"[2408:8888::8]:53",     //中国联通
	"[2409:8088::a]:53",     //中国移动
	"[2409:8088::b]:53",     //中国移动
	"[2001:dc7:1000::1]:53", //CNNIC
	"[2400:da00::6666]:53",  //百度
}

var DefaultIPv4DNSServerList = []string{
	"1.1.1.1:53",
	"1.2.4.8:53",
	"8.8.8.8:53",
	"9.9.9.9:53",
	"8.8.4.4:53",
	"114.114.114.114:53",
	"223.5.5.5:53",
	"223.6.6.6:53",
	"101.226.4.6:53",
	"218.30.118.6:53",
	"119.28.28.28:53",
}

// func SetDDNSTaskIpCacheForceCompareByTaskKey(taskKey string, force bool) {
// 	programConfigureMutex.Lock()
// 	defer programConfigureMutex.Unlock()
// 	taskIndex := -1

// 	for i := range programConfigure.DDNSTaskList {
// 		if programConfigure.DDNSTaskList[i].TaskKey == taskKey {
// 			taskIndex = i
// 			break
// 		}
// 	}
// 	if taskIndex == -1 {
// 		return
// 	}
// 	programConfigure.DDNSTaskList[taskIndex].IpCache.ForceCompare = force
// }

func DDNSTaskListConfigureCheck() {
	programConfigureMutex.Lock()
	defer programConfigureMutex.Unlock()
	for i := range programConfigure.DDNSTaskList {
		if programConfigure.DDNSTaskList[i].DNS.ForceInterval < 60 {
			programConfigure.DDNSTaskList[i].DNS.ForceInterval = 60
		} else if programConfigure.DDNSTaskList[i].DNS.ForceInterval > 360000 {
			programConfigure.DDNSTaskList[i].DNS.ForceInterval = 360000
		}

		if programConfigure.DDNSTaskList[i].HttpClientTimeout < 3 {
			programConfigure.DDNSTaskList[i].HttpClientTimeout = 3
		} else if programConfigure.DDNSTaskList[i].HttpClientTimeout > 60 {
			programConfigure.DDNSTaskList[i].HttpClientTimeout = 60
		}
	}
}

func DDNSTaskSetWebhookCallResult(taskKey string, result bool, message string) {
	programConfigureMutex.Lock()
	defer programConfigureMutex.Unlock()
	taskIndex := -1

	for i := range programConfigure.DDNSTaskList {
		if programConfigure.DDNSTaskList[i].TaskKey == taskKey {
			taskIndex = i
			break
		}
	}
	if taskIndex == -1 {
		return
	}

	log.Printf("DDNSTaskSetWebhookCallResult %s", taskKey)

}

func GetDDNSTaskConfigureList() []*DDNSTask {
	programConfigureMutex.RLock()
	defer programConfigureMutex.RUnlock()

	var resList []*DDNSTask

	for i := range programConfigure.DDNSTaskList {
		task := programConfigure.DDNSTaskList[i]
		resList = append(resList, &task)
	}
	return resList
}

func GetDDNSTaskByKey(taskKey string) *DDNSTask {
	programConfigureMutex.RLock()
	defer programConfigureMutex.RUnlock()
	taskIndex := -1

	for i := range programConfigure.DDNSTaskList {
		if programConfigure.DDNSTaskList[i].TaskKey == taskKey {
			taskIndex = i
			break
		}
	}
	if taskIndex == -1 {
		return nil
	}
	res := programConfigure.DDNSTaskList[taskIndex]

	return &res
}

func DDNSTaskListAdd(task *DDNSTask) error {
	programConfigureMutex.Lock()
	defer programConfigureMutex.Unlock()
	task.TaskKey = stringsp.GetRandomString(16)
	programConfigure.DDNSTaskList = append(programConfigure.DDNSTaskList, *task)
	return Save()
}

func DDNSTaskListDelete(taskKey string) error {
	programConfigureMutex.Lock()
	defer programConfigureMutex.Unlock()

	taskIndex := -1

	for i := range programConfigure.DDNSTaskList {
		if programConfigure.DDNSTaskList[i].TaskKey == taskKey {
			taskIndex = i
			break
		}
	}

	if taskIndex == -1 {
		return fmt.Errorf("找不到需要删除的DDNS任务")
	}

	programConfigure.DDNSTaskList = DeleteDDNSTaskListlice(programConfigure.DDNSTaskList, taskIndex)
	return Save()
}

func CheckDDNSTaskAvalid(task *DDNSTask) error {
	if len(task.URL) == 0 {
		if task.TaskType == "IPv6" {
			task.URL = checkIPv6URLList
		} else {
			task.URL = checkIPv4URLList
		}
	}

	switch task.DNS.Name {
	case "cloudflare":
		if task.DNS.Secret == "" {
			return fmt.Errorf("cloudflare token不能为空")
		}
	case "callback":
		if task.DNS.Callback.URL == "" {
			return fmt.Errorf("callback URL不能为空")
		}

		if task.DNS.Callback.Method == "" {
			return fmt.Errorf("请选择callback method")
		}
	default:
		if task.DNS.ID == "" || task.DNS.Secret == "" {
			return fmt.Errorf("dns服务商相关参数不能为空")
		}
	}

	if len(task.Domains) <= 0 {
		return fmt.Errorf("域名列表不能为空")
	}

	return nil
}

func EnableDDNSTaskByKey(taskKey string, enable bool) error {
	programConfigureMutex.Lock()
	defer programConfigureMutex.Unlock()
	taskIndex := -1

	for i := range programConfigure.DDNSTaskList {
		if programConfigure.DDNSTaskList[i].TaskKey == taskKey {
			taskIndex = i
			break
		}
	}
	if taskIndex == -1 {
		return fmt.Errorf("开关DDNS任务失败,TaskKey不存在")
	}
	programConfigure.DDNSTaskList[taskIndex].Enable = enable

	return Save()
}

func UpdateTaskToDDNSTaskList(taskKey string, task DDNSTask) error {
	programConfigureMutex.Lock()
	defer programConfigureMutex.Unlock()
	taskIndex := -1

	for i := range programConfigure.DDNSTaskList {
		if programConfigure.DDNSTaskList[i].TaskKey == taskKey {
			taskIndex = i
			break
		}
	}

	if taskIndex == -1 {
		return fmt.Errorf("找不到需要更新的DDNS任务")
	}

	programConfigure.DDNSTaskList[taskIndex].TaskName = task.TaskName
	programConfigure.DDNSTaskList[taskIndex].TaskType = task.TaskType
	programConfigure.DDNSTaskList[taskIndex].Enable = task.Enable
	programConfigure.DDNSTaskList[taskIndex].GetType = task.GetType
	programConfigure.DDNSTaskList[taskIndex].URL = task.URL
	programConfigure.DDNSTaskList[taskIndex].NetInterface = task.NetInterface
	programConfigure.DDNSTaskList[taskIndex].IPReg = task.IPReg
	programConfigure.DDNSTaskList[taskIndex].Domains = task.Domains
	programConfigure.DDNSTaskList[taskIndex].DNS = task.DNS
	programConfigure.DDNSTaskList[taskIndex].Webhook = task.Webhook
	programConfigure.DDNSTaskList[taskIndex].TTL = task.TTL
	programConfigure.DDNSTaskList[taskIndex].HttpClientTimeout = task.HttpClientTimeout

	return Save()
}

func DeleteDDNSTaskListlice(a []DDNSTask, deleteIndex int) []DDNSTask {
	j := 0
	for i := range a {
		if i != deleteIndex {
			a[j] = a[i]
			j++
		}
	}
	return a[:j]
}

//****************************
