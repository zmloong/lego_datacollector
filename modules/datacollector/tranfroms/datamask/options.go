package datamask

import (
	"crypto/aes"
	"errors"
	"lego_datacollector/modules/datacollector/core"

	"github.com/liwei1dao/lego/utils/mapstructure"
)

type (
	IOptions interface {
		core.ITransformsOptions //继承基础配置
		GetSignKey() string
		GetKeys() []string
	}
	Options struct {
		core.TransformsOptions          //继承基础配置
		SignKey                string   `json:"signkey"`
		Keys                   []string `json:"keys"`
	}
)

func (this *Options) GetSignKey() string {
	return this.SignKey
}
func (this *Options) GetKeys() []string {
	return this.Keys
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
	if len(options.Keys) == 0 {
		err = errors.New("Transforms datamask Keys is null")
		return
	}
	if _, err = aes.NewCipher([]byte(options.SignKey)); err != nil {
		return
	}
	opt = options
	return
}
