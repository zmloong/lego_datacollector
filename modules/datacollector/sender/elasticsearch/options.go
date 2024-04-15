package elasticsearch

import (
	"lego_datacollector/modules/datacollector/core"
	"strings"

	"github.com/mitchellh/mapstructure"
)

type (
	IOptions interface {
		core.ISenderOptions //继承基础配置
		GetEs_host() string
		GetEs_index() string
		GetEs_type() string
		GetEs_id() string
	}
	Options struct {
		core.SenderOptions        //继承基础配置
		Es_host            string `json:"es_host"`  //
		Es_index           string `json:"es_index"` //
		Es_type            string `json:"es_type"`  //
		Es_id              string `json:"es_id"`    //
	}
)

func newOptions(config map[string]interface{}) (opt IOptions, err error) {
	options := &Options{}
	if config != nil {
		if err = mapstructure.Decode(config, options); err == nil {
			err = mapstructure.Decode(config, &options.SenderOptions)
		}
	}
	eshost := options.GetEs_host()
	if !strings.HasPrefix(eshost, "http://") && !strings.HasPrefix(eshost, "https://") {
		options.Es_host = "http://" + eshost
	}
	opt = options
	return
}
func (this *Options) GetEs_host() string {
	return this.Es_host
}
func (this *Options) GetEs_index() string {
	return this.Es_index
}
func (this *Options) GetEs_type() string {
	return this.Es_type
}
func (this *Options) GetEs_id() string {
	return this.Es_id
}
