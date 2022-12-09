package providers

import "github.com/gdy666/lucky/module/ddns/ddnscore.go"

// Provider interface
type Provider interface {
	Init(task *ddnscore.DDNSTaskInfo)
	// 添加或更新IPv4/IPv6记录
	AddUpdateDomainRecords() string
}
