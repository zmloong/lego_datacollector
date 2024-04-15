package netflow

import (
	"lego_datacollector/modules/datacollector/core"

	"github.com/liwei1dao/lego/utils/mapstructure"
)

type (
	IOptions interface {
		core.IReaderOptions //继承基础配置
		GetNetflow_listen_port() int
		GetNetflow_read_size() int
	}
	Options struct {
		core.ReaderOptions      //继承基础配置
		Netflow_listen_port int //Netflow 监听端口
		Netflow_read_size   int //Netflow 读取缓存大小
	}
)

func newOptions(config map[string]interface{}) (opt IOptions, err error) {
	options := &Options{
		Netflow_read_size: 2048,
	}
	if config != nil {
		if err = mapstructure.Decode(config, options); err == nil {
			err = mapstructure.Decode(config, &options.ReaderOptions)
		}
	}
	opt = options
	return
}

func (this *Options) GetNetflow_listen_port() int {
	return this.Netflow_listen_port
}

func (this *Options) GetNetflow_read_size() int {
	return this.Netflow_read_size
}
