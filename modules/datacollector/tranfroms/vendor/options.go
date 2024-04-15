package vendor

import (
	"fmt"
	"lego_datacollector/modules/datacollector/core"

	"github.com/liwei1dao/lego/utils/mapstructure"
)

type (
	IOptions interface {
		core.ITransformsOptions //继承基础配置
		GetVendorInfo() []VendorInfo
	}
	Options struct {
		core.TransformsOptions //继承基础配置
		VendorInfo             []VendorInfo
	}
	VendorInfo struct {
		Eqpt_ip          string `json:"eqpt_ip"`
		Vendor           string `json:"vendor"`
		Dev_type         string `json:"dev_type"`
		Product          string `json:"product"`
		Eqpt_name        string `json:"eqpt_name"`
		Eqpt_device_type string `json:"eqpt_device_type"`
	}
)

func (this *Options) GetVendorInfo() []VendorInfo {
	return this.VendorInfo
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
	fmt.Println(options)
	return
}
