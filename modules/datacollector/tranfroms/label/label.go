package label

import (
	"lego_datacollector/modules/datacollector/core"
	"lego_datacollector/modules/datacollector/tranfroms"
)

type Transforms struct {
	tranfroms.Transforms
	options IOptions
	keys    []string
}

func (this *Transforms) Init(index int, runner core.IRunner, transforms core.ITransforms, options core.ITransformsOptions) (err error) {
	err = this.Transforms.Init(index, runner, transforms, options)
	this.options = options.(IOptions)
	this.keys = tranfroms.GetKeys(this.options.GetKey())
	return
}

func (this *Transforms) Trans(bucket core.ICollDataBucket) {
	for _, v := range bucket.Items() {
		for _, k := range this.keys {
			v.GetData()[k] = this.options.GetValue()
		}
	}
	this.Transforms.Trans(bucket)
}
