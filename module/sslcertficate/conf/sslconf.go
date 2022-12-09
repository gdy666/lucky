package sslconf

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"fmt"
	"strings"
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

type CertInfo struct {
	Domains       []string
	NotBeforeTime string `json:"NotBeforeTime"` //time.Time
	NotAfterTime  string `json:"NotAfterTime"`  //time.Time
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
