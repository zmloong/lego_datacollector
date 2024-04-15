package mongodb

import (
	"lego_datacollector/modules/datacollector/core"
	"lego_datacollector/modules/datacollector/metaer/sql"
	"lego_datacollector/modules/datacollector/reader"
)

func init() {
	reader.RegisterReader(ReaderType, NewReader)
}

const (
	ReaderType = "mongodb"
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
	if err = r.Init(runner, r, sql.NewMeta("reader"), opt); err != nil {
		return
	}
	rder = r
	return
}
