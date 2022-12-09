package dnsutils

import (
	"fmt"
	"strings"

	"github.com/gdy666/lucky/thirdlib/gdylib/stringsp"
	"github.com/miekg/dns"
	"golang.org/x/net/idna"
)

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

func ResolveDomainAtServerList(queryType, domain string, dnsServerList []string) (string, error) {

	if len(dnsServerList) == 0 {
		if queryType == "AAAA" {
			dnsServerList = DefaultIPv6DNSServerList
		} else {
			dnsServerList = DefaultIPv4DNSServerList
		}
	}

	//some name that ought to exist, does not exist (NXDOMAIN)

	querytype, querytypeOk := dns.StringToType[strings.ToUpper(queryType)]
	if !querytypeOk {
		return "", fmt.Errorf("queryType error:%s", queryType)
	}

	if strings.HasPrefix(domain, "*.") {
		randomStr := stringsp.GetRandomString(8)
		domain = strings.Replace(domain, "*", randomStr, 1)
	}

	domain = dns.Fqdn(domain)
	domain, err := idna.ToASCII(domain)
	if err != nil {
		return "", fmt.Errorf(` idna.ToASCII(domain) error:%s`, err.Error())
	}

	m := new(dns.Msg)
	m.SetQuestion(domain, querytype)
	m.MsgHdr.RecursionDesired = true

	c := new(dns.Client)
	noExistTimes := 0
	for _, dnsServer := range dnsServerList {
		c.Net = ""
		ipaddr, err := resolveDomain(m, c, dnsServer)
		if err != nil {
			//log.Printf("[%s]===>[%s][%s] ResolveDomain error:%s", dnsServer, queryType, domain, err.Error())
			if strings.Contains(err.Error(), "some name that ought to exist, does not exist (NXDOMAIN)") {
				noExistTimes++
				if noExistTimes >= 4 {
					return "", fmt.Errorf("解析域名[%s][%s]IP失败:noExistTimes", queryType, domain)
				}
			}
			continue
		}
		return ipaddr, nil
	}

	return "", fmt.Errorf("解析域名[%s][%s]IP失败", queryType, domain)
}

func resolveDomain(msg *dns.Msg, client *dns.Client, dnsServer string) (string, error) {

Redo:
	if in, _, err := client.Exchange(msg, dnsServer); err == nil { // Second return value is RTT, not used for now
		if in.MsgHdr.Truncated {
			client.Net = "tcp"
			goto Redo
		}

		switch in.MsgHdr.Rcode {
		case dns.RcodeServerFailure:
			return "", fmt.Errorf("the name server encountered an internal failure while processing this request (SERVFAIL)")
		case dns.RcodeNameError:
			return "", fmt.Errorf("some name that ought to exist, does not exist (NXDOMAIN)")
		case dns.RcodeRefused:
			return "", fmt.Errorf("the name server refuses to perform the specified operation for policy or security reasons (REFUSED)")
		default:
			//fmt.Printf("in.Answer len:%d\n", len(in.Answer))
			for _, rr := range in.Answer {
				//fmt.Printf("rr.String :%s\n", rr.String())
				return strings.Replace(rr.String(), rr.Header().String(), "", -1), nil
			}
		}
	}
	return "", fmt.Errorf("DNS server could not be reached")
}
