package raw

import (
	"lego_datacollector/modules/datacollector/core"
	"lego_datacollector/modules/datacollector/parser"
)

func init() {
	parser.RegisterParser(ParserType, NewParser)
}

const (
	ParserType = "raw"
)

func NewParser(runner core.IRunner, conf map[string]interface{}) (rder core.IParser, err error) {
	var (
		opt IOptions
		p   *Parser
	)
	if opt, err = newOptions(conf); err != nil {
		return
	}
	p = &Parser{}
	if err = p.Init(runner, p, opt); err != nil {
		return
	}
	rder = p
	return
}
