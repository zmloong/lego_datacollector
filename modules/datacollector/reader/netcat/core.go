package netcat

import (
	"lego_datacollector/modules/datacollector/core"
	"lego_datacollector/modules/datacollector/reader"
	"net"
)

type (
	NetflowMsg struct {
		raddr *net.UDPAddr
		body  []byte
	}
	INetCat interface {
		Start() (err error)
		Close() (err error)
	}
)

func init() {
	reader.RegisterReader(ReaderType, NewReader)
}

const (
	ReaderType = "netcat"
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
	if err = r.Init(runner, r, nil, opt); err != nil {
		return
	}
	rder = r
	return
}
