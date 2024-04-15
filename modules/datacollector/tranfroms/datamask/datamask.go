package datamask

import (
	"fmt"
	"lego_datacollector/modules/datacollector/core"
	"lego_datacollector/modules/datacollector/tranfroms"

	"github.com/liwei1dao/lego/utils/crypto/aes"
)

type Transforms struct {
	tranfroms.Transforms
	options IOptions
}

func (this *Transforms) Init(index int, runner core.IRunner, transforms core.ITransforms, options core.ITransformsOptions) (err error) {
	err = this.Transforms.Init(index, runner, transforms, options)
	this.options = options.(IOptions)
	return
}

func (this *Transforms) Trans(bucket core.ICollDataBucket) {
	for _, v := range bucket.Items() {
		data, ok := v.GetValue().(map[string]interface{})
		if ok {
			for _, k := range this.options.GetKeys() {
				if d, ok := data[k]; ok && d != nil {
					data[k] = aes.AesEncryptCBC(fmt.Sprintf("%v", d), this.options.GetSignKey())
				} else {
					this.Runner.Debugf("Transforms no fund key:%s", k)
				}
			}
		} else {
			this.Runner.Debugf("Transforms data is no map %v", v.GetValue())
		}
	}
	this.Transforms.Trans(bucket)
}
