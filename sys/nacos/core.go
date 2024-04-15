package nacos

import "lego_datacollector/comm"

type (
	ISys interface {
		GetServiceList() (services []*comm.ServiceNode, err error)
		UpdataServiceList(services []*comm.ServiceNode) (err error)
	}
)

var (
	defsys ISys
)

func OnInit(config map[string]interface{}, option ...Option) (err error) {
	defsys, err = newSys(newOptions(config, option...))
	return
}

func NewSys(option ...Option) (sys ISys, err error) {
	sys, err = newSys(newOptionsByOption(option...))
	return
}

func GetServiceList() (services []*comm.ServiceNode, err error) {
	return defsys.GetServiceList()
}

func UpdataServiceList(services []*comm.ServiceNode) (err error) {
	return defsys.UpdataServiceList(services)
}
