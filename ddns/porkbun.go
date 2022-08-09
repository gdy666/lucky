package ddns

import (
	"bytes"

	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gdy666/lucky/ddnscore.go"
	"github.com/gdy666/lucky/thirdlib/gdylib/httputils"
)

const (
	porkbunEndpoint string = "https://porkbun.com/api/json/v3/dns"
)

type Porkbun struct {
	DNSCommon
	TTL string
}
type PorkbunDomainRecord struct {
	Name    string `json:"name"`    // subdomain
	Type    string `json:"type"`    // record type, e.g. A AAAA CNAME
	Content string `json:"content"` // value
	Ttl     string `json:"ttl"`     // default 300
}

type PorkbunResponse struct {
	Status string `json:"status"`
}

type PorkbunDomainQueryResponse struct {
	*PorkbunResponse
	Records []PorkbunDomainRecord `json:"records"`
}

type PorkbunApiKey struct {
	AccessKey string `json:"apikey"`
	SecretKey string `json:"secretapikey"`
}

type PorkbunDomainCreateOrUpdateVO struct {
	*PorkbunApiKey
	*PorkbunDomainRecord
}

// Init 初始化
func (pb *Porkbun) Init(task *ddnscore.DDNSTaskInfo) {
	pb.DNSCommon.Init(task)
	if task.TTL == "" {
		// 默认600s
		pb.TTL = "600"
	} else {
		pb.TTL = task.TTL
	}
	pb.SetCreateUpdateDomainFunc(pb.createUpdateDomain)
}

func (pb *Porkbun) createUpdateDomain(recordType, ipAddr string, domain *ddnscore.Domain) {

	var record PorkbunDomainQueryResponse
	// 获取当前域名信息
	err := pb.request(
		porkbunEndpoint+fmt.Sprintf("/retrieveByNameType/%s/%s/%s", domain.DomainName, recordType, domain.SubDomain),
		&PorkbunApiKey{
			AccessKey: pb.task.DNS.ID,
			SecretKey: pb.task.DNS.Secret,
		},
		&record,
	)

	if err != nil {
		return
	}
	if record.Status == "SUCCESS" {
		if len(record.Records) > 0 {
			// 存在，更新
			pb.modify(&record, domain, recordType, ipAddr)
		} else {
			// 不存在，创建
			pb.create(domain, recordType, ipAddr)
		}
	} else {
		domain.SetDomainUpdateStatus(ddnscore.UpdatedFailed, "查询现有域名记录失败")
	}
}

// 创建
func (pb *Porkbun) create(domain *ddnscore.Domain, recordType string, ipAddr string) {
	var response PorkbunResponse

	err := pb.request(
		porkbunEndpoint+fmt.Sprintf("/create/%s", domain.DomainName),
		&PorkbunDomainCreateOrUpdateVO{
			PorkbunApiKey: &PorkbunApiKey{
				AccessKey: pb.task.DNS.ID,
				SecretKey: pb.task.DNS.Secret,
			},
			PorkbunDomainRecord: &PorkbunDomainRecord{
				Name:    domain.SubDomain,
				Type:    recordType,
				Content: ipAddr,
				Ttl:     pb.TTL,
			},
		},
		&response,
	)

	if err == nil && response.Status == "SUCCESS" {
		domain.SetDomainUpdateStatus(ddnscore.UpdatedSuccess, "")
	} else {
		errMsg := fmt.Sprintf("新增域名失败:%v\n", response)
		if err != nil {
			errMsg += err.Error()
		}
		domain.SetDomainUpdateStatus(ddnscore.UpdatedFailed, errMsg)
	}
}

// 修改
func (pb *Porkbun) modify(record *PorkbunDomainQueryResponse, domain *ddnscore.Domain, recordType string, ipAddr string) {

	// 相同不修改
	if len(record.Records) > 0 && record.Records[0].Content == ipAddr {
		if domain.UpdateStatus == ddnscore.UpdatedFailed {
			domain.SetDomainUpdateStatus(ddnscore.UpdatedSuccess, "")
		} else {
			domain.SetDomainUpdateStatus(ddnscore.UpdatedNothing, "")
		}
		return
	}

	var response PorkbunResponse

	err := pb.request(
		porkbunEndpoint+fmt.Sprintf("/editByNameType/%s/%s/%s", domain.DomainName, recordType, domain.SubDomain),
		&PorkbunDomainCreateOrUpdateVO{
			PorkbunApiKey: &PorkbunApiKey{
				AccessKey: pb.task.DNS.ID,
				SecretKey: pb.task.DNS.Secret,
			},
			PorkbunDomainRecord: &PorkbunDomainRecord{
				Content: ipAddr,
				Ttl:     pb.TTL,
			},
		},
		&response,
	)

	if err == nil && response.Status == "SUCCESS" {
		domain.SetDomainUpdateStatus(ddnscore.UpdatedSuccess, "")
	} else {
		errMsg := fmt.Sprintf("更新域名解析失败:%v\n", response)
		if err != nil {
			errMsg += err.Error()
		}
		domain.SetDomainUpdateStatus(ddnscore.UpdatedFailed, errMsg)
	}
}

// request 统一请求接口
func (pb *Porkbun) request(url string, data interface{}, result interface{}) (err error) {
	jsonStr := make([]byte, 0)
	if data != nil {
		jsonStr, _ = json.Marshal(data)
	}
	req, err := http.NewRequest(
		"POST",
		url,
		bytes.NewBuffer(jsonStr),
	)
	if err != nil {
		log.Println("http.NewRequest失败. Error: ", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")

	client, e := pb.CreateHTTPClient()
	if e != nil {
		err = e
		return
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	return httputils.GetAndParseJSONResponseFromHttpResponse(resp, result)
}
