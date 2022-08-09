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
	"github.com/gdy666/lucky/thirdlib/jeessy2/ddns-go/util"
)

const (
	huaweicloudEndpoint string = "https://dns.myhuaweicloud.com"
)

// https://support.huaweicloud.com/api-dns/dns_api_64001.html
// Huaweicloud Huaweicloud
type Huaweicloud struct {
	DNSCommon
	TTL int
}

// HuaweicloudZonesResp zones response
type HuaweicloudZonesResp struct {
	Zones []struct {
		ID         string
		Name       string
		Recordsets []HuaweicloudRecordsets
	}
}

// HuaweicloudRecordsResp 记录返回结果
type HuaweicloudRecordsResp struct {
	Recordsets []HuaweicloudRecordsets
}

// HuaweicloudRecordsets 记录
type HuaweicloudRecordsets struct {
	ID      string
	Name    string `json:"name"`
	ZoneID  string `json:"zone_id"`
	Status  string
	Type    string   `json:"type"`
	TTL     int      `json:"ttl"`
	Records []string `json:"records"`
}

// Init 初始化
func (hw *Huaweicloud) Init(task *ddnscore.DDNSTaskInfo) {
	hw.DNSCommon.Init(task)

	if task.TTL == "" {
		// 默认300s
		hw.TTL = 300
	} else {
		ttl, err := strconv.Atoi(task.TTL)
		if err != nil {
			hw.TTL = 300
		} else {
			hw.TTL = ttl
		}
	}
	hw.SetCreateUpdateDomainFunc(hw.createUpdateDomain)
}

func (hw *Huaweicloud) createUpdateDomain(recordType, ipAddr string, domain *ddnscore.Domain) {
	var records HuaweicloudRecordsResp

	err := hw.request(
		"GET",
		fmt.Sprintf(huaweicloudEndpoint+"/v2/recordsets?type=%s&name=%s", recordType, domain),
		nil,
		&records,
	)

	if err != nil {
		errMsg := "更新失败[001]:\n"
		errMsg += err.Error()
		domain.SetDomainUpdateStatus(ddnscore.UpdatedFailed, errMsg)
		return
	}

	find := false
	for _, record := range records.Recordsets {
		// 名称相同才更新。华为云默认是模糊搜索
		if record.Name == domain.String()+"." {
			// 更新
			hw.modify(record, domain, recordType, ipAddr)
			find = true
			break
		}
	}

	if !find {
		// 新增
		hw.create(domain, recordType, ipAddr)
	}
}

// 创建
func (hw *Huaweicloud) create(domain *ddnscore.Domain, recordType string, ipAddr string) {
	zone, err := hw.getZones(domain)
	if err != nil {
		return
	}
	if len(zone.Zones) == 0 {
		log.Println("未能找到公网域名, 请检查域名是否添加")
		return
	}

	zoneID := zone.Zones[0].ID
	for _, z := range zone.Zones {
		if z.Name == domain.DomainName+"." {
			zoneID = z.ID
			break
		}
	}

	record := &HuaweicloudRecordsets{
		Type:    recordType,
		Name:    domain.String() + ".",
		Records: []string{ipAddr},
		TTL:     hw.TTL,
	}
	var result HuaweicloudRecordsets
	err = hw.request(
		"POST",
		fmt.Sprintf(huaweicloudEndpoint+"/v2/zones/%s/recordsets", zoneID),
		record,
		&result,
	)
	if err == nil && (len(result.Records) > 0 && result.Records[0] == ipAddr) {
		//log.Printf("新增域名解析 %s 成功！IP: %s", domain, ipAddr)
		domain.SetDomainUpdateStatus(ddnscore.UpdatedSuccess, "")
	} else {
		errMsg := fmt.Sprintf("新增域名失败:%v\n", result)
		if err != nil {
			errMsg += err.Error()
		}
		domain.SetDomainUpdateStatus(ddnscore.UpdatedFailed, errMsg)
	}
}

// 修改
func (hw *Huaweicloud) modify(record HuaweicloudRecordsets, domain *ddnscore.Domain, recordType string, ipAddr string) {

	// 相同不修改
	if len(record.Records) > 0 && record.Records[0] == ipAddr {
		if domain.UpdateStatus == ddnscore.UpdatedFailed {
			domain.SetDomainUpdateStatus(ddnscore.UpdatedSuccess, "")
		} else {
			domain.SetDomainUpdateStatus(ddnscore.UpdatedNothing, "")
		}
		return
	}

	var request map[string]interface{} = make(map[string]interface{})
	request["records"] = []string{ipAddr}
	request["ttl"] = hw.TTL

	var result HuaweicloudRecordsets

	err := hw.request(
		"PUT",
		fmt.Sprintf(huaweicloudEndpoint+"/v2/zones/%s/recordsets/%s", record.ZoneID, record.ID),
		&request,
		&result,
	)

	if err == nil && (len(result.Records) > 0 && result.Records[0] == ipAddr) {
		//log.Printf("更新域名解析 %s 成功！IP: %s, 状态: %s", domain, ipAddr, result.Status)
		domain.SetDomainUpdateStatus(ddnscore.UpdatedSuccess, "")
	} else {
		errMsg := fmt.Sprintf("更新域名解析:%v\n", result)
		if err != nil {
			errMsg += err.Error()
		}
		domain.SetDomainUpdateStatus(ddnscore.UpdatedFailed, errMsg)
	}
}

// 获得域名记录列表
func (hw *Huaweicloud) getZones(domain *ddnscore.Domain) (result HuaweicloudZonesResp, err error) {
	err = hw.request(
		"GET",
		fmt.Sprintf(huaweicloudEndpoint+"/v2/zones?name=%s", domain.DomainName),
		nil,
		&result,
	)

	return
}

// request 统一请求接口
func (hw *Huaweicloud) request(method string, url string, data interface{}, result interface{}) (err error) {
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

	s := util.Signer{
		Key:    hw.task.DNS.ID,
		Secret: hw.task.DNS.Secret,
	}
	s.Sign(req)

	req.Header.Add("content-type", "application/json")

	client, err := hw.CreateHTTPClient()
	if err != nil {
		return err
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	return httputils.GetAndParseJSONResponseFromHttpResponse(resp, result)
}
