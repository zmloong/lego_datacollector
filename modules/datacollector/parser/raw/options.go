package raw

import (
	"lego_datacollector/modules/datacollector/core"
	"strings"

	"github.com/liwei1dao/lego/utils/mapstructure"
)

type (
	IOptions interface {
		core.IParserOptions //继承基础配置
		GetKeyTimestamp() bool
		GetIdss_collect_ip() bool
		GetEqpt_ip() bool
		GetInternalKeyPrefix() string
	}
	Options struct {
		core.ParserOptions         //继承基础配置
		Idss_collect_time   bool   `json:"idss_collect_time"`   //是否开始时间值
		Idss_collect_ip     bool   `json:"idss_collect_ip"`     //是否开始时间值
		Eqpt_ip             bool   `json:"eqpt_ip"`             //是否开始时间值
		Internal_key_prefix string `json:"internal_key_prefix"` //采集数据key 前缀
	}
)

func newOptions(config map[string]interface{}) (opt IOptions, err error) {
	options := &Options{}
	if config != nil {
		if err = mapstructure.Decode(config, options); err == nil {
			err = mapstructure.Decode(config, &options.ParserOptions)
		}
	}
	options.Internal_key_prefix = strings.TrimSpace(options.Internal_key_prefix) //去除控股
	opt = options
	return
}
func (this *Options) GetKeyTimestamp() bool {
	return this.Idss_collect_time
}
func (this *Options) GetIdss_collect_ip() bool {
	return this.Idss_collect_ip
}
func (this *Options) GetEqpt_ip() bool {
	return this.Eqpt_ip
}
func (this *Options) GetInternalKeyPrefix() string {
	return this.Internal_key_prefix
}
