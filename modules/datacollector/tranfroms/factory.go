package tranfroms

import (
	"fmt"
	"lego_datacollector/modules/datacollector/core"
)

func newFactory() (factory *Factory) {
	factory = &Factory{
		transFactory: make(map[string]TFactoryFunc),
	}
	return
}

type Factory struct {
	transFactory map[string]TFactoryFunc
}

func (this *Factory) NewTransforms(index int, runner core.IRunner, conf map[string]interface{}) (rder core.ITransforms, err error) {
	var (
		options core.ITransformsOptions
	)
	if options, err = core.NewTransformsOptions(conf); err != nil {
		return
	}
	if f, ok := this.transFactory[options.GetType()]; ok {
		return f(index, runner, conf)
	} else {
		err = fmt.Errorf("Transforms No RegisterTransforms : %s", options.GetType())
	}
	return
}
func (this *Factory) RegisterTransforms(rtype string, f TFactoryFunc) (err error) {
	if _, ok := this.transFactory[rtype]; ok {
		err = fmt.Errorf("already RegisterTransforms : %s", rtype)
		return
	} else {
		this.transFactory[rtype] = f
	}
	return
}
