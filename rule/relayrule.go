// Copyright 2022 gdy, 272288813@qq.com
package rule

import (
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"

	"github.com/gdy666/lucky/socketproxy"
)

type RelayRule struct {
	Name                     string                       `json:"Name"`
	MainConfigure            string                       `json:"Mainconfigure"`
	RelayType                string                       `json:"RelayType"`
	ListenIP                 string                       `json:"ListenIP"`
	ListenPorts              string                       `json:"ListenPorts"`
	TargetIP                 string                       `json:"TargetIP"`
	TargetPorts              string                       `json:"TargetPorts"`
	BalanceTargetAddressList []string                     `json:"BalanceTargetAddressList"`
	Options                  socketproxy.RelayRuleOptions `json:"Options"`
	SubRuleList              []SubRelayRule               `json:"SubRuleList"`
	From                     string                       `json:"From"`
	IsEnable                 bool                         `json:"Enable"`
	proxyList                *[]socketproxy.Proxy         `json:"-"`
}

type SubRelayRule struct {
	ProxyType                   string   `json:"ProxyType"`
	BindIP                      string   `json:"BindIP"`
	ListenPorts                 []int    `json:"ListenPorts"`
	TargetHost                  string   `json:"TargetHost"`
	TargetPorts                 []int    `json:"TargetPorts"`
	BalanceTargetAddressAddress []string `json:"BalanceTargetAddressAddress"`
}

func (r *RelayRule) Enable() {
	r.IsEnable = true
	if r.proxyList == nil {
		return
	}
	for _, p := range *r.proxyList {
		p.StartProxy()
	}
}

func (r *RelayRule) GetProxyCount() int64 {
	if r.proxyList == nil {
		return 0
	}
	return int64(len(*r.proxyList))
}

func (r *RelayRule) Disable() {
	r.IsEnable = false
	if r.proxyList == nil {
		return
	}
	for _, p := range *r.proxyList {
		p.StopProxy()
	}
}

func GetRelayRulesFromCMD(configureList []string, options *socketproxy.RelayRuleOptions) (relayRules *[]RelayRule, err error) {
	//proxyMap := make(map[string]socketproxy.Proxy)

	var relayRuleList []RelayRule

	for _, configure := range configureList {

		relayRule, err := CreateRuleByConfigureAndOptions("", configure, *options)
		if err != nil {
			return nil, err
		}
		relayRule.From = "cmd" //规则来源

		relayRuleList = append(relayRuleList, *relayRule)
	}

	return &relayRuleList, nil
}

func (r *RelayRule) CreateMainConfigure() (configure string) {
	if len(r.BalanceTargetAddressList) > 0 {
		configure = fmt.Sprintf("%s@%s:%sto%s", r.RelayType, r.ListenIP, r.ListenPorts, strings.Join(r.BalanceTargetAddressList, "|"))
	} else {
		if strings.Compare(r.ListenPorts, r.TargetPorts) == 0 {
			configure = fmt.Sprintf("%s@%s:%sto%s", r.RelayType, r.ListenIP, r.ListenPorts, r.TargetIP)
		} else {
			configure = fmt.Sprintf("%s@%s:%sto%s:%s", r.RelayType, r.ListenIP, r.ListenPorts, r.TargetIP, r.TargetPorts)
		}
	}
	return configure
}

func CreateRuleByConfigureAndOptions(name, configureStr string, options socketproxy.RelayRuleOptions) (rule *RelayRule, err error) {
	var r RelayRule
	r.Options = options
	r.SubRuleList, r.RelayType, r.ListenIP, r.ListenPorts, r.TargetIP, r.TargetPorts, r.BalanceTargetAddressList, err = createSubRuleListFromConfigure(configureStr)
	if err != nil {
		return nil, err
	}

	r.MainConfigure = r.CreateMainConfigure()

	// if len(r.BalanceTargetAddressList) > 0 {
	// 	r.MainConfigure = fmt.Sprintf("%s@%s:%sto%s", r.RelayType, r.ListenIP, r.ListenPorts, strings.Join(r.BalanceTargetAddressList, ","))
	// } else {
	// 	if strings.Compare(r.ListenPorts, r.TargetPorts) == 0 {
	// 		r.MainConfigure = fmt.Sprintf("%s@%s:%sto%s", r.RelayType, r.ListenIP, r.ListenPorts, r.TargetIP)
	// 	} else {
	// 		r.MainConfigure = fmt.Sprintf("%s@%s:%sto%s:%s", r.RelayType, r.ListenIP, r.ListenPorts, r.TargetIP, r.TargetPorts)
	// 	}
	// }

	var pl []socketproxy.Proxy

	for i := range r.SubRuleList {
		if len(r.BalanceTargetAddressList) == 0 {
			for j := range r.SubRuleList[i].ListenPorts {
				p, e := socketproxy.CreateProxy(r.SubRuleList[i].ProxyType,
					r.SubRuleList[i].BindIP,
					r.SubRuleList[i].TargetHost,
					nil,
					r.SubRuleList[i].ListenPorts[j],
					r.SubRuleList[i].TargetPorts[j],
					&options)
				if e != nil {
					log.Printf("CreateProxy error:%s", e.Error())
					continue
				}
				p.SetFromRule(r.MainConfigure)
				pl = append(pl, p)
			}
			continue
		}

		p, e := socketproxy.CreateProxy(r.SubRuleList[i].ProxyType,
			r.SubRuleList[i].BindIP,
			r.SubRuleList[i].TargetHost,
			&r.BalanceTargetAddressList,
			r.SubRuleList[i].ListenPorts[0],
			0,
			&options)
		if e != nil {
			log.Printf("CreateProxy error:%s", e.Error())
			continue
		}
		p.SetFromRule(r.MainConfigure)
		pl = append(pl, p)

	}

	r.proxyList = &pl

	r.Name = name

	return &r, nil
}

func createSubRuleListFromConfigure(str string) (subRelyList []SubRelayRule, proxytypeListStr, listenIP, listenPortsStr, targetIP, targetePortsStr string, targetAddressList []string, err error) {
	//log.Printf("ConfigureStr:%s\n", str)
	splitRes := strings.Split(str, "@")
	if len(splitRes) > 2 {
		err = fmt.Errorf("relay参数:%s格式有误!000", str)
		return
	}

	proxytypeListStr = "tcp,udp"

	relayConfig := splitRes[0]

	if len(splitRes) == 2 {
		proxytypeListStr = splitRes[0]
		relayConfig = splitRes[1]
	}

	proxyTypeList := getProxyTypeList(proxytypeListStr)
	err = checkProxyType(proxyTypeList)
	if err != nil {
		return
	}
	proxytypeListStr = convertProxyTypeByList(proxyTypeList)

	relayConfigArray := strings.Split(relayConfig, "to")
	if len(relayConfigArray) > 2 {
		err = fmt.Errorf("relay参数:%s格式有误!001", str)
		return
	}

	var listenPorts []int
	var targetPorts []int

	switch len(relayConfigArray) {
	case 1: //监听端口没有指定 比如 192.168.31.22:80,443,20000-20010
		tip, tPortsStr, e := getIpAndPortFromAddress(relayConfigArray[0], true, true)
		if e != nil {
			err = fmt.Errorf("参数中目标地址部分参数[%s]格式有误", relayConfigArray[0])
			return
		}

		targetPorts, e = portsStrToIList(tPortsStr)
		if e != nil {
			err = fmt.Errorf("参数[%s]中的目标端口部分出错:%s", str, e.Error())
			return
		}
		targetePortsStr = tPortsStr
		listenPortsStr = tPortsStr
		listenPorts = targetPorts
		targetIP = tip

	case 2: //监听端口有指定 比如 80,443,20000-20010to192.168.31.222 ,但目标地址的端口不一定指定
		bindAddress := relayConfigArray[0]
		bip, bPortsStr, e := getIpAndPortFromAddress(bindAddress, false, false)
		if e != nil {
			err = fmt.Errorf("参数[%s]中的监听端口部分出错:%s", str, e.Error())
			return
		}
		listenIP = bip

		listenPortsStr = bPortsStr
		//fmt.Printf("bip:%s bindPortsStr:%s\n", bip, bindPortsStr)

		listenPorts, e = portsStrToIList(bPortsStr)
		if e != nil {
			err = fmt.Errorf("参数中绑定端口部分参数[%s]格式有误", bPortsStr)
			return
		}

		//log.Printf("relayConfigArray[1]:\t[%s]", relayConfigArray[1])

		if strings.Contains(relayConfigArray[1], "|") { //均衡负载模式
			if len(listenPorts) != 1 {
				err = fmt.Errorf("均衡负载模式一条配置指定监听一个端口")
				return
			}
			targetAddressList = strings.Split(relayConfigArray[1], "|")

		} else {
			targetAddress := relayConfigArray[1]

			tip, tPortsStr, e := getIpAndPortFromAddress(targetAddress, true, false)
			if e != nil {
				err = fmt.Errorf("参数中目标地址部分参数[%s]格式有误", targetAddress)
				return
			}
			targetePortsStr = tPortsStr

			targetPorts, e = portsStrToIList(tPortsStr)

			if e != nil {
				err = fmt.Errorf("参数[%s]中的目标端口部分出错:%s", str, e.Error())
				return
			}

			if len(listenPorts) > 0 && len(targetPorts) == 0 {
				targetPorts = listenPorts
				targetePortsStr = listenPortsStr
			} else if len(listenPorts) == 0 && len(targetPorts) > 0 {
				listenPorts = targetPorts
				listenPortsStr = targetePortsStr
			}

			if len(listenPorts) != len(targetPorts) {
				err = fmt.Errorf("参数[%s]中监听端口数量和目标端口数量不一致", str)

				// fmt.Printf("listenPorts:%v\n", listenPorts)
				// fmt.Printf("targetPorts:%v\n", targetPorts)
				return
			}
			targetIP = tip
		}

	default:

	}

	var SubBaseRule SubRelayRule
	SubBaseRule.BindIP = listenIP
	SubBaseRule.ListenPorts = append(SubBaseRule.ListenPorts, listenPorts...)

	if len(targetAddressList) == 0 {
		SubBaseRule.TargetHost = targetIP
		SubBaseRule.TargetPorts = append(SubBaseRule.TargetPorts, targetPorts...)

	} else {
		SubBaseRule.BalanceTargetAddressAddress = targetAddressList
	}

	for i := range proxyTypeList {
		dt := SubBaseRule
		dt.ProxyType = proxyTypeList[i]
		subRelyList = append(subRelyList, dt)
	}

	return
}

func convertProxyTypeByList(proxyTypeList []string) (proxyType string) {
	for i := range proxyTypeList {
		if i == 0 {
			proxyType = proxyTypeList[i]
			continue
		}
		proxyType += "," + proxyTypeList[i]
	}
	return
}

func getProxyTypeList(proxyTypeListStr string) (proxyTypeList []string) {
	//tmpList = strings.Split(proxyTypeListStr, ",")
	//var
	tmpMap := make(map[string]int)

	if strings.Contains(proxyTypeListStr, "tcp4") && strings.Contains(proxyTypeListStr, "tcp6") {
		proxyTypeList = append(proxyTypeList, "tcp")
		tmpMap["tcp"] = 1
		proxyTypeListStr = strings.Replace(proxyTypeListStr, "tcp4", ",", -1)
		proxyTypeListStr = strings.Replace(proxyTypeListStr, "tcp6", ",", -1)
	}

	if strings.Contains(proxyTypeListStr, "udp4") && strings.Contains(proxyTypeListStr, "udp6") {
		proxyTypeList = append(proxyTypeList, "udp")
		tmpMap["udp"] = 1
		proxyTypeListStr = strings.Replace(proxyTypeListStr, "udp4", ",", -1)
		proxyTypeListStr = strings.Replace(proxyTypeListStr, "udp6", ",", -1)
	}

	tmpList := strings.Split(proxyTypeListStr, ",")
	for i := range tmpList {
		if len(tmpList[i]) <= 2 {
			continue
		}
		if _, ok := tmpMap[tmpList[i]]; ok {
			continue
		}

		_, tcpOK := tmpMap["tcp"]
		_, udpOK := tmpMap["udp"]

		if (tmpList[i] == "tcp4" || tmpList[i] == "tcp6") && tcpOK {
			continue
		}

		if (tmpList[i] == "udp4" || tmpList[i] == "udp6") && udpOK {
			continue
		}

		proxyTypeList = append(proxyTypeList, tmpList[i])
		tmpMap[tmpList[i]] = 1
	}

	return
}

func checkProxyType(proxyTypeList []string) error {
	for _, proxyType := range proxyTypeList {

		switch proxyType {
		case "tcp", "tcp4", "tcp6", "udp", "udp4", "udp6":
			{
				return nil
			}
		default:
			{
				return fmt.Errorf("unsupport Proxy Type:%s", proxyType)
			}
		}

	}
	return nil

}

// CheckProxyConflict 冲突检查
func CheckProxyConflict(proxyList *[]socketproxy.Proxy, proxyType, listenIP string, listenPort int) error {
	proxyMap := make(map[string]socketproxy.Proxy)
	for i, p := range *proxyList {
		proxyMap[p.GetKey()] = (*proxyList)[i]
	}

	key := socketproxy.GetProxyKey(proxyType, listenIP, listenPort)
	if _, ok := proxyMap[key]; ok {
		return fmt.Errorf("绑定的地址和端口存在冲突！[%s]", key)
	}
	anyBindKey := fmt.Sprintf("%s@:%d", proxyType, listenPort)

	if strings.Compare(key, anyBindKey) == 0 {
		for exitsKey := range proxyMap {
			if strings.HasSuffix(exitsKey, anyBindKey) {
				return fmt.Errorf("绑定的地址和端口存在冲突![%s][%s]", key, exitsKey)
			}
		}

	} else {
		if _, ok := proxyMap[anyBindKey]; ok {
			return fmt.Errorf("绑定的地址和端口存在冲突！[%s]", key)
		}
	}

	return nil
}

// getIpAndPortsFromAddress
func getIpAndPortFromAddress(address string, needip bool, needports bool) (ip string, ports string, err error) {
	ipAndPortIndex := strings.LastIndex(address, ":")

	// defer func() {
	// 	fmt.Printf("\nFuck: [%s]--->[%s]", ip, ports)
	// }()

	if ipAndPortIndex < 0 || (!needip && !needports) {

		switch {
		case (!needip && needports): //地址中仅有端口
			ports = address
		case (needip && !needports): //地址中仅有ip
			{
				ip = address
			}
		case (!needip && !needports): //地址中 端口和ip都不是必须
			{

				if ipAndPortIndex > 0 {
					ip = address[:ipAndPortIndex]
					ports = address[ipAndPortIndex+1:]
					break
				}

				//但address非空,判断
				if address == "" {
					break
				}
				if net.ParseIP(address) != nil {
					ip = address
				} else {
					ports = address
					if strings.HasPrefix(ports, ":") {
						ports = ports[1:]
					}
				}
			}
		default:

		}
		return
	}

	ports = address[ipAndPortIndex+1:]

	if ipAndPortIndex <= 1 {
		//fmt.Printf("Fuck:%s\n", ports)
		return
	}

	addressHost := address[:ipAndPortIndex]
	addressHost = strings.Replace(addressHost, "[", "", -1)
	addressHost = strings.Replace(addressHost, "]", "", -1)

	if net.ParseIP(addressHost) == nil {
		err = fmt.Errorf("ip[%s]格式有误", address[:ipAndPortIndex])
		return
	}

	ip = addressHost

	return
}

// portsStrToIList
func portsStrToIList(portsStr string) (ports []int, err error) {

	if portsStr == "" {
		return
	}

	if strings.Contains(portsStr, ",") {
		tmpStrList := strings.Split(portsStr, ",")
		for i := range tmpStrList {
			tps, e := portsStrToIList(tmpStrList[i])
			if e != nil {
				err = fmt.Errorf("端口参数处理出错:%s", e.Error())
				return
			}
			ports = append(ports, tps...)
		}

		return
	}

	portsStrList := strings.Split(portsStr, "-")
	if len(portsStrList) > 2 {
		err = fmt.Errorf("端口%s格式有误", portsStr)
		return
	}

	if len(portsStrList) == 1 { //single listen port
		listenPort, e := portStrToi(portsStrList[0])
		if e != nil {
			err = fmt.Errorf("端口格式有误!%s", e.Error())
			return
		}
		ports = append(ports, listenPort)
	}

	if len(portsStrList) == 2 {
		minListenPort, e := portStrToi(portsStrList[0])
		if e != nil {
			err = fmt.Errorf("端口格式有误!%s", portsStrList[0])
			return
		}
		maxListenPort, e := portStrToi(portsStrList[1])
		if e != nil {
			err = fmt.Errorf("端口格式有误!%s", portsStrList[1])
			return
		}

		if maxListenPort <= minListenPort {
			err = fmt.Errorf("前一个端口[%d]要小于后一个端口[%d]", minListenPort, maxListenPort)
			return
		}
		i := minListenPort
		for {
			if i > maxListenPort {
				break
			}
			ports = append(ports, i)
			i++
		}
	}

	return
}

func portStrToi(portStr string) (int, error) {
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return 0, fmt.Errorf("端口格式有误:%s", err.Error())
	}
	if port < 1 || port > 65535 {
		return 0, fmt.Errorf("端口[%d]超出范围", port)
	}
	return port, nil
}
