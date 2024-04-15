package nacos

import (
	"encoding/json"
	"lego_datacollector/comm"

	"github.com/liwei1dao/lego/sys/log"
	"github.com/liwei1dao/lego/sys/sdks/aliyun/nacos"
	"github.com/nacos-group/nacos-sdk-go/vo"
)

func newSys(options Options) (sys *Nacos, err error) {
	sys = &Nacos{
		options: options,
	}
	sys.nacos, err = nacos.NewSys(
		nacos.SetNacosClientType(options.NacosClientType),
		nacos.SetNamespaceId(options.NamespaceId),
		nacos.SetNacosAddr(options.NacosAddr),
		nacos.SetPort(options.Port),
		nacos.SetTimeoutMs(options.TimeoutMs),
	)
	return
}

type Nacos struct {
	options Options
	nacos   nacos.ISys
}

func (this *Nacos) GetServiceList() (services []*comm.ServiceNode, err error) {
	var str string
	services = make([]*comm.ServiceNode, 0)
	if str, err = this.nacos.Config_GetConfig(vo.ConfigParam{
		DataId: this.options.DataId,
		Group:  this.options.Group,
	}); err == nil && len(str) > 0 {
		if err = json.Unmarshal([]byte(str), &services); err != nil {
			log.Errorf("GetServiceList str:%s err:%v", str, err)
		}
	}
	return
}

//更新集群服务列表
func (this *Nacos) UpdataServiceList(services []*comm.ServiceNode) (err error) {
	var content []byte
	if content, err = json.Marshal(services); err != nil {
		return
	}
	if _, err = this.nacos.Config_PublishConfig(vo.ConfigParam{
		DataId:  this.options.DataId,
		Group:   this.options.Group,
		Content: string(content),
	}); err != nil {
		log.Errorf("UpdataServiceList  err:%v", err)
	}
	return
}
