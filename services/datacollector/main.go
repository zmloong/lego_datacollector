package main

import (
	"flag"
	"lego_datacollector/comm"
	"lego_datacollector/modules/datacollector"
	"lego_datacollector/services"

	"github.com/liwei1dao/lego"
	"github.com/liwei1dao/lego/base/cluster"
	"github.com/liwei1dao/lego/core"
	"github.com/liwei1dao/lego/sys/rpc"
)

var (
	conf = flag.String("conf", "./conf/datacollector.yaml", "获取需要启动的服务配置文件")
)

func main() {
	flag.Parse()
	s := NewService(
		cluster.SetConfPath(*conf),
		cluster.SetVersion("3.0.1.7"),
	)
	s.OnInstallComp( //装备组件
		services.NewServiceNacosComp(),            //集群信息输出组件
		services.NewServiceDataSyncInfoPushComp(), //数据同步同居信息同步组件
	)
	lego.Run(s, //运行模块
		datacollector.NewModule(),
	)
}

func NewService(ops ...cluster.Option) core.IService {
	s := new(Service)
	s.Configure(ops...)
	return s
}

type Service struct {
	services.ServiceBase
}

func (this *Service) InitSys() {
	this.ServiceBase.InitSys()
	rpc.OnRegisterJsonRpcData(&comm.RunnerConfig{}) //注册rpc通信数据
}
