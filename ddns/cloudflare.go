package ddns

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gdy666/lucky/ddnscore.go"
	"github.com/gdy666/lucky/thirdlib/gdylib/httputils"
)

const (
	zonesAPI string = "https://api.cloudflare.com/client/v4/zones"
)

// Cloudflare Cloudflare实现
type Cloudflare struct {
	DNSCommon
	TTL int
}

// CloudflareZonesResp cloudflare zones返回结果
type CloudflareZonesResp struct {
	CloudflareStatus
	Result []struct {
		ID     string
		Name   string
		Status string
		Paused bool
	}
}

// CloudflareRecordsResp records
type CloudflareRecordsResp struct {
	CloudflareStatus
	Result []CloudflareRecord
}

// CloudflareRecord 记录实体
type CloudflareRecord struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Type    string `json:"type"`
	Content string `json:"content"`
	Proxied bool   `json:"proxied"`
	TTL     int    `json:"ttl"`
}

// CloudflareStatus 公共状态
type CloudflareStatus struct {
	Success  bool
	Messages []string
}

// Init 初始化
func (cf *Cloudflare) Init(task *ddnscore.DDNSTaskInfo) {
	cf.DNSCommon.Init(task)

	if task.TTL == "" {
		// 默认1 auto ttl
		cf.TTL = 1
	} else {
		ttl, err := strconv.Atoi(task.TTL)
		if err != nil {
			cf.TTL = 1
		} else {
			cf.TTL = ttl
		}
	}
	cf.SetCreateUpdateDomainFunc(cf.createUpdateDomain)
}

func (cf *Cloudflare) createUpdateDomain(recordType, ipAddr string, domain *ddnscore.Domain) {
	result, err := cf.getZones(domain)
	if err != nil || len(result.Result) != 1 {
		errMsg := "更新失败[001]:\n"
		if err != nil {
			errMsg += err.Error()
		} else {
			errMsg += fmt.Sprintf("%v", result)
		}
		domain.SetDomainUpdateStatus(ddnscore.UpdatedFailed, errMsg)
		return
	}
	zoneID := result.Result[0].ID

	var records CloudflareRecordsResp
	// getDomains 最多更新前50条
	err = cf.request(
		"GET",
		fmt.Sprintf(zonesAPI+"/%s/dns_records?type=%s&name=%s&per_page=50", zoneID, recordType, domain),
		nil,
		&records,
	)

	if err != nil || !records.Success {
		errMsg := "更新失败[002]:\n"
		if err != nil {
			errMsg += err.Error()
		}
		domain.SetDomainUpdateStatus(ddnscore.UpdatedFailed, errMsg)
		return
	}

	if len(records.Result) > 0 {
		// 更新
		cf.modify(records, zoneID, domain, recordType, ipAddr)
	} else {
		// 新增
		cf.create(zoneID, domain, recordType, ipAddr)
	}
}

// 创建
func (cf *Cloudflare) create(zoneID string, domain *ddnscore.Domain, recordType string, ipAddr string) {

	record := &CloudflareRecord{
		Type:    recordType,
		Name:    domain.String(),
		Content: ipAddr,
		Proxied: false,
		TTL:     cf.TTL,
	}

	var status CloudflareStatus
	err := cf.request(
		"POST",
		fmt.Sprintf(zonesAPI+"/%s/dns_records", zoneID),
		record,
		&status,
	)
	if err == nil && status.Success {
		//log.Printf("新增域名解析 %s 成功！IP: %s", domain, ipAddr)
		domain.SetDomainUpdateStatus(ddnscore.UpdatedSuccess, "")
	} else {
		errMsg := fmt.Sprintf("创建域名失败:\n%v\n", status)
		if err != nil {
			errMsg += fmt.Sprintf(":%s", err.Error())
		}
		domain.SetDomainUpdateStatus(ddnscore.UpdatedFailed, errMsg)
	}
}

// 修改
func (cf *Cloudflare) modify(result CloudflareRecordsResp, zoneID string, domain *ddnscore.Domain, recordType string, ipAddr string) {

	for _, record := range result.Result {
		// 相同不修改
		if record.Content == ipAddr {
			if domain.UpdateStatus == ddnscore.UpdatedFailed {
				domain.SetDomainUpdateStatus(ddnscore.UpdatedSuccess, "")
			} else {
				domain.SetDomainUpdateStatus(ddnscore.UpdatedNothing, "")
			}
			continue
		}
		var status CloudflareStatus
		record.Content = ipAddr
		record.TTL = cf.TTL

		err := cf.request(
			"PUT",
			fmt.Sprintf(zonesAPI+"/%s/dns_records/%s", zoneID, record.ID),
			record,
			&status,
		)

		if err == nil && status.Success {
			//log.Printf("更新域名解析 %s 成功！IP: %s", domain, ipAddr)
			domain.SetDomainUpdateStatus(ddnscore.UpdatedSuccess, "")
		} else {
			//log.Printf("更新域名解析 %s 失败！Messages: %s", domain, status.Messages)
			errMsg := fmt.Sprintf("更新域名解析失败:%v\n", status)
			if err != nil {
				errMsg += err.Error()
			}
			domain.SetDomainUpdateStatus(ddnscore.UpdatedFailed, errMsg)
		}
	}
}

// 获得域名记录列表
func (cf *Cloudflare) getZones(domain *ddnscore.Domain) (result CloudflareZonesResp, err error) {
	err = cf.request(
		"GET",
		fmt.Sprintf(zonesAPI+"?name=%s&status=%s&per_page=%s", domain.DomainName, "active", "50"),
		nil,
		&result,
	)

	return
}

// request 统一请求接口
func (cf *Cloudflare) request(method string, url string, data interface{}, result interface{}) (err error) {
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
	req.Header.Set("Authorization", "Bearer "+cf.task.DNS.Secret)
	req.Header.Set("Content-Type", "application/json")

	client, err := cf.CreateHTTPClient()
	if err != nil {
		return err
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	return httputils.GetAndParseJSONResponseFromHttpResponse(resp, result)
}
