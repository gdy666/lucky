package safeconf

import "net"

type BlackListItem WhiteListItem

type BlackListConfigure struct {
	BlackList []BlackListItem `json:"BlackList"` //黑名单列表
}

func (w *BlackListItem) Contains(ip string) bool {
	netIP := net.ParseIP(ip)
	if netIP == nil {
		return false
	}
	if w.NetIP != nil {
		return w.NetIP.Equal(netIP)
	}

	if w.Cidr != nil {
		return w.Cidr.Contains(netIP)
	}
	return false
}
