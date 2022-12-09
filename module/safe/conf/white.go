package safeconf

import "net"

type WhiteListBaseConfigure struct {
	URL                string `json:"URL"`
	ActivelifeDuration int32  `json:"ActivelifeDuration"` //有效期限,小时
	BasicAccount       string `json:"BasicAccount"`
	BasicPassword      string `json:"BasicPassword"`
}

type WhiteListConfigure struct {
	BaseConfigure WhiteListBaseConfigure `json:"BaseConfigure"`
	WhiteList     []WhiteListItem        `json:"WhiteList"` //白名单列表
}

type WhiteListItem struct {
	IP            string     `json:"IP"`
	EffectiveTime string     `json:"Effectivetime"` //有效时间
	NetIP         net.IP     `json:"-"`
	Cidr          *net.IPNet `json:"-"`
}

func (w *WhiteListItem) Contains(ip string) bool {
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
