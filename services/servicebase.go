package services

import (
	"fmt"
	"lego_datacollector/sys/db"
	"lego_datacollector/sys/nacos"

	"github.com/liwei1dao/lego/base/cluster"
	"github.com/liwei1dao/lego/sys/log"
)

type ServiceBase struct {
	cluster.ClusterService
}

func (this *ServiceBase) InitSys() {
	this.ClusterService.InitSys()
	if err := db.OnInit(this.ClusterService.GetSettings().Sys["db"]); err != nil {
		log.Panicf(fmt.Sprintf("statr db err:%v", err))
	} else {
		log.Infof("Sys db Init success !")
	}
	if err := nacos.OnInit(this.ClusterService.GetSettings().Sys["nacos"]); err != nil {
		log.Panicf(fmt.Sprintf("statr nacos err:%v", err))
	} else {
		log.Infof("Sys nacos Init success !")
	}
}
