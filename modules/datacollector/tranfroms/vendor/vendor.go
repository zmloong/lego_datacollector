package vendor

import (
	"lego_datacollector/modules/datacollector/core"
	"lego_datacollector/modules/datacollector/tranfroms"
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
		if _, ok := v.GetData()[core.KeyEqptip]; ok {
			for _, k := range this.options.GetVendorInfo() {
				if v.GetData()[core.KeyEqptip] == k.Eqpt_ip {
					v.GetData()[core.KeyVendor] = k.Vendor
					v.GetData()[core.KeyDevtype] = k.Dev_type
					v.GetData()[core.KeyProduct] = k.Product
					v.GetData()[core.KeyEqptname] = k.Eqpt_name
					v.GetData()[core.KeyEqptDeviceType] = k.Eqpt_device_type
				}
			}
		}
	}
	this.Transforms.Trans(bucket)
}
