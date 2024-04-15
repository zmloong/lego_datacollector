package canal

import (
	"lego_datacollector/modules/datacollector/core"
	"lego_datacollector/modules/datacollector/reader"
	"sync"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/liwei1dao/lego/sys/log"
	"github.com/withlin/canal-go/client"
	"github.com/withlin/canal-go/protocol"
	pbe "github.com/withlin/canal-go/protocol/entry"
)

type Reader struct {
	reader.Reader
	options IOptions //以接口对象传递参数 方便后期继承扩展
	client  *client.SimpleCanalConnector
	wg      sync.WaitGroup //用于等待采集器完全关闭
}

func (this *Reader) Init(runner core.IRunner, reader core.IReader, meta core.IMetaerData, options core.IReaderOptions) (err error) {
	if err = this.Reader.Init(runner, reader, meta, options); err != nil {
		return
	}
	this.options = options.(IOptions)
	// 192.168.199.17 替换成你的canal server的地址
	// example 替换成-e canal.destinations=example 你自己定义的名字
	this.client = client.NewSimpleCanalConnector(
		this.options.GetCanal_Addr(),
		this.options.GetCanal_Port(),
		this.options.GetCanal_User(),
		this.options.GetCanal_Password(),
		this.options.GetCanal_Destination(), 60000, 60*60*1000)
	err = this.client.Connect()
	if err != nil {
		this.Runner.Errorf("Connect err:%v", err)
	}
	return
}

func (this *Reader) Start() (err error) {
	if err = this.Reader.Start(); err != nil {
		return
	}
	err = this.client.Subscribe(this.options.GetCanal_Filter())
	if err != nil {
		this.Runner.Errorf("Start err:%v", err)
	}
	this.wg.Add(1)
	go this.run()
	return
}

func (this *Reader) Close() (err error) {
	err = this.client.DisConnection()
	this.wg.Wait()
	return this.Reader.Close()
}

func (this *Reader) run() {
	defer this.wg.Done()
	var (
		err     error
		message *protocol.Message
		batchId int64
		entry   pbe.Entry
		data    *ReaderData
	)
	for {
		message, err = this.client.Get(100, nil, nil)
		if err != nil {
			this.Runner.Errorf("run err:%v", err)
			break
		}
		batchId = message.Id
		if batchId == -1 || len(message.Entries) <= 0 {
			time.Sleep(300 * time.Millisecond)
			// fmt.Println("===没有数据了===")
			continue
		}
		for _, entry = range message.Entries {
			if entry.GetEntryType() == pbe.EntryType_TRANSACTIONBEGIN || entry.GetEntryType() == pbe.EntryType_TRANSACTIONEND {
				continue
			}
			/*
				数据读取逻辑
			*/
			rowChange := new(pbe.RowChange)
			if err = proto.Unmarshal(entry.GetStoreValue(), rowChange); err != nil {
				this.Runner.Errorf("Reader err:%v", err)
			} else {
				if rowChange != nil {
					header := entry.GetHeader()
					data = &ReaderData{
						EventType:     getEventType(rowChange.GetEventType()),
						LogfileName:   header.GetLogfileName(),
						LogfileOffset: header.GetLogfileOffset(),
						SchemaName:    header.GetSchemaName(),
						TableName:     header.GetTableName(),
						RowDatas:      rowChange.GetRowDatas(),
					}
					str, _ := data.ToString()
					log.Debugf("str:%s", str)
					this.Reader.Input() <- core.NewCollData(this.options.GetCanal_Addr(), str)
				}
			}
		}
	}
	this.Runner.Debugf("Reade exit succ!")
}
