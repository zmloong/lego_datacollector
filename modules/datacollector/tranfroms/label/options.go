package label

import (
	"lego_datacollector/modules/datacollector/core"

	"github.com/liwei1dao/lego/utils/mapstructure"
)

type (
	IOptions interface {
		core.ITransformsOptions //继承基础配置
		GetKey() string
		GetValue() string
		GetOverride() bool
	}
	Options struct {
		core.TransformsOptions        //继承基础配置
		Key                    string `json:"key"`
		Value                  string `json:"value"`
		Override               bool   `json:"override"`
	}
)

func (this *Options) GetKey() string {
	return this.Key
}
func (this *Options) GetValue() string {
	return this.Value
}
func (this *Options) GetOverride() bool {
	return this.Override
}

func newOptions(config map[string]interface{}) (opt IOptions, err error) {
	options := &Options{}
	if config != nil {
		if err = mapstructure.Decode(config, options); err == nil {
			err = mapstructure.Decode(config, &options.TransformsOptions)
		}
		if err != nil {
			return
		}
	}

	opt = options
	return
}
