package sender

import (
	"fmt"
	"lego_datacollector/modules/datacollector/core"
)

func newFactory() (factory *Factory) {
	factory = &Factory{
		senderFactory: make(map[string]SFactoryFunc),
	}
	return
}

type Factory struct {
	senderFactory map[string]SFactoryFunc
}

func (this *Factory) NewSender(runner core.IRunner, conf map[string]interface{}) (rder core.ISender, err error) {
	var (
		options core.ISenderOptions
	)
	if options, err = core.NewSenderOptions(conf); err != nil {
		return
	}
	if f, ok := this.senderFactory[options.GetType()]; ok {
		return f(runner, conf)
	} else {
		err = fmt.Errorf("Sender No RegisterReader : %s", options.GetType())
	}
	return
}
func (this *Factory) RegisterSender(rtype string, f SFactoryFunc) (err error) {
	if _, ok := this.senderFactory[rtype]; ok {
		err = fmt.Errorf("already RegisterReader : %s", rtype)
		return
	} else {
		this.senderFactory[rtype] = f
	}
	return
}
