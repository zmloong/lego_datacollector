package pulsar

import (
	"fmt"
	"lego_datacollector/modules/datacollector/core"
	"lego_datacollector/modules/datacollector/sender"

	"github.com/apache/pulsar-client-go/pulsar"
	lgpulsar "github.com/liwei1dao/lego/sys/pulsar"
)

type Sender struct {
	sender.Sender
	options IOptions
	pulsars []lgpulsar.IPulsar
}

func (this *Sender) Init(rer core.IRunner, ser core.ISender, options core.ISenderOptions) (err error) {
	this.Sender.Init(rer, ser, options)
	this.options = options.(IOptions)
	this.pulsars = make([]lgpulsar.IPulsar, 0)
	return
}

func (this *Sender) Start() (err error) {
	var pulsar lgpulsar.IPulsar
	if this.Procs < 1 {
		this.Procs = 1
	}
	if pipe, ok := this.Runner.SenderPipe(this.GetType()); !ok {
		err = fmt.Errorf("no found SenderPope:%s", this.GetType())
		return
	} else {
		for i := 0; i < this.Procs; i++ {
			if pulsar, err = this.createpulsarclient(); err == nil {
				this.Wg.Add(1)
				this.pulsars = append(this.pulsars, pulsar)
				go this.Run(i, pipe, pulsar)
			} else {
				this.Runner.Errorf("Send pulsar err:%v", err)
				return
			}
		}
	}
	if pulsar, err = this.createpulsarclient(); err == nil {
		this.pulsars = append(this.pulsars, pulsar)
		go this.pulsar_msgerrhandle(pulsar)
	}
	return
}

func (this *Sender) Send(pipeId int, bucket core.ICollDataBucket, params ...interface{}) {
	if pls, ok := params[0].(lgpulsar.IPulsar); ok {
		for _, v := range bucket.SuccItems() {
			var (
				message []byte
				err     error
			)
			if message, err = v.MarshalJSON(); err == nil {
				msg := &pulsar.ProducerMessage{
					Payload: message,
				}
				pls.Producer_SendAsync() <- msg
			} else {
				v.SetError(err)
			}
		}
		this.Sender.Send(pipeId, bucket)
		return
	} else {
		this.Runner.Errorf("Sender kafka err:params[0] no is pulsar")
	}
}

func (this *Sender) createpulsarclient() (sys lgpulsar.IPulsar, err error) {
	sys, err = lgpulsar.NewSys()
	return
}

func (this *Sender) pulsar_msgerrhandle(pls lgpulsar.IPulsar) {
	go func() {
		for v := range pls.Producer_Errors() {
			pls.Producer_SendAsync() <- &pulsar.ProducerMessage{
				Payload: v.Msg.Payload,
			}
		}
	}()
}
