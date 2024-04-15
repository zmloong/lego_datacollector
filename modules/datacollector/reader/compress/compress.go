package compress

import (
	"fmt"
	"lego_datacollector/modules/datacollector/core"

	compress "lego_datacollector/modules/datacollector/metaer/compress"

	"lego_datacollector/modules/datacollector/reader"
	"sync"

	"github.com/axgle/mahonia"
	lgcron "github.com/liwei1dao/lego/sys/cron"
	"github.com/robfig/cron/v3"
)

type Reader struct {
	reader.Reader
	options      IOptions //以接口对象传递参数 方便后期继承扩展
	meta         compress.IFolderMetaData
	decoder      mahonia.Decoder //字符串解码器
	runId        cron.EntryID
	rungoroutine int32 //运行采集携程数
	isend        int32 //采集任务结束
	lock         sync.Mutex
	wg           sync.WaitGroup                //用于等待采集器完全关闭
	colltask     chan *compress.CopresFileMeta //文件采集任务
	collection   func()
}

func (this *Reader) GetMetaerData() (meta core.IMetaerData) {
	return this.meta
}

func (this *Reader) Init(runner core.IRunner, reader core.IReader, meta core.IMetaerData, options core.IReaderOptions) (err error) {
	if err = this.Reader.Init(runner, reader, meta, options); err != nil {
		return
	}
	this.meta = meta.(compress.IFolderMetaData)
	this.options = options.(IOptions)
	this.colltask = make(chan *compress.CopresFileMeta)
	if this.options.GetCompress_collectiontype() == FTPCollection {
		this.collection = this.ftp_collection
	} else {
		err = fmt.Errorf("collectiontype err:%d", this.options.GetCompress_collectiontype())
	}
	return
}
func (this *Reader) Start() (err error) {
	err = this.Reader.Start()

	if this.runId, err = lgcron.AddFunc(this.options.GetCompress_interval(), this.collection); err != nil {
		err = fmt.Errorf("lgcron.AddFunc corn:%s err:%v", this.options.GetCompress_interval(), err)
		return
	}
	if this.options.GetCompress_exec_onstart() {
		go this.collection()
	}
	return
}

///外部调度器 驱动执行  此接口 不可阻塞
func (this *Reader) Drive() (err error) {
	err = this.Reader.Drive()
	go this.collection()
	return
}

//关闭 关闭接口只能有上层runner调用
func (this *Reader) Close() (err error) {
	lgcron.Remove(this.runId)
	this.wg.Wait()
	err = this.Reader.Close()
	return
}

///采集结束
func (this *Reader) collectionend() {
	close(this.colltask)
}
