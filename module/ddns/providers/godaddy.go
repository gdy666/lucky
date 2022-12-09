package providers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gdy666/lucky/module/ddns/ddnscore.go"
	"github.com/gdy666/lucky/thirdlib/gdylib/httputils"
)

type godaddyRecord struct {
	Data string `json:"data"`
	Name string `json:"name"`
	TTL  int    `json:"ttl"`
	Type string `json:"type"`
}

type godaddyRecords []godaddyRecord

type GoDaddy struct {
	ProviderCommon
	TTL    int
	header http.Header
	client *http.Client
}

// Init 初始化
func (gd *GoDaddy) Init(task *ddnscore.DDNSTaskInfo) {
	gd.ProviderCommon.Init(task)
	// if task.TTL == "" {
	// 	// 默认600s
	// 	gd.TTL = 600
	// } else {
	// 	gd.TTL = task.TTL
	// }
	if task.TTL == "" {
		// 默认300s
		gd.TTL = 600
	} else {
		ttl, err := strconv.Atoi(task.TTL)
		if err != nil {
			gd.TTL = 600
		} else {
			gd.TTL = ttl
		}
	}
	gd.header = map[string][]string{
		"Authorization": {fmt.Sprintf("sso-key %s:%s", task.DNS.ID, task.DNS.Secret)},
		"Content-Type":  {"application/json"},
	}
	//g.throttle, _ = util.GetThrottle(55)
	gd.client, _ = gd.CreateHTTPClient()

	gd.SetCreateUpdateDomainFunc(gd.createUpdateDomain)
}

func (gd *GoDaddy) createUpdateDomain(recordType, ipAddr string, domain *ddnscore.Domain) {

	_, err := gd.sendReq(http.MethodPut, recordType, domain, &godaddyRecords{godaddyRecord{
		Data: ipAddr,
		Name: domain.SubDomain,
		TTL:  gd.TTL,
		Type: recordType,
	}})
	if err != nil {
		domain.SetDomainUpdateStatus(ddnscore.UpdatedFailed, err.Error())
	}
}

func (gd *GoDaddy) sendReq(method string, rType string, domain *ddnscore.Domain, data any) (*godaddyRecords, error) {

	var body *bytes.Buffer
	if data != nil {
		if buffer, err := json.Marshal(data); err != nil {
			return nil, err
		} else {
			body = bytes.NewBuffer(buffer)
		}
	}
	path := fmt.Sprintf("https://api.godaddy.com/v1/domains/%s/records/%s/%s",
		domain.DomainName, rType, domain.SubDomain)
	req, err := http.NewRequest(method, path, body)
	if err != nil {
		return nil, err
	}
	req.Header = gd.header

	resp, err := gd.client.Do(req)
	if err != nil {
		return nil, err
	}
	result := &godaddyRecords{}

	httputils.GetAndParseJSONResponseFromHttpResponse(resp, result)

	return result, nil
}
