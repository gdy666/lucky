package ddnscore

import (
	"net/url"
	"time"
)

const (
	// UpdatedNothing 未改变
	UpdatedNothing string = "域名IP和公网IP一致"
	// UpdatedFailed 更新失败
	UpdatedFailed = "失败"
	// UpdatedSuccess 更新成功
	UpdatedSuccess = "成功"
	// UpdateStop 暂停
	UpdateStop = "停止同步"
	//UpdatePause 暂停 获取IP失败时暂停
	UpdatePause = "暂停同步"
	// UpdateWaiting
	UpdateWaiting = "等待更新"
)

// Domain 域名实体
type Domain struct {
	DomainName   string
	SubDomain    string
	CustomParams string

	UpdateStatus         string // 更新状态
	LastUpdateStatusTime string //最后更新状态的时间
	Message              string
	UpdateHistroy        []any
}

type UpdateHistroyItem struct {
	UpdateStatus string
	UpdateTime   string
}

func (d *Domain) String() string {
	if d.SubDomain != "" {
		return d.SubDomain + "." + d.DomainName
	}
	return d.DomainName
}

// GetFullDomain 返回完整子域名
func (d *Domain) GetFullDomain() string {
	if d.SubDomain != "" {
		return d.SubDomain + "." + d.DomainName
	}
	return "@" + "." + d.DomainName
}

// GetCustomParams not be nil
func (d *Domain) GetCustomParams() url.Values {
	if d.CustomParams != "" {
		q, err := url.ParseQuery(d.CustomParams)
		if err == nil {
			return q
		}
	}
	return url.Values{}
}

// GetSubDomain 获得子域名，为空返回@
// 阿里云，dnspod需要
func (d *Domain) GetSubDomain() string {
	if d.SubDomain != "" {
		return d.SubDomain
	}
	return "@"
}

func (d *Domain) SetDomainUpdateStatus(status string, message string) {

	if status != UpdateWaiting {
		if status != UpdateStop || d.UpdateStatus != UpdateStop {
			d.LastUpdateStatusTime = time.Now().Format("2006-01-02 15:04:05")
			// 状态更新历史记录
			hi := UpdateHistroyItem{UpdateStatus: string(status), UpdateTime: d.LastUpdateStatusTime}
			d.UpdateHistroy = append(d.UpdateHistroy, hi)
			if len(d.UpdateHistroy) > 10 {
				d.UpdateHistroy = DeleteAnyListlice(d.UpdateHistroy, 0)
			}
		}
	}
	d.UpdateStatus = status
	d.Message = message

}
