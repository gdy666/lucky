//go:build adminweb
// +build adminweb

package main

import (
	"fmt"
	"github.com/ljymc/goports/web"
	"log"
)

func RunAdminWeb(listenPort int) {
	listen := fmt.Sprintf(":%d", listenPort)
	go web.RunAdminWeb(listen)
	log.Printf("AdminWeb listen on %s", listen)
}
