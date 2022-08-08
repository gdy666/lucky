package httputils

import (
	"net/http"
	"time"
)

func CreateHttpClient(transportNetwork, localAddr string, secureSkipVerify bool, proxyType, proxyUrl, user, passwd string, timeout time.Duration) (*http.Client, error) {

	transport, err := NewTransport(transportNetwork, localAddr, secureSkipVerify, proxyType, proxyUrl, user, passwd)
	if err != nil {
		return nil, err
	}

	httpClient := &http.Client{
		Timeout:   timeout,
		Transport: transport}

	return httpClient, nil
}
