package reader

import (
	"lego_datacollector/modules/datacollector/core"
)

type Reader struct {
	Runner  core.IRunner
	reader  core.IReader
	meta    core.IMetaerData
	options core.IReaderOptions
}

func (this *Reader) GetRunner() core.IRunner {
	return this.Runner
}
func (this *Reader) GetType() string {
	return this.options.GetType()
}

func (this *Reader) GetEncoding() core.Encoding {
	return this.options.GetEncoding()
}

func (this *Reader) Init(runner core.IRunner, reader core.IReader, meta core.IMetaerData, options core.IReaderOptions) (err error) {
	defer runner.Infof("NewReader options:%+v", options)
	this.Runner = runner
	this.meta = meta
	this.options = options
	if meta != nil { //有些采集器不需要保存记录
		err = this.Runner.Metaer().Read(meta)
	}
	return
}

func (this *Reader) Start() (err error) {
	return
}

///外部调度器 驱动执行  此接口 不可阻塞
func (this *Reader) Drive() (err error) {

	return
}

//关闭 只允许 Runner 对象调用
func (this *Reader) Close() (err error) {
	// err = this.SyncMeta()													此处无需加同步数据代码 Runner 在关闭最后 会将通过 Metaer 同步所有的 MetaeData
	this.Runner.Debugf("Reader Start Close!")
	return
}

func (this *Reader) SyncMeta() (err error) {
	return this.Runner.Metaer().Write(this.meta)
}

func (this *Reader) Input() chan<- core.ICollData {
	return this.Runner.ReaderPipe()
}
