package ddnscore

import (
	"io"
	"log"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/gdy666/lucky/config"
	"github.com/gdy666/lucky/thirdlib/gdylib/httputils"
	"github.com/gdy666/lucky/thirdlib/gdylib/netinterfaces"
)

// Ipv4Reg IPv4正则
const Ipv4Reg = `((25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])\.){3,3}(25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])`

// Ipv6Reg IPv6正则
const Ipv6Reg = `((([0-9A-Fa-f]{1,4}:){7}([0-9A-Fa-f]{1,4}|:))|(([0-9A-Fa-f]{1,4}:){6}(:[0-9A-Fa-f]{1,4}|((25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)(\.(25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)){3})|:))|(([0-9A-Fa-f]{1,4}:){5}(((:[0-9A-Fa-f]{1,4}){1,2})|:((25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)(\.(25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)){3})|:))|(([0-9A-Fa-f]{1,4}:){4}(((:[0-9A-Fa-f]{1,4}){1,3})|((:[0-9A-Fa-f]{1,4})?:((25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)(\.(25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)){3}))|:))|(([0-9A-Fa-f]{1,4}:){3}(((:[0-9A-Fa-f]{1,4}){1,4})|((:[0-9A-Fa-f]{1,4}){0,2}:((25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)(\.(25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)){3}))|:))|(([0-9A-Fa-f]{1,4}:){2}(((:[0-9A-Fa-f]{1,4}){1,5})|((:[0-9A-Fa-f]{1,4}){0,3}:((25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)(\.(25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)){3}))|:))|(([0-9A-Fa-f]{1,4}:){1}(((:[0-9A-Fa-f]{1,4}){1,6})|((:[0-9A-Fa-f]{1,4}){0,4}:((25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)(\.(25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)){3}))|:))|(:(((:[0-9A-Fa-f]{1,4}){1,7})|((:[0-9A-Fa-f]{1,4}){0,5}:((25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)(\.(25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)){3}))|:)))`

var ipUrlAddrMap sync.Map

type DDNSTaskInfo struct {
	config.DDNSTask
	TaskState DDNSTaskState `json:"TaskState"`
}

func (d *DDNSTaskInfo) getIpAddr() (result string) {
	if d.TaskType == "IPv6" {
		return d.getIpv6Addr()
	}
	return d.getIpv4Addr()
}

// CheckIPChange 检测公网IP是否改变
func (d *DDNSTaskInfo) CheckIPChange() (ipAddr string, change bool) {
	ipAddr = d.getIpAddr()

	checkIPChange := d.TaskState.IPChangeCheck(ipAddr)

	if checkIPChange {
		return ipAddr, true
	}

	//IP没变化
	return ipAddr, false
}

// getIpv4Addr 获得IPv4地址
func (d *DDNSTaskInfo) getIpv4Addr() (result string) {
	// 判断从哪里获取IP
	if d.GetType == "netInterface" {
		result = netinterfaces.GetIPFromNetInterface("IPv4", d.NetInterface, d.IPReg)
		return
	}

	ddnsGlobalConf := config.GetDDNSConfigure()

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
func (d *DDNSTaskInfo) getIpv6Addr() (result string) {
	// 判断从哪里获取IP
	if d.GetType == "netInterface" {
		result = netinterfaces.GetIPFromNetInterface("IPv6", d.NetInterface, d.IPReg)
		return
	}

	ddnsGlobalConf := config.GetDDNSConfigure()
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
