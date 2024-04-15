package doris

import (
	"lego_datacollector/modules/datacollector/core"
	"lego_datacollector/modules/datacollector/sender"
	"path/filepath"
	"strconv"
)

func init() {
	sender.RegisterSender(SenderType, NewSender)
}

const (
	SenderType = "doris"
)

func NewSender(runner core.IRunner, conf map[string]interface{}) (rder core.ISender, err error) {
	var (
		opt IOptions
		s   *Sender
	)
	if opt, err = newOptions(conf); err != nil {
		return
	}
	s = &Sender{}
	if err = s.Init(runner, s, opt); err != nil {
		return
	}
	rder = s
	return
}

func getPartitionFolder(nowStr string, idx int) string {
	base := filepath.Base(nowStr)
	dir := filepath.Dir(nowStr)
	return filepath.Join(dir, "partition"+strconv.Itoa(idx), base)
}
