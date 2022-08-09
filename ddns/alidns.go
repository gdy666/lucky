package ddns

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"net/url"

	"github.com/gdy666/lucky/ddnscore.go"
	"github.com/gdy666/lucky/thirdlib/gdylib/httputils"
	"github.com/gdy666/lucky/thirdlib/jeessy2/ddns-go/util"
)

const (
	alidnsEndpoint string = "https://alidns.aliyuncs.com/"
)

// https://help.aliyun.com/document_detail/29776.html?spm=a2c4g.11186623.6.672.715a45caji9dMA
// Alidns Alidns
type Alidns struct {
	DNSCommon
	TTL string
}

// AlidnsRecord record
type AlidnsRecord struct {
	DomainName string
	RecordID   string
	Value      string
}

// AlidnsSubDomainRecords 记录
type AlidnsSubDomainRecords struct {
	TotalCount    int
	DomainRecords struct {
		Record []AlidnsRecord
	}
}

// AlidnsResp 修改/添加返回结果
type AlidnsResp struct {
	RecordID  string
	RequestID string
}

// Init 初始化
func (ali *Alidns) Init(task *ddnscore.DDNSTaskInfo) {
	ali.DNSCommon.Init(task)

	if task.TTL == "" {
		// 默认600s
		ali.TTL = "600"
	} else {
		ali.TTL = task.TTL
	}
	ali.SetCreateUpdateDomainFunc(ali.createUpdateDomain)
}

func (ali *Alidns) createUpdateDomain(recordType, ipAddr string, domain *ddnscore.Domain) {
	var records AlidnsSubDomainRecords
	// 获取当前域名信息
	params := domain.GetCustomParams()
	params.Set("Action", "DescribeSubDomainRecords")
	params.Set("DomainName", domain.DomainName)
	params.Set("SubDomain", domain.GetFullDomain())
	params.Set("Type", recordType)
	err := ali.request(params, &records)

	if err != nil {
		domain.SetDomainUpdateStatus(ddnscore.UpdatedFailed, err.Error())
		return
	}

	if records.TotalCount > 0 {
		// 默认第一个
		recordSelected := records.DomainRecords.Record[0]
		if params.Has("RecordId") {
			for i := 0; i < len(records.DomainRecords.Record); i++ {
				if records.DomainRecords.Record[i].RecordID == params.Get("RecordId") {
					recordSelected = records.DomainRecords.Record[i]
				}
			}
		}
		// 存在，更新
		ali.modify(recordSelected, domain, recordType, ipAddr)
	} else {
		// 不存在，创建
		ali.create(domain, recordType, ipAddr)
	}
}

// 创建
func (ali *Alidns) create(domain *ddnscore.Domain, recordType string, ipAddr string) {
	params := domain.GetCustomParams()
	params.Set("Action", "AddDomainRecord")
	params.Set("DomainName", domain.DomainName)
	params.Set("RR", domain.GetSubDomain())
	params.Set("Type", recordType)
	params.Set("Value", ipAddr)
	params.Set("TTL", ali.TTL)

	var result AlidnsResp
	err := ali.request(params, &result)

	if err == nil && result.RecordID != "" {
		domain.SetDomainUpdateStatus(ddnscore.UpdatedSuccess, "")
	} else {
		errMsg := fmt.Sprintf("创建域名失败:\n%v\n", result)
		if err != nil {
			errMsg += err.Error()
		}
		domain.SetDomainUpdateStatus(ddnscore.UpdatedFailed, errMsg)
	}
}

// 修改
func (ali *Alidns) modify(recordSelected AlidnsRecord, domain *ddnscore.Domain, recordType string, ipAddr string) {

	// 相同不修改
	if recordSelected.Value == ipAddr {
		if domain.UpdateStatus == ddnscore.UpdatedFailed {
			domain.SetDomainUpdateStatus(ddnscore.UpdatedSuccess, "")
		} else {
			domain.SetDomainUpdateStatus(ddnscore.UpdatedNothing, "")
		}
		return
	}

	params := domain.GetCustomParams()
	params.Set("Action", "UpdateDomainRecord")
	params.Set("RR", domain.GetSubDomain())
	params.Set("RecordId", recordSelected.RecordID)
	params.Set("Type", recordType)
	params.Set("Value", ipAddr)
	params.Set("TTL", ali.TTL)

	var result AlidnsResp
	err := ali.request(params, &result)

	if err == nil && result.RecordID != "" {
		domain.SetDomainUpdateStatus(ddnscore.UpdatedSuccess, "")
	} else {
		errMsg := fmt.Sprintf("更新域名解析失败:%v\n", result)
		if err != nil {
			errMsg += err.Error()
		}
		domain.SetDomainUpdateStatus(ddnscore.UpdatedFailed, errMsg)
	}
}

// request 统一请求接口
func (ali *Alidns) request(params url.Values, result interface{}) (err error) {

	util.AliyunSigner(ali.task.DNS.ID, ali.task.DNS.Secret, &params)

	req, err := http.NewRequest(
		"GET",
		alidnsEndpoint,
		bytes.NewBuffer(nil),
	)
	req.URL.RawQuery = params.Encode()

	if err != nil {
		log.Println("http.NewRequest失败. Error: ", err)
		return
	}

	client, err := ali.CreateHTTPClient()
	if err != nil {
		return err
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	return httputils.GetAndParseJSONResponseFromHttpResponse(resp, result)
}
