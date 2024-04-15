package kafka

import (
	"lego_datacollector/modules/datacollector/core"
	"lego_datacollector/modules/datacollector/reader"

	"github.com/Shopify/sarama"
	"github.com/liwei1dao/lego/sys/kafka"
	"github.com/liwei1dao/lego/sys/log"
)

type Reader struct {
	reader.Reader
	options   IOptions      //以接口对象传递参数 方便后期继承扩展
	closesign chan struct{} //关闭信号管道
	kafka     kafka.IKafka
}

func (this *Reader) Init(runner core.IRunner, reader core.IReader, meta core.IMetaerData, options core.IReaderOptions) (err error) {
	if err = this.Reader.Init(runner, reader, meta, options); err != nil {
		return
	}
	this.options = options.(IOptions)
	this.closesign = make(chan struct{})
	this.kafka, err = kafka.NewSys(
		kafka.SetStartType(kafka.Consumer), //消费者模式
		kafka.SetHosts(this.options.GetKafka_zookeeper()),
		kafka.SetTopics(this.options.GetKafka_topic()),
		kafka.SetGroupId(this.options.GetKafka_groupid()),
		kafka.SetClientID(this.Runner.ServiceId()),
		kafka.SetConsumer_Offsets_Initial(this.options.Getconsumer_Offsets_Initial()),
		kafka.SetSasl_Enable(this.options.GetKafka_Kerberos_Enable()),
		kafka.SetSasl_Mechanism(sarama.SASLTypeGSSAPI),
		kafka.SetSasl_GSSAPI(sarama.GSSAPIConfig{
			AuthType:           sarama.KRB5_KEYTAB_AUTH,
			Realm:              this.options.GetKafka_Kerberos_Realm(),
			ServiceName:        this.options.GetKafka_Kerberos_ServiceName(),
			Username:           this.options.GetKafka_Kerberos_Username(),
			KeyTabPath:         this.options.GetKafka_Kerberos_KeyTabPath(),
			KerberosConfigPath: this.options.GetKafka_Kerberos_KerberosConfigPath(),
			DisablePAFXFAST:    true,
		}),
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

func (this *Reader) SyncMeta() {
}

func (this *Reader) Close() (err error) {
	if err = this.kafka.Consumer_Close(); err != nil {
		log.Errorf("Close Reader err:%v", err)
		return
	}
	this.Runner.Infof("reader kafka Close succ")
	err = this.Reader.Close()
	return
}

//外部协成启动
func (this *Reader) run() {
	for v := range this.kafka.Consumer_Messages() {
		// fmt.Sprintf("Consumer_Messages:%d -------1", v.Offset)
		this.Input() <- core.NewCollData("", v.Value)
		// log.Debugf("Consumer_Messages:%d -------2", v.Offset)
	}
	this.Runner.Infof("reader kafka run Exit")
}
