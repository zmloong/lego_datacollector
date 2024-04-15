package pulsar

import (
	"lego_datacollector/modules/datacollector/core"
	"lego_datacollector/modules/datacollector/reader"

	"github.com/liwei1dao/lego/sys/pulsar"
)

type Reader struct {
	reader.Reader
	options IOptions //以接口对象传递参数 方便后期继承扩展
	pulsar  pulsar.IPulsar
}

func (this *Reader) Init(runner core.IRunner, reader core.IReader, meta core.IMetaerData, options core.IReaderOptions) (err error) {
	if err = this.Reader.Init(runner, reader, meta, options); err != nil {
		return
	}
	this.options = options.(IOptions)
	this.pulsar, err = pulsar.NewSys(
		pulsar.SetStartType(pulsar.Consumer),
		pulsar.SetPulsarUrl(this.options.GetPulsar_Url()),
		pulsar.SetTopics(this.options.GetPulsar_Topics()),
		pulsar.SetConsumerGroupId(this.options.GetPulsar_GroupId()),
	)
	return
}
func (this *Reader) Start() (err error) {
	err = this.Reader.Start()
	for i := 0; i < this.Runner.MaxProcs(); i++ {
		go this.run()
	}
	return
}

func (this *Reader) Close() (err error) {
	this.pulsar.Close()
	this.Runner.Infof("reader pulsar Close succ")
	err = this.Reader.Close()
	return
}

//外部协成启动
func (this *Reader) run() {
	for v := range this.pulsar.Consumer_Messages() {
		// fmt.Sprintf("Consumer_Messages:%d -------1", v.Offset)
		this.Input() <- core.NewCollData("", string(v.Message.Payload()))
		// log.Debugf("Consumer_Messages:%d -------2", v.Offset)
	}
	this.Runner.Infof("reader pulsar run Exit")
}
