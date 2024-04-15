package services

import (
	"lego_datacollector/comm"
	"lego_datacollector/sys/db"
	"lego_datacollector/sys/nacos"

	"github.com/liwei1dao/lego/base"
	"github.com/liwei1dao/lego/core"
	"github.com/liwei1dao/lego/core/cbase"
	"github.com/liwei1dao/lego/sys/event"
	"github.com/liwei1dao/lego/sys/log"
	"github.com/liwei1dao/lego/sys/registry"
)

func NewServiceNacosComp() core.IServiceComp {
	comp := new(ServiceNacosComp)
	return comp
}

/*
维护采集任务集群配置组件
*/
type ServiceNacosComp struct {
	cbase.ServiceCompBase
	service base.IClusterService
}

func (this *ServiceNacosComp) GetName() core.S_Comps {
	return "NacosComp"
}

func (this *ServiceNacosComp) Init(service core.IService, comp core.IServiceComp, options core.ICompOptions) (err error) {
	err = this.ServiceCompBase.Init(service, comp, options)
	this.service = service.(base.IClusterService)
	return
}

func (this *ServiceNacosComp) Start() (err error) {
	err = this.ServiceCompBase.Start()
	event.RegisterGO(core.Event_RegistryStart, this.Event_RegistryStart)
	event.RegisterGO(core.Event_FindNewService, this.Event_FindNewService)
	event.RegisterGO(comm.Event_UpdatRunner, this.Event_UpdatRunner)
	event.RegisterGO(core.Event_LoseService, this.Event_LoseService)
	return
}

//Event---------------------------------------------------------------------------------------------------------

//有新的服务加入到集群中
func (this *ServiceNacosComp) Event_RegistryStart() {
	var (
		module        core.IModule
		dateCollector comm.IDateCollector
		original      map[string]*comm.ServiceNode
		ss            []*comm.ServiceNode
		err           error
		ok            bool
		n             int
	)
	if module, err = this.service.GetModule(comm.SM_DataCollectorModule); err != nil {
		log.Errorf("Event_RegistryStart found:%v err:%v", comm.SM_DataCollectorModule, err)
		return
	}
	if dateCollector, ok = module.(comm.IDateCollector); !ok {
		log.Errorf("Event_RegistryStart %v not is comm.IDateCollector", comm.SM_DataCollectorModule)
	}
	log.Debugf("Event_RegistryStart ip:%v", this.service.GetIp())
	ok = false
	if original, err = db.UpDateServiceNodeInfo(&comm.ServiceNode{
		Id:        this.service.GetId(),
		IP:        this.service.GetIp(),
		Wight:     float32(this.service.GetPreWeight()),
		Port:      dateCollector.GetListenPort(),
		Runner:    make([]string, 0),
		RunnerNum: 0,
		State:     1,
	}); err == nil {
		n = 0
		ss = make([]*comm.ServiceNode, len(original))
		for _, v := range original {
			ss[n] = v
			n++
		}
		if err = nacos.UpdataServiceList(ss); err != nil {
			log.Errorf("Event_FindNewService err:%v", err)
			return
		}
	} else { //其他错误输出
		log.Errorf("Event_RegistryStart err:%v", err)
		return
	}
}

//有新的服务加入到集群中
func (this *ServiceNacosComp) Event_FindNewService(node registry.ServiceNode) {
	if node.Id != this.service.GetId() { //不是发现自己不处理
		return
	}

	var (
		module        core.IModule
		dateCollector comm.IDateCollector
		runner        []string
		original      map[string]*comm.ServiceNode
		ss            []*comm.ServiceNode
		err           error
		ok            bool
		n             int
	)
	if module, err = this.service.GetModule(comm.SM_DataCollectorModule); err != nil {
		log.Errorf("Event_FindNewService found:%v err:%v", comm.SM_DataCollectorModule, err)
		return
	}
	if dateCollector, ok = module.(comm.IDateCollector); !ok {
		log.Errorf("Event_FindNewService %v not is comm.IDateCollector", comm.SM_DataCollectorModule)
	}
	if ss, err = nacos.GetServiceList(); err != nil {
		log.Errorf("Event_FindNewService nacos.GetServiceList err:%v", err)
		return
	}
	ok = false
	runner = dateCollector.GetRunner()
	log.Debugf("Event_FindNewService runner:%v", runner)

	if original, err = db.UpDateServiceNodeInfo(&comm.ServiceNode{
		Id:        this.service.GetId(),
		IP:        this.service.GetIp(),
		Wight:     float32(this.service.GetPreWeight()),
		Port:      dateCollector.GetListenPort(),
		Runner:    runner,
		RunnerNum: len(runner),
		State:     1,
	}); err == nil {
		n = 0
		ss = make([]*comm.ServiceNode, len(original))
		for _, v := range original {
			ss[n] = v
			n++
		}
		if err = nacos.UpdataServiceList(ss); err != nil {
			log.Errorf("Event_FindNewService err:%v", err)
			return
		}
		return
	} else { //其他错误输出
		log.Errorf("StatisticRunner err:%v", err)
		return
	}

}

func (this *ServiceNacosComp) Event_UpdatRunner() {
	var (
		module        core.IModule
		dateCollector comm.IDateCollector
		original      map[string]*comm.ServiceNode
		ss            []*comm.ServiceNode
		err           error
		ok            bool
		runner        []string
		n             int
	)
	if module, err = this.service.GetModule(comm.SM_DataCollectorModule); err != nil {
		log.Errorf("Event_UpdatRunner found:%v err:%v", comm.SM_DataCollectorModule, err)
		return
	}
	if dateCollector, ok = module.(comm.IDateCollector); !ok {
		log.Errorf("Event_UpdatRunner %v not is comm.IDateCollector", comm.SM_DataCollectorModule)
	}
	if ss, err = nacos.GetServiceList(); err != nil {
		log.Errorf("Event_UpdatRunner nacos.GetServiceList err:%v", err)
		return
	}
	ok = false
	runner = dateCollector.GetRunner()
	log.Debugf("Event_UpdatRunner runner:%v", runner)
	if original, err = db.UpDateServiceNodeInfo(&comm.ServiceNode{
		Id:        this.service.GetId(),
		IP:        this.service.GetIp(),
		Wight:     float32(this.service.GetPreWeight()),
		Port:      dateCollector.GetListenPort(),
		Runner:    runner,
		RunnerNum: len(runner),
		State:     1,
	}); err == nil {
		n = 0
		ss = make([]*comm.ServiceNode, len(original))
		for _, v := range original {
			ss[n] = v
			n++
		}
		if err = nacos.UpdataServiceList(ss); err != nil {
			log.Errorf("Event_UpdatRunner err:%v", err)
			return
		}
		return
	} else { //其他错误输出
		log.Errorf("Event_UpdatRunner err:%v", err)
		return
	}

}

//有服务丢失
func (this *ServiceNacosComp) Event_LoseService(sID string) {
	var (
		original map[string]*comm.ServiceNode
		ss       []*comm.ServiceNode
		err      error
		n        int
	)
	if original, err = db.QueryServiceNodeInfo(); err != nil {
		log.Errorf("Event_LoseService sID:%s err:%v", sID, err)
		return
	}
	if v, ok := original[sID]; ok && v.State == 1 {
		v.State = 0
		if original, err = db.UpDateServiceNodeInfo(v); err == nil {
			n = 0
			ss = make([]*comm.ServiceNode, len(original))
			for _, v := range original {
				ss[n] = v
				n++
			}
			if err = nacos.UpdataServiceList(ss); err != nil {
				log.Errorf("Event_LoseService err:%v", err)
				return
			}
			return
		} else { //其他错误输出
			log.Errorf("Event_LoseService err:%v", err)
			return
		}
	}
}
