package stringsp

import (
	"net/url"
	"strings"
)

func GetHostAndPathFromURL(urlstr string) (string, string, string, string, error) {
	if !strings.HasPrefix(urlstr, "http") {
		urlstr = "http://" + urlstr
	}
	u, err := url.Parse(urlstr)
	if err != nil {
		return "", "", "", "", err
	}
	return u.Scheme, u.Hostname(), u.Port(), u.Path, nil
}
