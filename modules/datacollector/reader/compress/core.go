package compress

import (
	"lego_datacollector/modules/datacollector/core"
	compress "lego_datacollector/modules/datacollector/metaer/compress"
	"lego_datacollector/modules/datacollector/reader"
)

func init() {
	reader.RegisterReader(ReaderType, NewReader)
}

const (
	ReaderType = "compress"
)

func NewReader(runner core.IRunner, conf map[string]interface{}) (rder core.IReader, err error) {
	var (
		opt IOptions
		r   *Reader
	)
	if opt, err = newOptions(conf); err != nil {
		return
	}
	r = &Reader{}
	if err = r.Init(runner, r, compress.NewMeta("reader"), opt); err != nil {
		return
	}
	rder = r
	return
}
