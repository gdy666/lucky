package main

import (
	"github.com/gdy666/lucky/config"
	"github.com/gdy666/lucky/web"
)

func RunAdminWeb(conf *config.BaseConfigure) {
	//listen := fmt.Sprintf(":%d", listenPort)
	go web.RunAdminWeb(conf)
}
