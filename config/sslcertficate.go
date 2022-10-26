package config

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/gdy666/lucky/thirdlib/gdylib/stringsp"
)

type SSLCertficate struct {
	Key        string      `json:"Key"`
	Enable     bool        `json:"Enable"`
	Remark     string      `json:"Remark"` //备注
	CertBase64 string      `json:"CertBase64"`
	KeyBase64  string      `json:"KeyBase64"`
	AddTime    string      `json:"AddTime"` //添加时间
	CertsInfo  *[]CertInfo `json:"-"`
	//---------------------
	Certificate *tls.Certificate `json:"-"`
}

func (s *SSLCertficate) Init() error {
	tc, err := CreateX509KeyPairByBase64Str(s.CertBase64, s.KeyBase64)
	if err != nil {
		return fmt.Errorf("CreateX509KeyPairByBase64Str error:%s", err.Error())
	}
	domainsInfo, err := GetCertDomainInfo(tc)
	if err != nil {
		return fmt.Errorf("GetCertDomainInfo error:%s", err.Error())
	}
	s.Certificate = tc
	s.CertsInfo = domainsInfo
	return nil
}

// GetOnlyDomain 返回证书第一条域名
func (s *SSLCertficate) GetFirstDomain() string {
	if s.CertsInfo == nil {
		return ""
	}
	if len(*s.CertsInfo) <= 0 {
		return ""
	}
	if len((*s.CertsInfo)[0].Domains) <= 0 {
		return ""
	}
	return (*s.CertsInfo)[0].Domains[0]
}

func CreateX509KeyPairByBase64Str(certBase64, keyBase64 string) (*tls.Certificate, error) {
	crtBytes, err := base64.StdEncoding.DecodeString(certBase64)
	if err != nil {
		return nil, fmt.Errorf("certBase64 decode error:%s", err.Error())
	}

	keyBytes, err := base64.StdEncoding.DecodeString(keyBase64)
	if err != nil {
		return nil, fmt.Errorf("keyBase64 decode error:%s", err.Error())
	}

	cert, err := tls.X509KeyPair(crtBytes, keyBytes)
	if err != nil {
		return nil, fmt.Errorf("create X509KeyPair error:%s", err.Error())
	}
	return &cert, nil
}

type CertInfo struct {
	Domains       []string
	NotBeforeTime string `json:"NotBeforeTime"` //time.Time
	NotAfterTime  string `json:"NotAfterTime"`  //time.Time
}

func GetCertDomainInfo(cert *tls.Certificate) (*[]CertInfo, error) {
	if cert == nil {
		return nil, fmt.Errorf("cert == nil")
	}

	var res []CertInfo

	for i := range cert.Certificate {
		xx, err := x509.ParseCertificate(cert.Certificate[i])
		if err != nil {
			continue
		}

		ds := GetDomainsTrimSpace(xx.DNSNames)
		if len(ds) == 0 {
			continue
		}

		info := CertInfo{Domains: ds, NotBeforeTime: xx.NotBefore.Format("2006-01-02 15:04:05"), NotAfterTime: xx.NotAfter.Format("2006-01-02 15:04:05")}
		res = append(res, info)
	}
	return &res, nil

}

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

// 除去空域名
func GetDomainsTrimSpace(dst []string) []string {
	var res []string
	for i := range dst {
		if strings.TrimSpace(dst[i]) == "" {
			continue
		}
		res = append(res, strings.TrimSpace(dst[i]))
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
	programConfigureMutex.RLock()
	defer programConfigureMutex.RUnlock()
	var err error
	for i := range programConfigure.SSLCertficateList {
		err = programConfigure.SSLCertficateList[i].Init()
		if err != nil {
			log.Printf("SSLCertficateListInit [%s]err:%s", programConfigure.SSLCertficateList[i].Key, err.Error())
		}
	}
}

func GetSSLCertficateList() []SSLCertficate {
	programConfigureMutex.RLock()
	defer programConfigureMutex.RUnlock()
	var res []SSLCertficate
	if programConfigure == nil {
		return res
	}

	for i := range programConfigure.SSLCertficateList {
		res = append(res, programConfigure.SSLCertficateList[i])
	}
	return res
}

func SSLCertficateListAdd(s *SSLCertficate) error {
	programConfigureMutex.Lock()
	defer programConfigureMutex.Unlock()

	//************
	//重复检测
	for i := range programConfigure.SSLCertficateList {
		if programConfigure.SSLCertficateList[i].CertBase64 == s.CertBase64 {
			return fmt.Errorf("绑定域名[%s]的相同证书已存在,请勿重复添加", (*s.CertsInfo)[0].Domains[0])
		}

		if programConfigure.SSLCertficateList[i].GetFirstDomain() != "" &&
			programConfigure.SSLCertficateList[i].GetFirstDomain() == s.GetFirstDomain() {
			return fmt.Errorf("绑定域名[%s]的证书已存在,如果要添加新证书请先手动删除旧证书", (*s.CertsInfo)[0].Domains[0])
		}
	}

	//************

	if s.Key == "" {
		s.Key = stringsp.GetRandomString(8)
	}
	s.AddTime = time.Now().Format("2006-01-02 15:04:05")
	s.Enable = true
	programConfigure.SSLCertficateList = append(programConfigure.SSLCertficateList, *s)
	return Save()
}

func SSLCertficateListDelete(key string) error {
	programConfigureMutex.Lock()
	defer programConfigureMutex.Unlock()
	deleteIndex := -1

	for i := range programConfigure.SSLCertficateList {
		if programConfigure.SSLCertficateList[i].Key == key {
			deleteIndex = i
			break
		}
	}

	if deleteIndex < 0 {
		return fmt.Errorf("key:%s 不存在", key)
	}
	programConfigure.SSLCertficateList = DeleteSSLCertficateListslice(programConfigure.SSLCertficateList, deleteIndex)
	return Save()
}

func SSLCertficateEnable(key string, enable bool) error {
	programConfigureMutex.Lock()
	defer programConfigureMutex.Unlock()
	index := -1
	for i := range programConfigure.SSLCertficateList {
		if programConfigure.SSLCertficateList[i].Key == key {
			index = i
			break
		}
	}
	if index < 0 {
		return fmt.Errorf("key:%s 不存在", key)
	}
	programConfigure.SSLCertficateList[index].Enable = enable
	return Save()
}

func SSLCertficateAlterRemark(key, remark string) error {
	programConfigureMutex.Lock()
	defer programConfigureMutex.Unlock()
	index := -1
	for i := range programConfigure.SSLCertficateList {
		if programConfigure.SSLCertficateList[i].Key == key {
			index = i
			break
		}
	}
	if index < 0 {
		return fmt.Errorf("key:%s 不存在", key)
	}
	programConfigure.SSLCertficateList[index].Remark = remark
	return Save()
}

func DeleteSSLCertficateListslice(a []SSLCertficate, deleteIndex int) []SSLCertficate {
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
