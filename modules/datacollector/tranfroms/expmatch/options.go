package expmatch

import (
	"lego_datacollector/modules/datacollector/core"

	"github.com/liwei1dao/lego/utils/mapstructure"
)

type (
	IOptions interface {
		core.ITransformsOptions //继承基础配置
		GetExpMatch() string    //正则字段
		GetIsNegate() bool      //是否取反
		GetIsDiscard() bool     //是否丢弃
	}
	Options struct {
		core.TransformsOptions        //继承基础配置
		ExpMatch               string `json:"expmatch"`  //正则字段
		IsNegate               bool   `json:"isnegate"`  //是否取反
		IsDiscard              bool   `json:"isdiscard"` //是否丢弃
	}
)

func (this *Options) GetExpMatch() string {
	return this.ExpMatch
}

func (this *Options) GetIsNegate() bool {
	return this.IsNegate
}

func (this *Options) GetIsDiscard() bool {
	return this.IsDiscard
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
