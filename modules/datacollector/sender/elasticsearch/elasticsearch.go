package elasticsearch

import (
	"context"
	"lego_datacollector/modules/datacollector/core"
	"lego_datacollector/modules/datacollector/sender"

	"github.com/olivere/elastic/v7"
)

type Sender struct {
	sender.Sender
	options IOptions
	client  *elastic.Client
}

func (this *Sender) Init(rer core.IRunner, ser core.ISender, options core.ISenderOptions) (err error) {
	this.Sender.Init(rer, ser, options)
	this.options = options.(IOptions)
	client, err := elastic.NewClient(
		elastic.SetSniff(false),
		elastic.SetHealthcheck(false),
		elastic.SetURL(this.options.GetEs_host()))
	if err != nil {
		this.Runner.Debugf("Sender Init elastic Connect err:", err)
	}
	this.client = client
	return
}

//关闭
func (this *Sender) Close() (err error) {
	if err = this.Sender.Close(); err != nil {
		this.Runner.Errorf("Sender Close err:%v", err)
		return
	}
	this.client.Stop()
	return
}

func (this *Sender) Send(pipeId int, bucket core.ICollDataBucket, params ...interface{}) {

	bulkRequest := this.client.Bulk()
	for _, v := range bucket.SuccItems() {
		var (
			message = ""
			err     error
		)
		if message, err = v.ToString(); err == nil {
			//this.es_Create(message)
			req := elastic.NewBulkIndexRequest().Index(this.options.GetEs_index()).Doc(string(message))
			bulkRequest.Add(req)
		} else {
			v.SetError(err)
		}
	}
	_, err := bulkRequest.Do(context.Background())
	if err != nil {
		bucket.SetError(err)
	}
	this.Sender.Send(pipeId, bucket)
	return

}

func (this *Sender) es_Create(data string) (err error) {
	_, err = this.client.Index().
		Index(this.options.GetEs_index()).
		BodyJson(data).
		Do(context.Background())
	if err != nil {
		return
	}
	return
}
