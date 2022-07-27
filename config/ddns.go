package config

import (
	"fmt"
	"io/ioutil"
	"log"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/gdy666/lucky/thirdlib/gdylib/httputils"
	"github.com/gdy666/lucky/thirdlib/gdylib/stringsp"
)

// Ipv4Reg IPv4正则
const Ipv4Reg = `((25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])\.){3,3}(25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])`

// Ipv6Reg IPv6正则
const Ipv6Reg = `((([0-9A-Fa-f]{1,4}:){7}([0-9A-Fa-f]{1,4}|:))|(([0-9A-Fa-f]{1,4}:){6}(:[0-9A-Fa-f]{1,4}|((25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)(\.(25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)){3})|:))|(([0-9A-Fa-f]{1,4}:){5}(((:[0-9A-Fa-f]{1,4}){1,2})|:((25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)(\.(25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)){3})|:))|(([0-9A-Fa-f]{1,4}:){4}(((:[0-9A-Fa-f]{1,4}){1,3})|((:[0-9A-Fa-f]{1,4})?:((25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)(\.(25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)){3}))|:))|(([0-9A-Fa-f]{1,4}:){3}(((:[0-9A-Fa-f]{1,4}){1,4})|((:[0-9A-Fa-f]{1,4}){0,2}:((25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)(\.(25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)){3}))|:))|(([0-9A-Fa-f]{1,4}:){2}(((:[0-9A-Fa-f]{1,4}){1,5})|((:[0-9A-Fa-f]{1,4}){0,3}:((25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)(\.(25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)){3}))|:))|(([0-9A-Fa-f]{1,4}:){1}(((:[0-9A-Fa-f]{1,4}){1,6})|((:[0-9A-Fa-f]{1,4}){0,4}:((25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)(\.(25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)){3}))|:))|(:(((:[0-9A-Fa-f]{1,4}){1,7})|((:[0-9A-Fa-f]{1,4}){0,5}:((25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)(\.(25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)){3}))|:)))`

var ipUrlAddrMap sync.Map

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
	//-------------------------------------
	//IpCache     IpCache `json:"-"`
	DomainsState DomainsState `json:"-"`
}

type Webhook struct {
	WebhookEnable          bool     `json:"WebhookEnable"`          //Webhook开关
	WebhookCallOnGetIPfail bool     `json:"WebhookCallOnGetIPfail"` //获取IP失败时触发Webhook 开关
	WebhookURL             string   `json:"WebhookURL"`
	WebhookMethod          string   `json:"WebhookMethod"`
	WebhookHeaders         []string `json:"WebhookHeaders"`
	WebhookRequestBody     string   `json:"WebhookRequestBody"`
	WebhookSuccessContent  []string `json:"WebhookSuccessContent"` //接口调用成功包含的内容
	WebhookProxy           string   `json:"WebhookProxy"`          //使用DNS代理设置  ""表示禁用，"dns"表示使用dns的代理设置
	WebhookProxyAddr       string   `json:"WebhookProxyAddr"`      //代理服务器IP
	WebhookProxyUser       string   `json:"WebhookProxyUser"`      //代理用户
	WebhookProxyPassword   string   `json:"WebhookProxyPassword"`  //代理密码
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
	Callback                DNSCallback `json:"Callback"`
	HttpClientProxyType     string      `json:"HttpClientProxyType"`     //http client代理服务器设置
	HttpClientProxyAddr     string      `json:"HttpClientProxyAddr"`     //代理服务器IP
	HttpClientProxyUser     string      `json:"HttpClientProxyUser"`     //代理用户
	HttpClientProxyPassword string      `json:"HttpClientProxyPassword"` //代理密码
}

type DNSCallback struct {
	URL                    string   `json:"URL"`    //请求地址
	Method                 string   `json:"Method"` //请求方法
	Headers                []string `json:"Headers"`
	RequestBody            string   `json:"RequestBody"`
	Server                 string   `json:"Server"`                 //预设服务商
	CallbackSuccessContent []string `json:"CallbackSuccessContent"` //接口调用成功包含内容
}

//Check 检测IP是否有改变
func (d *DDNSTask) IPChangeCheck(newAddr string) bool {
	if newAddr == "" {
		return true
	}
	// 地址改变
	if d.DomainsState.IpAddr != newAddr {
		//log.Printf("公网地址改变:[%s]===>[%s]", d.DomainsInfo.IpAddr, newAddr)
		d.DomainsState.IpAddr = newAddr
		return true
	}
	return false
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

func CleanIPUrlAddrMap() {
	keys := []string{}
	ipUrlAddrMap.Range(func(key, value any) bool {
		keys = append(keys, key.(string))
		return true
	})
	for _, k := range keys {
		ipUrlAddrMap.Delete(k)
	}
}

func DDNSTaskListTaskDetailsInit() {
	programConfigureMutex.Lock()
	defer programConfigureMutex.Unlock()
	for i := range programConfigure.DDNSTaskList {
		programConfigure.DDNSTaskList[i].DomainsState.Init(programConfigure.DDNSTaskList[i].Domains)
		programConfigure.DDNSTaskList[i].DomainsState.SetDomainUpdateStatus(UpdateWaiting, "")

		//
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

func DDNSTaskIPCacheCheck(taskKey, ip string) (bool, error) {
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
		return true, fmt.Errorf("找不到key对应的DDNS任务")
	}
	return programConfigure.DDNSTaskList[taskIndex].IPChangeCheck(ip), nil
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

type DDNSTaskDetails struct {
	DDNSTask
	TaskState DomainsState `json:"TaskState"`
}

func GetDDNSTaskList() []DDNSTaskDetails {
	programConfigureMutex.RLock()
	defer programConfigureMutex.RUnlock()

	var resList []DDNSTaskDetails

	for i := range programConfigure.DDNSTaskList {
		var info DDNSTaskDetails
		programConfigure.DDNSTaskList[i].DomainsState.Mutex.RLock()
		info.DDNSTask = programConfigure.DDNSTaskList[i]
		info.TaskState = programConfigure.DDNSTaskList[i].DomainsState
		programConfigure.DDNSTaskList[i].DomainsState.Mutex.RUnlock()
		resList = append(resList, info)
	}
	return resList
}

func GetDDNSTaskByKey(taskKey string) *DDNSTaskDetails {
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
		return nil
	}
	var info DDNSTaskDetails
	programConfigure.DDNSTaskList[taskIndex].DomainsState.Mutex.RLock()
	info.DDNSTask = programConfigure.DDNSTaskList[taskIndex]
	info.TaskState = programConfigure.DDNSTaskList[taskIndex].DomainsState
	programConfigure.DDNSTaskList[taskIndex].DomainsState.Mutex.RUnlock()
	return &info
}

func DDNSTaskListFlushDomainsDetails(taskKey string, state *DomainsState) {
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

	var checkDomains []*Domain
	//防止有域名被删除
	for _, new := range state.Domains {
		for j, pre := range programConfigure.DDNSTaskList[taskIndex].DomainsState.Domains {
			if strings.Compare(new.String(), pre.String()) == 0 {
				checkDomains = append(checkDomains, programConfigure.DDNSTaskList[taskIndex].DomainsState.Domains[j])
				break
			}
		}
	}

	state.Domains = checkDomains

	programConfigure.DDNSTaskList[taskIndex].DomainsState = *state
}

func DDNSTaskListAdd(task *DDNSTask) error {
	programConfigureMutex.Lock()
	defer programConfigureMutex.Unlock()
	task.TaskKey = stringsp.GetRandomString(16)
	task.DomainsState.Init(task.Domains)
	task.DomainsState.SetDomainUpdateStatus(UpdateWaiting, "")
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
	if enable {
		programConfigure.DDNSTaskList[taskIndex].DomainsState.SetDomainUpdateStatus(UpdateWaiting, "")
	} else {
		programConfigure.DDNSTaskList[taskIndex].DomainsState.SetDomainUpdateStatus(UpdateStop, "")
	}
	return Save()
}

func UpdateDomainsStateByTaskKey(taskKey string, status updateStatusType, message string) {
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
	programConfigure.DDNSTaskList[taskIndex].DomainsState.SetDomainUpdateStatus(status, message)
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
	programConfigure.DDNSTaskList[taskIndex].DomainsState.IpAddr = ""
	programConfigure.DDNSTaskList[taskIndex].DomainsState.Init(task.Domains)
	programConfigure.DDNSTaskList[taskIndex].DomainsState.IPAddrHistory = task.DomainsState.IPAddrHistory
	programConfigure.DDNSTaskList[taskIndex].DomainsState.WebhookCallHistroy = task.DomainsState.WebhookCallHistroy
	programConfigure.DDNSTaskList[taskIndex].DomainsState.SetDomainUpdateStatus(UpdateWaiting, "")
	programConfigure.DDNSTaskList[taskIndex].HttpClientTimeout = task.HttpClientTimeout
	programConfigure.DDNSTaskList[taskIndex].DomainsState.WebhookCallErrorMsg = ""
	programConfigure.DDNSTaskList[taskIndex].DomainsState.WebhookCallResult = false
	programConfigure.DDNSTaskList[taskIndex].DomainsState.WebhookCallTime = ""

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

func (d *DDNSTask) GetIpAddr() (result string) {
	if d.TaskType == "IPv6" {
		return d.getIpv6Addr()
	}
	return d.getIpv4Addr()
}

// getIpv4Addr 获得IPv4地址
func (d *DDNSTask) getIpv4Addr() (result string) {
	// 判断从哪里获取IP
	if d.GetType == "netInterface" {
		result = GetIPFromNetInterface("IPv4", d.NetInterface, d.IPReg)
		// 从网卡获取IP
		// ipv4, _, err := GetNetInterface()
		// if err != nil {
		// 	log.Println("从网卡获得IPv4失败!")
		// 	return
		// }

		// for _, netInterface := range ipv4 {
		// 	if netInterface.NetInterfaceName == d.NetInterface && len(netInterface.AddressList) > 0 {
		// 		return netInterface.AddressList[0]
		// 	}
		// }

		// log.Println("从网卡中获得IPv4失败! 网卡名: ", d.NetInterface)
		return
	}

	ddnsGlobalConf := GetDDNSConfigure()

	client, err := httputils.CreateHttpClient(
		ddnsGlobalConf.HttpClientSecureVerify,
		"",
		"",
		"",
		"",
		time.Duration(d.HttpClientTimeout)*time.Second)

	if err != nil {
		log.Printf("%s", err.Error())
		return
	}

	for _, url := range d.URL {
		url = strings.TrimSpace(url)

		mapIp, ok := ipUrlAddrMap.Load(url)
		if ok {
			//log.Printf("URL[%s]已缓存IP[%s]", url, mapIp)
			result = mapIp.(string)
			return
		}

		resp, err := client.Get(url)
		if err != nil {
			//log.Printf("连接失败!%s查看接口能否返回IPv4地址</a>,", url)
			continue
		}
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Printf("读取IPv4结果失败! 接口:%s", url)
			continue
		}
		comp := regexp.MustCompile(Ipv4Reg)
		result = comp.FindString(string(body))
		if result != "" {
			ipUrlAddrMap.Store(url, result)
			return
		}
		//  else {
		// 	log.Printf("获取IPv4结果失败! 接口: %s ,返回值: %s\n", url, result)
		// }
	}

	log.Printf("所有查询公网IPv4的接口均获取IPv4结果失败,请检查接口或当前网络情况")
	return
}

// getIpv6Addr 获得IPv6地址
func (d *DDNSTask) getIpv6Addr() (result string) {
	// 判断从哪里获取IP
	if d.GetType == "netInterface" {
		// 从网卡获取IP
		// _, ipv6, err := GetNetInterface()
		// if err != nil {
		// 	log.Println("从网卡获得IPv6失败!")
		// 	return
		// }

		// for _, netInterface := range ipv6 {
		// 	if netInterface.NetInterfaceName == d.NetInterface && len(netInterface.AddressList) > 0 {
		// 		if d.IPReg != "" {
		// 			log.Printf("IPv6将使用正则表达式 %s 进行匹配\n", d.IPReg)
		// 			for i := 0; i < len(netInterface.AddressList); i++ {
		// 				matched, err := regexp.MatchString(d.IPReg, netInterface.AddressList[i])
		// 				if matched && err == nil {
		// 					log.Println("匹配成功! 匹配到地址: ", netInterface.AddressList[i])
		// 					return netInterface.AddressList[i]
		// 				}
		// 				log.Printf("第 %d 个地址 %s 不匹配, 将匹配下一个地址\n", i+1, netInterface.AddressList[i])
		// 			}
		// 			log.Println("没有匹配到任何一个IPv6地址, 将使用第一个地址")
		// 		}
		// 		return netInterface.AddressList[0]
		// 	}
		// }

		// log.Println("从网卡中获得IPv6失败! 网卡名: ", d.NetInterface)
		result = GetIPFromNetInterface("IPv6", d.NetInterface, d.IPReg)
		return
	}

	ddnsGlobalConf := GetDDNSConfigure()
	client, err := httputils.CreateHttpClient(
		!ddnsGlobalConf.HttpClientSecureVerify,
		"",
		"",
		"",
		"",
		time.Duration(d.HttpClientTimeout)*time.Second)

	if err != nil {
		log.Printf("%s", err.Error())
		return
	}

	for _, url := range d.URL {
		url = strings.TrimSpace(url)

		mapIp, ok := ipUrlAddrMap.Load(url)
		if ok {
			//log.Printf("URL[%s]已缓存IP[%s]", url, mapIp)
			result = mapIp.(string)
			return
		}

		resp, err := client.Get(url)
		if err != nil {
			//log.Printf("连接失败! %s查看接口能否返回IPv6地址 ", url)
			continue
		}

		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Println("读取IPv6结果失败! 接口: ", url)
			continue
		}
		comp := regexp.MustCompile(Ipv6Reg)
		result = comp.FindString(string(body))
		if result != "" {
			ipUrlAddrMap.Store(url, result)
			return
		}
	}
	log.Printf("所有查询公网IPv6的接口均获取IPv6结果失败,请检查接口或当前网络情况")

	return
}
