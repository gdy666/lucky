package ddns

import "github.com/gdy666/lucky/config"

// DNS interface
type DNS interface {
	Init(task *config.DDNSTask)
	// 添加或更新IPv4/IPv6记录
	AddUpdateDomainRecords() string
}
