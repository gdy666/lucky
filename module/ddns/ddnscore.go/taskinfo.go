package ddnscore

import (
	ddnsconf "github.com/gdy666/lucky/module/ddns/conf"
)

type DDNSTaskInfo struct {
	ddnsconf.DDNSTask
	TaskState DDNSTaskState `json:"TaskState"`
}

// CheckIPChange 检测公网IP是否改变
func (d *DDNSTaskInfo) CheckIPChange() (ipAddr string, change bool) {
	ipAddr = d.GetIpAddr()
	checkIPChange := d.TaskState.IPChanged(ipAddr)
	if checkIPChange {
		return ipAddr, true
	}
	//IP没变化
	return ipAddr, false
}

func (d *DDNSTaskInfo) SyncDomains() {
	if d.ModifyTime == d.TaskState.ModifyTime {
		//fmt.Printf("不需要syncDomains\n")
		return
	}
	//fmt.Printf("需要syncDomains\n")
	domains, _ := checkParseDomains(d.Domains)

	for i := range domains {
		index := getDomainIndex(d.TaskState.Domains, &domains[i])
		if index < 0 {
			continue
		}
		domains[i] = d.TaskState.Domains[index]
	}
	d.TaskState.Domains = domains
	d.TaskState.ModifyTime = d.ModifyTime
	taskInfoMap.Store(d.TaskKey, &d.TaskState)
}

func getDomainIndex(domains []Domain, domain *Domain) (index int) {
	index = -1
	for i := range domains {
		if domains[i].RawStr == domain.RawStr {
			index = i
			return
		}
	}
	return
}
