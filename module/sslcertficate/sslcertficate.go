package ssl

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/gdy666/lucky/config"
	sslconf "github.com/gdy666/lucky/module/sslcertficate/conf"
	"github.com/gdy666/lucky/thirdlib/gdylib/stringsp"
)

func GetCertDomains(cert *tls.Certificate) []string {
	var res []string
	if cert == nil {
		return res
	}
	for i := range cert.Certificate {
		xx, err := x509.ParseCertificate(cert.Certificate[i])
		if err != nil {
			continue
		}
		for j := range xx.DNSNames {
			d := strings.TrimSpace(xx.DNSNames[j])
			if d == "" {
				continue
			}
			res = append(res, d)
		}
	}
	return res
}

func GetDomainsStrByDomains(dst []string) string {
	var res strings.Builder
	for i := range dst {
		d := strings.TrimSpace(dst[i])
		if d == "" {
			continue
		}
		if res.Len() > 0 {
			res.WriteString(",")
		}
		res.WriteString(d)
	}
	return res.String()
}

//---------------------------------

func SSLCertficateListInit() {
	config.ConfigureMutex.RLock()
	defer config.ConfigureMutex.RUnlock()
	var err error
	for i := range config.Configure.SSLCertficateList {
		err = config.Configure.SSLCertficateList[i].Init()
		if err != nil {
			log.Printf("SSLCertficateListInit [%s]err:%s", config.Configure.SSLCertficateList[i].Key, err.Error())
		}
	}
}

func GetSSLCertficateList() []sslconf.SSLCertficate {
	config.ConfigureMutex.RLock()
	defer config.ConfigureMutex.RUnlock()
	var res []sslconf.SSLCertficate
	if config.Configure == nil {
		return res
	}

	for i := range config.Configure.SSLCertficateList {
		res = append(res, config.Configure.SSLCertficateList[i])
	}
	return res
}

func SSLCertficateListAdd(s *sslconf.SSLCertficate) error {
	config.ConfigureMutex.Lock()
	defer config.ConfigureMutex.Unlock()

	//************
	//重复检测
	for i := range config.Configure.SSLCertficateList {
		if config.Configure.SSLCertficateList[i].CertBase64 == s.CertBase64 {
			return fmt.Errorf("绑定域名[%s]的相同证书已存在,请勿重复添加", (*s.CertsInfo)[0].Domains[0])
		}

		if config.Configure.SSLCertficateList[i].GetFirstDomain() != "" &&
			config.Configure.SSLCertficateList[i].GetFirstDomain() == s.GetFirstDomain() {
			return fmt.Errorf("绑定域名[%s]的证书已存在,如果要添加新证书请先手动删除旧证书", (*s.CertsInfo)[0].Domains[0])
		}
	}

	//************

	if s.Key == "" {
		s.Key = stringsp.GetRandomString(8)
	}
	s.AddTime = time.Now().Format("2006-01-02 15:04:05")
	s.Enable = true
	config.Configure.SSLCertficateList = append(config.Configure.SSLCertficateList, *s)
	return config.Save()
}

func SSLCertficateListDelete(key string) error {
	config.ConfigureMutex.Lock()
	defer config.ConfigureMutex.Unlock()
	deleteIndex := -1

	for i := range config.Configure.SSLCertficateList {
		if config.Configure.SSLCertficateList[i].Key == key {
			deleteIndex = i
			break
		}
	}

	if deleteIndex < 0 {
		return fmt.Errorf("key:%s 不存在", key)
	}
	config.Configure.SSLCertficateList = DeleteSSLCertficateListslice(config.Configure.SSLCertficateList, deleteIndex)
	return config.Save()
}

func SSLCertficateEnable(key string, enable bool) error {
	config.ConfigureMutex.Lock()
	defer config.ConfigureMutex.Unlock()
	index := -1
	for i := range config.Configure.SSLCertficateList {
		if config.Configure.SSLCertficateList[i].Key == key {
			index = i
			break
		}
	}
	if index < 0 {
		return fmt.Errorf("key:%s 不存在", key)
	}
	config.Configure.SSLCertficateList[index].Enable = enable
	return config.Save()
}

func SSLCertficateAlterRemark(key, remark string) error {
	config.ConfigureMutex.Lock()
	defer config.ConfigureMutex.Unlock()
	index := -1
	for i := range config.Configure.SSLCertficateList {
		if config.Configure.SSLCertficateList[i].Key == key {
			index = i
			break
		}
	}
	if index < 0 {
		return fmt.Errorf("key:%s 不存在", key)
	}
	config.Configure.SSLCertficateList[index].Remark = remark
	return config.Save()
}

func DeleteSSLCertficateListslice(a []sslconf.SSLCertficate, deleteIndex int) []sslconf.SSLCertficate {
	j := 0
	for i := range a {
		if i != deleteIndex {
			a[j] = a[i]
			j++
		}
	}
	return a[:j]
}

func GetValidSSLCertficateList() []tls.Certificate {
	var res []tls.Certificate
	var gdnRes []tls.Certificate
	sslListCache := GetSSLCertficateList()
	for _, s := range sslListCache {
		if !s.Enable {
			continue
		}
		if strings.HasPrefix(s.GetFirstDomain(), "*.") {
			gdnRes = append(gdnRes, *s.Certificate)
			continue
		}
		res = append(res, *s.Certificate)
	}
	res = append(res, gdnRes...)

	return res
}
