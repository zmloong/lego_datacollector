package parser

import (
	"fmt"
	"lego_datacollector/modules/datacollector/core"

	"github.com/liwei1dao/lego/sys/log"
)

type (
	PFactoryFunc func(runner core.IRunner, conf map[string]interface{}) (pder core.IParser, err error)
	//解析器工厂
	IParserFactory interface {
		NewParser(runner core.IRunner, conf map[string]interface{}) (per core.IParser, err error)
		RegisterParser(rtype string, f PFactoryFunc) (err error)
	}
	GrokLabel struct {
		Name  string
		Value string
	}
)

var (
	factory IParserFactory
)

func init() {
	factory = newFactory()
	return
}

func NewParser(runner core.IRunner, conf map[string]interface{}) (per core.IParser, err error) {
	if factory != nil {
		return factory.NewParser(runner, conf)
	} else {
		err = fmt.Errorf("Parser Factory No Init")
	}
	return
}

func RegisterParser(rtype string, f PFactoryFunc) (err error) {
	if factory != nil {
		if err = factory.RegisterParser(rtype, f); err != nil {
			log.Errorf("RegisterReader rtype:%s err:%v", rtype, err)
		}
	} else {
		log.Errorf("Reader Factory No Init")
	}
	return
}
