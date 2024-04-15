package parser

import (
	"fmt"
	"lego_datacollector/modules/datacollector/core"
	"strings"
	"sync"
)

type Parser struct {
	Runner  core.IRunner
	parser  core.IParser
	options core.IParserOptions
	Procs   int
	wg      *sync.WaitGroup
	labels  []GrokLabel
}

func (this *Parser) GetRunner() core.IRunner {
	return this.Runner
}
func (this *Parser) GetType() string {
	return this.options.GetType()
}

func (this *Parser) GetMetaerData() (meta core.IMetaerData) {
	return
}

func (this *Parser) Init(rer core.IRunner, per core.IParser, options core.IParserOptions) (err error) {
	defer rer.Infof("NewParser options:%+v", options)
	this.Runner = rer
	this.parser = per
	this.options = options
	this.Procs = this.Runner.MaxProcs()
	this.wg = new(sync.WaitGroup)
	this.labels, err = this.getGrokLabels(this.options.GetLabels())

	return
}
func (this *Parser) Start() (err error) {
	if this.Procs < 1 {
		this.Procs = 1
	}
	this.wg.Add(this.Procs)
	for i := 0; i < this.Procs; i++ {
		go this.run()
	}
	return
}

func (this *Parser) run() {
	defer this.wg.Done()
	for v := range this.Runner.ParserPipe() {
		this.parser.Parse(v)
	}
}

func (this *Parser) Parse(bucket core.ICollDataBucket) {
	this.Runner.Push_TransformsPipe(bucket)
}

//关闭
func (this *Parser) Close() (err error) {
	this.wg.Wait()
	this.Runner.Debugf("Parser Close Succ")
	return
}

func (this *Parser) Addlabels(data core.ICollData) {
	for _, label := range this.labels {
		data.GetData()[label.Name] = label.Value
	}
}

func (this *Parser) getGrokLabels(labelList []string) (labels []GrokLabel, err error) {
	nameMap := make(map[string]struct{})
	labels = make([]GrokLabel, 0)
	for _, f := range labelList {
		parts := strings.Fields(f)
		if len(parts) < 2 {
			err = fmt.Errorf("Parser label conf :%s  error:format should be \"labelName labelValue\", ignore this label...", f)
			return
		}
		labelName, labelValue := parts[0], parts[1]
		if _, ok := nameMap[labelName]; ok {
			err = fmt.Errorf("Parser label name %v was duplicated, ignore this lable <%v,%v>...", labelName, labelName, labelValue)
			return
		}
		nameMap[labelName] = struct{}{}
		l := GrokLabel{Name: labelName, Value: labelValue}
		labels = append(labels, l)
	}
	return
}
