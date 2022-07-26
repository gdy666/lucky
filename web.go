//go:build adminweb
// +build adminweb

package main

import (
	"fmt"
	"github.com/gdy666/lucky/web"
	"log"
)

func RunAdminWeb(listenPort int) {
	listen := fmt.Sprintf(":%d", listenPort)
	go web.RunAdminWeb(listen)
	log.Printf("AdminWeb listen on %s", listen)
}
