package avro

import (
	"lego_datacollector/modules/datacollector/core"
	"lego_datacollector/modules/datacollector/tranfroms"

	"github.com/linkedin/goavro/v2"
)

type Transforms struct {
	tranfroms.Transforms
	options IOptions
	codec   *goavro.Codec
}

func (this *Transforms) Init(index int, runner core.IRunner, transforms core.ITransforms, options core.ITransformsOptions) (err error) {
	err = this.Transforms.Init(index, runner, transforms, options)
	this.options = options.(IOptions)
	this.codec, err = goavro.NewCodec(this.options.GetSchema())
	return
}

func (this *Transforms) Trans(bucket core.ICollDataBucket) {
	for _, v := range bucket.Items() {
		if data, ok := v.GetValue().([]byte); ok {
			if native, _, err := this.codec.NativeFromBinary(data); err != nil {
				this.Runner.Errorf("Transforms avro err:%v", err)
			} else {
				v.GetData()[core.KeyRaw] = native
			}
		}
	}
	this.Transforms.Trans(bucket)
}
