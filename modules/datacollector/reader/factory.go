package reader

import (
	"fmt"
	"lego_datacollector/modules/datacollector/core"
)

func newFactory() (factory *Factory) {
	factory = &Factory{
		readerFactory: make(map[string]RFactoryFunc),
	}
	return
}

type Factory struct {
	readerFactory map[string]RFactoryFunc
}

func (this *Factory) NewReader(runner core.IRunner, conf map[string]interface{}) (rder core.IReader, err error) {
	var (
		options core.IReaderOptions
	)
	if options, err = core.NewReaderOptions(conf); err != nil {
		return
	}
	if f, ok := this.readerFactory[options.GetType()]; ok {
		return f(runner, conf)
	} else {
		err = fmt.Errorf("Reader No RegisterReader : %s", options.GetType())
	}
	return
}
func (this *Factory) RegisterReader(rtype string, f RFactoryFunc) (err error) {
	if _, ok := this.readerFactory[rtype]; ok {
		err = fmt.Errorf("already RegisterReader : %s", rtype)
		return
	} else {
		this.readerFactory[rtype] = f
	}
	return
}
