package sender

import (
	"fmt"
	"lego_datacollector/modules/datacollector/core"

	"github.com/liwei1dao/lego/sys/log"
)

type (
	SFactoryFunc func(runner core.IRunner, conf map[string]interface{}) (rder core.ISender, err error)
	//解析器工厂
	ISenderFactory interface {
		RegisterSender(stype string, f SFactoryFunc) (err error)
		NewSender(runner core.IRunner, conf map[string]interface{}) (sders core.ISender, err error)
	}
)

var (
	factory ISenderFactory
)

func init() {
	factory = newFactory()
	return
}

func NewSender(runner core.IRunner, conf []map[string]interface{}) (sders []core.ISender, err error) {
	sders = make([]core.ISender, len(conf))
	if factory != nil {
		for i, v := range conf {
			if sders[i], err = factory.NewSender(runner, v); err != nil {
				sders = nil
				break
			}
		}
	} else {
		err = fmt.Errorf("Parser Factory No Init")
	}
	return
}

func RegisterSender(rtype string, f SFactoryFunc) (err error) {
	if factory != nil {
		if err = factory.RegisterSender(rtype, f); err != nil {
			log.Errorf("RegisterReader rtype:%s err:%v", rtype, err)
		}
	} else {
		log.Errorf("Reader Factory No Init")
	}
	return
}
