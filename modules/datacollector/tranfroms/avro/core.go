package avro

import (
	"lego_datacollector/modules/datacollector/core"
	"lego_datacollector/modules/datacollector/tranfroms"
)

func init() {
	tranfroms.RegisterTransforms(TranfromsType, NewTranfroms)
}

const (
	TranfromsType = "avro"
)

func NewTranfroms(index int, runner core.IRunner, conf map[string]interface{}) (trans core.ITransforms, err error) {
	var (
		opt IOptions
		t   *Transforms
	)
	if opt, err = newOptions(conf); err != nil {
		return
	}
	t = &Transforms{}
	if err = t.Init(index, runner, t, opt); err != nil {
		return
	}
	trans = t
	return
}
