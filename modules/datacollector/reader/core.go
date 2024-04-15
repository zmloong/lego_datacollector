package reader

import (
	"fmt"
	"lego_datacollector/modules/datacollector/core"

	"github.com/liwei1dao/lego/sys/log"
)

const (
	StatusInit int32 = iota
	StatusStopped
	StatusStopping
	StatusRunning
)

type (
	RFactoryFunc func(runner core.IRunner, conf map[string]interface{}) (rder core.IReader, err error)
	//读取器工厂
	IReaderFactory interface {
		NewReader(runner core.IRunner, conf map[string]interface{}) (rder core.IReader, err error)
		RegisterReader(rtype string, f RFactoryFunc) (err error)
	}

	//读取器 Maps数据读取
	IReaderMaps interface {
		core.IReader
		ReadMaps() (map[string]interface{}, int64, error)
	}
	//读取器 Strings数据读取
	IReaderString interface {
		core.IReader
		ReadStrings() (data string, err error)
	}
)

var (
	factory IReaderFactory
)

func init() {
	factory = newFactory()
	return
}

func NewReader(runner core.IRunner, conf map[string]interface{}) (rder core.IReader, err error) {
	if factory != nil {
		return factory.NewReader(runner, conf)
	} else {
		err = fmt.Errorf("Reader Factory No Init")
	}
	return
}

func RegisterReader(rtype string, f RFactoryFunc) (err error) {
	if factory != nil {
		if err = factory.RegisterReader(rtype, f); err != nil {
			log.Errorf("RegisterReader rtype:%s err:%v", rtype, err)
		}
	} else {
		log.Errorf("Reader Factory No Init")
	}
	return
}
