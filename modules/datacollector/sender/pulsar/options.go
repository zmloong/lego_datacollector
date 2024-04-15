package pulsar

import (
	"lego_datacollector/modules/datacollector/core"

	"github.com/liwei1dao/lego/utils/mapstructure"
)

type (
	IOptions interface {
		core.IReaderOptions //继承基础配置
		GetPulsar_Url() string
		GetPulsar_Topics() []string
		GetPulsar_GroupId() string
	}
	Options struct {
		core.ReaderOptions //继承基础配置
		Pulsar_Url         string
		Pulsar_Topics      []string
		Pulsar_GroupId     string
	}
)

func (this *Options) GetPulsar_Url() string {
	return this.Pulsar_Url
}
func (this *Options) GetPulsar_Topics() []string {
	return this.Pulsar_Topics
}
func (this *Options) GetPulsar_GroupId() string {
	return this.Pulsar_GroupId
}
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
