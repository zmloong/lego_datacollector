package expmatch

import (
	"errors"
	"lego_datacollector/modules/datacollector/core"
	"lego_datacollector/modules/datacollector/tranfroms"
	"regexp"
)

type Transforms struct {
	tranfroms.Transforms
	options IOptions
	exp     *regexp.Regexp
}

func (this *Transforms) Init(index int, runner core.IRunner, transforms core.ITransforms, options core.ITransformsOptions) (err error) {
	err = this.Transforms.Init(index, runner, transforms, options)
	this.options = options.(IOptions)
	this.exp, err = regexp.Compile(this.options.GetExpMatch())
	return
}

func (this *Transforms) Trans(bucket core.ICollDataBucket) {
	for _, v := range bucket.Items() {
		if v1, ok := v.GetData()[core.KeyRaw].(string); ok {
			ismatch := this.exp.MatchString(v1)
			if !this.options.GetIsNegate() {
				if this.options.GetIsDiscard() {
					if ismatch { //丢弃掉
						v.SetError(errors.New("正则过滤掉了"))
					}
				} else {
					v.GetData()[core.KeyIdsCollecexpmatch] = ismatch
				}
			} else {
				if this.options.GetIsDiscard() {
					if !ismatch { //丢弃掉
						v.SetError(errors.New("正则过滤掉了"))
					}
				} else {
					v.GetData()[core.KeyIdsCollecexpmatch] = !ismatch
				}

			}

		}
	}
	this.Transforms.Trans(bucket)
}
