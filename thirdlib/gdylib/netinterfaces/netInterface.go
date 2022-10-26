package netinterfaces

import (
	"log"
	"net"
	"regexp"
	"strconv"
	"strings"
)

// NetInterface 本机网络
type NetInterface struct {
	NetInterfaceName string
	AddressList      []string
}

// GetNetInterface 获得网卡地址
// 返回ipv4, ipv6地址
func GetNetInterface() (ipv4NetInterfaces []NetInterface, ipv6NetInterfaces []NetInterface, err error) {
	allNetInterfaces, err := net.Interfaces()
	if err != nil {
		//fmt.Println("net.Interfaces failed, err:", err.Error())
		return ipv4NetInterfaces, ipv6NetInterfaces, err
	}

	// https://en.wikipedia.org/wiki/IPv6_address#General_allocation
	//_, ipv6Unicast, _ := net.ParseCIDR("2000::/3")

	for i := 0; i < len(allNetInterfaces); i++ {
		if (allNetInterfaces[i].Flags & net.FlagUp) != 0 {
			addrs, _ := allNetInterfaces[i].Addrs()
			ipv4 := []string{}
			ipv6 := []string{}

			for _, address := range addrs {
				if ipnet, ok := address.(*net.IPNet); ok && ipnet.IP.IsGlobalUnicast() {
					_, bits := ipnet.Mask.Size()
					// 需匹配全局单播地址
					//if bits == 128 && ipv6Unicast.Contains(ipnet.IP) {
					if bits == 128 {
						ipv6 = append(ipv6, ipnet.IP.String())
					}
					if bits == 32 {
						ipv4 = append(ipv4, ipnet.IP.String())
					}
				}
			}

			if len(ipv4) > 0 {
				ipv4NetInterfaces = append(
					ipv4NetInterfaces,
					NetInterface{
						NetInterfaceName: allNetInterfaces[i].Name,
						AddressList:      ipv4,
					},
				)
			}

			if len(ipv6) > 0 {
				ipv6NetInterfaces = append(
					ipv6NetInterfaces,
					NetInterface{
						NetInterfaceName: allNetInterfaces[i].Name,
						AddressList:      ipv6,
					},
				)
			}

		}
	}

	return ipv4NetInterfaces, ipv6NetInterfaces, nil
}

func GetIPFromNetInterface(ipType, netinterface, ipreg string) string {
	ipv4NetInterfaces, ipv6NetInterfaces, err := GetNetInterface()
	if err != nil {
		log.Printf("获取网卡信息出错：%s", err.Error())
		return ""
	}

	var netInterfaces []NetInterface

	switch ipType {
	case "IPv6":
		netInterfaces = ipv6NetInterfaces
	case "IPv4":
		netInterfaces = ipv4NetInterfaces
	default:
		log.Printf("未知IP类型")
		return ""
	}

	var addressList []string
	for i := range netInterfaces {
		if netInterfaces[i].NetInterfaceName == netinterface {
			addressList = netInterfaces[i].AddressList
			break
		}
	}

	if len(addressList) <= 0 {
		return ""
	}

	if ipreg == "" { //默认返回第一个IP
		return addressList[0]
	}

	ipN, err := strconv.Atoi(ipreg)
	if err == nil { //选择第N个IP
		if len(addressList) < ipN {
			log.Printf("当前选择网卡[%s]的第[%d]个IP,超出列表范围", netinterface, ipN)
			return ""
		}
		return addressList[ipN-1]
	}

	for i := range addressList {
		matched, err := regexp.MatchString(ipreg, addressList[i])
		if matched && err == nil {
			//log.Printf("正则匹配上")
			return addressList[i]
		}
	}

	if len(ipreg) <= 1 {
		return ""
	}

	if ipreg[len(ipreg)-1] == '*' {
		prefixStr := ipreg[:len(ipreg)-1]
		log.Printf("匹配以 %s 开头的IP", prefixStr)
		for i := range addressList {
			if strings.HasPrefix(addressList[i], prefixStr) {
				return addressList[i]
			}
		}
		return ""
	}

	if ipreg[0] == '*' {
		suffixStr := ipreg[1:]
		log.Printf("匹配以 %s 结尾的IP", suffixStr)
		for i := range addressList {
			if strings.HasSuffix(addressList[i], suffixStr) {
				return addressList[i]
			}
		}
		return ""
	}

	return ""
}

func GetGlobalIPv4BroadcastList() []string {
	var res []string
	allNetInterfaces, err := net.Interfaces()
	if err != nil {
		return res
	}

	for i := 0; i < len(allNetInterfaces); i++ {
		if (allNetInterfaces[i].Flags & net.FlagUp) != 0 {
			addrs, _ := allNetInterfaces[i].Addrs()

			for _, address := range addrs {
				if ipnet, ok := address.(*net.IPNet); ok && ipnet.IP.IsGlobalUnicast() {
					_, bits := ipnet.Mask.Size()
					// 需匹配全局单播地址
					//if bits == 128 && ipv6Unicast.Contains(ipnet.IP) {

					if bits == 32 {
						//ipv4 = append(ipv4, ipnet.IP.String())
						bcst := GetBroadcast(ipnet.IP, ipnet.Mask)
						res = append(res, bcst)
					}
				}
			}

		}
	}

	return res
}

func GetBroadcast(ip net.IP, mask net.IPMask) string {

	bcst := make(net.IP, len(ip))
	copy(bcst, ip)
	for i := 0; i < len(mask); i++ {
		ipIdx := len(bcst) - i - 1
		bcst[ipIdx] = ip[ipIdx] | ^mask[len(mask)-i-1]
	}
	return bcst.String()
}
