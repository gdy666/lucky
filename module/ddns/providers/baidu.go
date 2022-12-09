package providers

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/gdy666/lucky/module/ddns/ddnscore.go"
	"github.com/gdy666/lucky/thirdlib/gdylib/httputils"
	"github.com/gdy666/lucky/thirdlib/jeessy2/ddns-go/util"
)

// https://cloud.baidu.com/doc/BCD/s/4jwvymhs7

const (
	baiduEndpoint = "https://bcd.baidubce.com"
)

type BaiduCloud struct {
	ProviderCommon
	TTL int
}

// BaiduRecord 单条解析记录
type BaiduRecord struct {
	RecordId uint   `json:"recordId"`
	Domain   string `json:"domain"`
	View     string `json:"view"`
	Rdtype   string `json:"rdtype"`
	TTL      int    `json:"ttl"`
	Rdata    string `json:"rdata"`
	ZoneName string `json:"zoneName"`
	Status   string `json:"status"`
}

// BaiduRecordsResp 获取解析列表拿到的结果
type BaiduRecordsResp struct {
	TotalCount int           `json:"totalCount"`
	Result     []BaiduRecord `json:"result"`
}

// BaiduListRequest 获取解析列表请求的body json
type BaiduListRequest struct {
	Domain   string `json:"domain"`
	PageNum  int    `json:"pageNum"`
	PageSize int    `json:"pageSize"`
}

// BaiduModifyRequest 修改解析请求的body json
type BaiduModifyRequest struct {
	RecordId uint   `json:"recordId"`
	Domain   string `json:"domain"`
	View     string `json:"view"`
	RdType   string `json:"rdType"`
	TTL      int    `json:"ttl"`
	Rdata    string `json:"rdata"`
	ZoneName string `json:"zoneName"`
}

// BaiduCreateRequest 创建新解析请求的body json
type BaiduCreateRequest struct {
	Domain   string `json:"domain"`
	RdType   string `json:"rdType"`
	TTL      int    `json:"ttl"`
	Rdata    string `json:"rdata"`
	ZoneName string `json:"zoneName"`
}

func (baidu *BaiduCloud) Init(task *ddnscore.DDNSTaskInfo) {
	baidu.ProviderCommon.Init(task)

	if task.TTL == "" {
		// 默认300s
		baidu.TTL = 300
	} else {
		ttl, err := strconv.Atoi(task.TTL)
		if err != nil {
			baidu.TTL = 300
		} else {
			baidu.TTL = ttl
		}
	}
	baidu.SetCreateUpdateDomainFunc(baidu.createUpdateDomain)
}

func (baidu *BaiduCloud) createUpdateDomain(recordType, ipAddr string, domain *ddnscore.Domain) {
	var records BaiduRecordsResp

	requestBody := BaiduListRequest{
		Domain:   domain.DomainName,
		PageNum:  1,
		PageSize: 1000,
	}

	err := baidu.request("POST", baiduEndpoint+"/v1/domain/resolve/list", requestBody, &records)
	if err != nil {
		errMsg := "更新失败[001]:\n"
		errMsg += err.Error()
		domain.SetDomainUpdateStatus(ddnscore.UpdatedFailed, errMsg)
		return
	}

	find := false
	for _, record := range records.Result {
		if record.Domain == domain.GetSubDomain() {
			//存在就去更新
			baidu.modify(record, domain, recordType, ipAddr)
			find = true
			break
		}
	}
	if !find {
		//没找到，去创建
		baidu.create(domain, recordType, ipAddr)
	}
}

// create 创建新的解析
func (baidu *BaiduCloud) create(domain *ddnscore.Domain, recordType string, ipAddr string) {
	var baiduCreateRequest = BaiduCreateRequest{
		Domain:   domain.GetSubDomain(), //处理一下@
		RdType:   recordType,
		TTL:      baidu.TTL,
		Rdata:    ipAddr,
		ZoneName: domain.DomainName,
	}
	var result BaiduRecordsResp

	err := baidu.request("POST", baiduEndpoint+"/v1/domain/resolve/add", baiduCreateRequest, &result)
	if err == nil {
		//log.Printf("新增域名解析 %s 成功！IP: %s", domain, ipAddr)
		domain.SetDomainUpdateStatus(ddnscore.UpdatedSuccess, "")
	} else {
		//log.Printf("新增域名解析 %s 失败！", domain)
		domain.SetDomainUpdateStatus(ddnscore.UpdatedFailed, err.Error())
	}
}

// modify 更新解析
func (baidu *BaiduCloud) modify(record BaiduRecord, domain *ddnscore.Domain, rdType string, ipAddr string) {
	//没有变化直接跳过
	if record.Rdata == ipAddr {
		if domain.UpdateStatus == ddnscore.UpdatedFailed {
			domain.SetDomainUpdateStatus(ddnscore.UpdatedSuccess, "")
		} else {
			domain.SetDomainUpdateStatus(ddnscore.UpdatedNothing, "")
		}
		return
	}
	var baiduModifyRequest = BaiduModifyRequest{
		RecordId: record.RecordId,
		Domain:   record.Domain,
		View:     record.View,
		RdType:   rdType,
		TTL:      record.TTL,
		Rdata:    ipAddr,
		ZoneName: record.ZoneName,
	}
	var result BaiduRecordsResp

	err := baidu.request("POST", baiduEndpoint+"/v1/domain/resolve/edit", baiduModifyRequest, &result)
	if err == nil {
		//log.Printf("更新域名解析 %s 成功！IP: %s", domain, ipAddr)
		domain.SetDomainUpdateStatus(ddnscore.UpdatedSuccess, "")
	} else {
		//log.Printf("更新域名解析 %s 失败！", domain)
		domain.SetDomainUpdateStatus(ddnscore.UpdatedFailed, err.Error())
	}
}

// request 统一请求接口
func (baidu *BaiduCloud) request(method string, url string, data interface{}, result interface{}) (err error) {
	jsonStr := make([]byte, 0)
	if data != nil {
		jsonStr, _ = json.Marshal(data)
	}

	req, err := http.NewRequest(
		method,
		url,
		bytes.NewBuffer(jsonStr),
	)

	if err != nil {
		log.Println("http.NewRequest失败. Error: ", err)
		return
	}

	util.BaiduSigner(baidu.task.DNS.ID, baidu.task.DNS.Secret, req)

	client, err := baidu.CreateHTTPClient()
	if err != nil {
		return err
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	return httputils.GetAndParseJSONResponseFromHttpResponse(resp, result)
}
