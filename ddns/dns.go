package ddns

import "github.com/gdy666/lucky/ddnscore.go"

// DNS interface
type DNS interface {
	Init(task *ddnscore.DDNSTaskInfo)
	// 添加或更新IPv4/IPv6记录
	AddUpdateDomainRecords() string
}
