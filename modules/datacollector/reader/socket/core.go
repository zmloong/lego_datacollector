package socket

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"lego_datacollector/modules/datacollector/core"
	"lego_datacollector/modules/datacollector/metaer/folder"
	"lego_datacollector/modules/datacollector/reader"
	"strings"
)

func init() {
	reader.RegisterReader(ReaderType, NewReader)
}

type (
	SocketMessage struct {
		Address string
		Message string
	}
	IReader interface {
		core.IReader
		GetOptions() IOptions
		GetNetproto() string
	}
	ISocketReader interface {
		listen()
		close() error
	}
)

const (
	ReaderType = "socket"
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

func TruncateStrSize(err string, size int) string {
	if len(err) <= size {
		return err
	}

	return fmt.Sprintf(err[:size]+"......(only show %d bytes, remain %d bytes)",
		size, len(err)-size)
}

func tryDecodeReader(decoder *json.Decoder) bool {
	bufferReader := decoder.Buffered()
	readBytes, err := ioutil.ReadAll(bufferReader)
	if err == io.EOF {
		return false
	}
	if err != nil {
		return true
	}
	if len(strings.TrimSpace(string(readBytes))) <= 0 {
		return false
	}
	return true
}
