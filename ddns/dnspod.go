package ddns

import (
	"fmt"
	"net/url"

	"github.com/gdy666/lucky/ddnscore.go"
	"github.com/gdy666/lucky/thirdlib/gdylib/httputils"
)

const (
	recordListAPI   string = "https://dnsapi.cn/Record.List"
	recordModifyURL string = "https://dnsapi.cn/Record.Modify"
	recordCreateAPI string = "https://dnsapi.cn/Record.Create"
)

// https://cloud.tencent.com/document/api/302/8516
// Dnspod 腾讯云dns实现
type Dnspod struct {
	DNSCommon
	TTL string
}

// DnspodRecordListResp recordListAPI结果
type DnspodRecordListResp struct {
	DnspodStatus
	Records []struct {
		ID      string
		Name    string
		Type    string
		Value   string
		Enabled string
	}
}

// DnspodStatus DnspodStatus
type DnspodStatus struct {
	Status struct {
		Code    string
		Message string
	}
}

// Init 初始化
func (dnspod *Dnspod) Init(task *ddnscore.DDNSTaskInfo) {
	dnspod.DNSCommon.Init(task)

	if task.TTL == "" {
		// 默认600s
		dnspod.TTL = "600"
	} else {
		dnspod.TTL = task.TTL
	}
	dnspod.SetCreateUpdateDomainFunc(dnspod.createUpdateDomain)
}

func (dnspod *Dnspod) createUpdateDomain(recordType, ipAddr string, domain *ddnscore.Domain) {
	result, err := dnspod.getRecordList(domain, recordType)
	if err != nil {
		errMsg := "更新失败[001]:\n"
		errMsg += err.Error()
		domain.SetDomainUpdateStatus(ddnscore.UpdatedFailed, errMsg)
		return
	}

	if len(result.Records) > 0 {
		// 更新
		dnspod.modify(result, domain, recordType, ipAddr)
	} else {
		// 新增
		dnspod.create(result, domain, recordType, ipAddr)
	}
}

// 创建
func (dnspod *Dnspod) create(result DnspodRecordListResp, domain *ddnscore.Domain, recordType string, ipAddr string) {
	params := domain.GetCustomParams()
	params.Add("login_token", dnspod.task.DNS.ID+","+dnspod.task.DNS.Secret)
	params.Add("domain", domain.DomainName)
	params.Add("sub_domain", domain.GetSubDomain())
	params.Add("record_type", recordType)
	params.Add("value", ipAddr)
	params.Add("ttl", dnspod.TTL)
	params.Add("format", "json")

	if !params.Has("record_line") {
		params.Add("record_line", "默认")
	}

	status, err := dnspod.commonRequest(recordCreateAPI, params, domain)
	if err == nil && status.Status.Code == "1" {
		//log.Printf("新增域名解析 %s 成功！IP: %s", domain, ipAddr)
		domain.SetDomainUpdateStatus(ddnscore.UpdatedSuccess, "")
	} else {
		errMsg := fmt.Sprintf("创建域名失败:%v\n", status)
		if err != nil {
			errMsg += err.Error()
		}
		domain.SetDomainUpdateStatus(ddnscore.UpdatedFailed, errMsg)
	}
}

// 修改
func (dnspod *Dnspod) modify(result DnspodRecordListResp, domain *ddnscore.Domain, recordType string, ipAddr string) {
	for _, record := range result.Records {
		// 相同不修改
		if record.Value == ipAddr {
			if domain.UpdateStatus == ddnscore.UpdatedFailed {
				domain.SetDomainUpdateStatus(ddnscore.UpdatedSuccess, "")
			} else {
				domain.SetDomainUpdateStatus(ddnscore.UpdatedNothing, "")
			}
			continue
		}
		params := domain.GetCustomParams()
		params.Add("login_token", dnspod.task.DNS.ID+","+dnspod.task.DNS.Secret)
		params.Add("domain", domain.DomainName)
		params.Add("sub_domain", domain.GetSubDomain())
		params.Add("record_type", recordType)
		params.Add("value", ipAddr)
		params.Add("ttl", dnspod.TTL)
		params.Add("format", "json")
		params.Add("record_id", record.ID)

		if !params.Has("record_line") {
			params.Add("record_line", "默认")
		}
		status, err := dnspod.commonRequest(recordModifyURL, params, domain)
		if err == nil && status.Status.Code == "1" {
			//log.Printf("更新域名解析 %s 成功！IP: %s", domain, ipAddr)
			domain.SetDomainUpdateStatus(ddnscore.UpdatedSuccess, "")
		} else {
			errMsg := fmt.Sprintf("更新域名解析失败:%v\n", status)
			if err != nil {
				errMsg += err.Error()
			}
			domain.SetDomainUpdateStatus(ddnscore.UpdatedFailed, errMsg)
		}
	}
}

// 公共
func (dnspod *Dnspod) commonRequest(apiAddr string, values url.Values, domain *ddnscore.Domain) (status DnspodStatus, err error) {
	client, e := dnspod.CreateHTTPClient()
	if e != nil {
		err = e
		return
	}
	resp, e := client.PostForm(
		apiAddr,
		values,
	)

	if e != nil {
		err = e
		return
	}

	err = httputils.GetAndParseJSONResponseFromHttpResponse(resp, &status)

	return
}

// 获得域名记录列表
func (dnspod *Dnspod) getRecordList(domain *ddnscore.Domain, typ string) (result DnspodRecordListResp, err error) {
	params := domain.GetCustomParams()
	params.Add("login_token", dnspod.task.DNS.ID+","+dnspod.task.DNS.Secret)
	params.Add("domain", domain.DomainName)
	params.Add("record_type", typ)
	params.Add("sub_domain", domain.GetSubDomain())
	params.Add("format", "json")

	if !params.Has("record_line") {
		params.Add("record_line", "默认")
	}

	client, e := dnspod.CreateHTTPClient()
	if e != nil {
		err = e
		return
	}

	resp, err := client.PostForm(
		recordListAPI,
		params,
	)

	if err != nil {
		err = e
		return
	}

	err = httputils.GetAndParseJSONResponseFromHttpResponse(resp, &result)

	return
}
