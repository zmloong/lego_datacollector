package netcat

import (
	"lego_datacollector/modules/datacollector/core"

	"github.com/liwei1dao/lego/utils/mapstructure"
)

type (
	ProtoType uint8
	IOptions  interface {
		core.IReaderOptions //继承基础配置
		GetNetcat_listen_port() int
	}
	Options struct {
		core.ReaderOptions //继承基础配置
		Netcat_listen_port int
	}
)

func newOptions(config map[string]interface{}) (opt IOptions, err error) {
	options := &Options{}
	if config != nil {
		if err = mapstructure.Decode(config, options); err == nil {
			err = mapstructure.Decode(config, &options.ReaderOptions)
		}
	}
	opt = options
	return
}

func (this *Options) GetNetcat_listen_port() int {
	return this.Netcat_listen_port
}
