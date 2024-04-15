package binary

import (
	"lego_datacollector/modules/datacollector/core"
	"lego_datacollector/modules/datacollector/metaer/folder"
	"lego_datacollector/modules/datacollector/reader"
)

func init() {
	reader.RegisterReader(ReaderType, NewReader)
}

const (
	ReaderType = "binary"
)

type (
	BinaryData struct {
		Name     string //相对路径
		Size     uint64 //文件大小
		CheckNum int    //切片数量
		Check    int    //切片
		Start    uint64
		End      uint64
		Base64   string
	}
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
	if err = r.Init(runner, r, folder.NewMeta("reader"), opt); err != nil {
		return
	}
	rder = r
	return
}
