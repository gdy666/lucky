package ddnsconf

import (
	"fmt"
	"io"
	"log"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/gdy666/lucky/thirdlib/gdylib/httputils"
	"github.com/gdy666/lucky/thirdlib/gdylib/netinterfaces"
)

var getDDNSConfigureFunc func() DDNSConfigure

func SetGetDDNSConfigureFunc(f func() DDNSConfigure) {
	getDDNSConfigureFunc = f
}

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

	GetType      string    `json:"GetType"` //IP获取方式
	URL          []string  `json:"URL"`
	NetInterface string    `json:"NetInterface"`
	IPReg        string    `json:"IPReg"`
	Domains      []string  `json:"Domains"`
	DNS          DNSConfig `json:"DNS"`
	Webhook
	TTL               string `json:"TTL"`
	HttpClientTimeout int    `json:"HttpClientTimeout"`
	ModifyTime        int64  `json:"ModifyTime"`
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

// Ipv4Reg IPv4正则
const Ipv4Reg = `((25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])\.){3,3}(25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])`

// Ipv6Reg IPv6正则
const Ipv6Reg = `((([0-9A-Fa-f]{1,4}:){7}([0-9A-Fa-f]{1,4}|:))|(([0-9A-Fa-f]{1,4}:){6}(:[0-9A-Fa-f]{1,4}|((25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)(\.(25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)){3})|:))|(([0-9A-Fa-f]{1,4}:){5}(((:[0-9A-Fa-f]{1,4}){1,2})|:((25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)(\.(25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)){3})|:))|(([0-9A-Fa-f]{1,4}:){4}(((:[0-9A-Fa-f]{1,4}){1,3})|((:[0-9A-Fa-f]{1,4})?:((25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)(\.(25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)){3}))|:))|(([0-9A-Fa-f]{1,4}:){3}(((:[0-9A-Fa-f]{1,4}){1,4})|((:[0-9A-Fa-f]{1,4}){0,2}:((25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)(\.(25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)){3}))|:))|(([0-9A-Fa-f]{1,4}:){2}(((:[0-9A-Fa-f]{1,4}){1,5})|((:[0-9A-Fa-f]{1,4}){0,3}:((25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)(\.(25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)){3}))|:))|(([0-9A-Fa-f]{1,4}:){1}(((:[0-9A-Fa-f]{1,4}){1,6})|((:[0-9A-Fa-f]{1,4}){0,4}:((25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)(\.(25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)){3}))|:))|(:(((:[0-9A-Fa-f]{1,4}){1,7})|((:[0-9A-Fa-f]{1,4}){0,5}:((25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)(\.(25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)){3}))|:)))`

var ipUrlAddrMap sync.Map

func (d *DDNSTask) GetIpAddr() (result string) {
	if d.TaskType == "IPv6" {
		return d.getIpv6Addr()
	}
	return d.getIpv4Addr()
}

// getIpv4Addr 获取IPv4地址
func (d *DDNSTask) getIpv4Addr() (result string) {
	// 判断从哪里获取IP
	if d.GetType == "netInterface" {
		result = netinterfaces.GetIPFromNetInterface("IPv4", d.NetInterface, d.IPReg)
		return
	}

	ddnsGlobalConf := getDDNSConfigureFunc()
	client, err := httputils.CreateHttpClient(
		"tcp4",
		"",
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
		body, err := io.ReadAll(resp.Body)
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
		result = netinterfaces.GetIPFromNetInterface("IPv6", d.NetInterface, d.IPReg)
		return
	}

	ddnsGlobalConf := getDDNSConfigureFunc()
	client, err := httputils.CreateHttpClient(
		"tcp6",
		"",
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
		body, err := io.ReadAll(resp.Body)
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

var checkIPv4URLList = []string{"https://4.ipw.cn", "http://v4.ip.zxinc.org/getip", "https://myip4.ipip.net", "https://www.taobao.com/help/getip.php", "https://ddns.oray.com/checkip", "https://ip.3322.net", "https://v4.myip.la"}
var checkIPv6URLList = []string{"https://6.ipw.cn", "https://ipv6.ddnspod.com", "http://v6.ip.zxinc.org/getip", "https://speed.neu6.edu.cn/getIP.php", "https://v6.ident.me", "https://v6.myip.la"}

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
