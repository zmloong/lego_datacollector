package parser

import (
	"fmt"
	"lego_datacollector/modules/datacollector/core"
)

func newFactory() (factory *Factory) {
	factory = &Factory{
		parserFactory: make(map[string]PFactoryFunc),
	}
	return
}

type Factory struct {
	parserFactory map[string]PFactoryFunc
}

func (this *Factory) NewParser(runner core.IRunner, conf map[string]interface{}) (pder core.IParser, err error) {
	var (
		options core.IParserOptions
	)
	if options, err = core.NewParserOptions(conf); err != nil {
		return
	}
	if f, ok := this.parserFactory[options.GetType()]; ok {
		return f(runner, conf)
	} else {
		err = fmt.Errorf("Parser No RegisterReader : %s", options.GetType())
	}
	return
}
func (this *Factory) RegisterParser(rtype string, f PFactoryFunc) (err error) {
	if _, ok := this.parserFactory[rtype]; ok {
		err = fmt.Errorf("already RegisterReader : %s", rtype)
		return
	} else {
		this.parserFactory[rtype] = f
	}
	return
}
