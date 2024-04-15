package kafka

import (
	"fmt"
	"lego_datacollector/modules/datacollector/core"
	"lego_datacollector/modules/datacollector/sender"

	"github.com/Shopify/sarama"
	"github.com/liwei1dao/lego/sys/kafka"
)

type Sender struct {
	sender.Sender
	options IOptions
	kafkas  []kafka.IKafka
}

func (this *Sender) Init(rer core.IRunner, ser core.ISender, options core.ISenderOptions) (err error) {
	this.Sender.Init(rer, ser, options)
	this.options = options.(IOptions)
	this.kafkas = make([]kafka.IKafka, 0)
	return
}

func (this *Sender) Start() (err error) {
	var kafka kafka.IKafka
	if this.Procs < 1 {
		this.Procs = 1
	}
	if pipe, ok := this.Runner.SenderPipe(this.GetType()); !ok {
		err = fmt.Errorf("no found SenderPope:%s", this.GetType())
		return
	} else {
		for i := 0; i < this.Procs; i++ {
			if kafka, err = this.createkafkaclient(); err == nil {
				this.Wg.Add(1)
				this.kafkas = append(this.kafkas, kafka)
				go this.kafka_msgerrhandle(kafka)
				go this.Run(i, pipe, kafka)
			} else {
				this.Runner.Errorf("Send kafka err:%v", err)
				return
			}
		}
	}
	return
}

func (this *Sender) Run(pipeId int, pipe <-chan core.ICollDataBucket, params ...interface{}) {
	defer this.Wg.Done()
	for v := range pipe {
		this.Send(pipeId, v, params...)
	}
}

//关闭
func (this *Sender) Close() (err error) {
	if err = this.Sender.Close(); err != nil {
		this.Runner.Errorf("Sender Close err:%v", err)
	}
	for _, v := range this.kafkas {
		v.Asyncproducer_Close()
	}
	return
}

func (this *Sender) Send(pipeId int, bucket core.ICollDataBucket, params ...interface{}) {
	if kafka, ok := params[0].(kafka.IKafka); ok {
		for _, v := range bucket.SuccItems() {
			var (
				message = ""
				err     error
			)
			if message, err = v.ToString(); err == nil {
				msg := &sarama.ProducerMessage{
					Topic: this.options.GetKafka_topic(),
					Value: sarama.StringEncoder(message),
				}
				kafka.Asyncproducer_Input() <- msg
			} else {
				v.SetError(err)
			}
		}
		this.Sender.Send(pipeId, bucket)
		return
	} else {
		this.Runner.Errorf("Sender kafka err:params[0] no is kafka")
	}
}

func (this *Sender) createkafkaclient() (sys kafka.IKafka, err error) {
	sys, err = kafka.NewSys(
		kafka.SetStartType(kafka.Asyncproducer),
		kafka.SetHosts(this.options.GetKafka_host()),
		kafka.SetClientID(this.options.GetKafka_client_id()),
		kafka.SetNet_DialTimeout(this.options.GetNet_DialTimeout()),
		kafka.SetNet_KeepAlive(this.options.GetNet_KeepAlive()),
		kafka.SetProducer_MaxMessageBytes(int(this.Runner.MaxCollDataSzie())),
		kafka.SetProducer_Compression(this.options.GetProducer_Compression()),
		kafka.SetProducer_Return_Errors(true),
		kafka.SetProducer_CompressionLevel(this.options.GetProducer_CompressionLevel()),
		kafka.SetProducer_Retry_Max(this.options.GetKafka_retry_max()),
	)
	return
}

func (this *Sender) kafka_msgerrhandle(kafka kafka.IKafka) {
	go func() {
		for v := range kafka.Asyncproducer_Errors() {
			this.Runner.Errorf("Sender kafka err:%v", v.Err)
			kafka.Asyncproducer_Input() <- &sarama.ProducerMessage{
				Topic: this.options.GetKafka_topic(),
				Value: v.Msg.Value,
			}
		}
	}()
}
