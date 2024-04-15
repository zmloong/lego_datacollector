package avro

import (
	"errors"
	"lego_datacollector/modules/datacollector/core"

	"github.com/liwei1dao/lego/utils/mapstructure"
)

type (
	IOptions interface {
		core.ITransformsOptions //继承基础配置
		GetSchema() string
	}
	Options struct {
		core.TransformsOptions        //继承基础配置
		Schema                 string `json:"schema"`
	}
)

func (this *Options) GetSchema() string {
	return this.Schema
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
	if len(options.Schema) == 0 {
		err = errors.New("Transforms avro Schema is null")
		return
	}
	opt = options
	return
}
