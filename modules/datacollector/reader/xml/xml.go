package xml

import (
	"fmt"
	"lego_datacollector/modules/datacollector/core"
	"lego_datacollector/modules/datacollector/metaer/file"
	"lego_datacollector/modules/datacollector/metaer/folder"
	"lego_datacollector/modules/datacollector/reader"
	"sync"

	"github.com/axgle/mahonia"
	lgcron "github.com/liwei1dao/lego/sys/cron"
	"github.com/robfig/cron/v3"
)

type Reader struct {
	reader.Reader
	options      IOptions //以接口对象传递参数 方便后期继承扩展
	meta         folder.IFolderMetaData
	decoder      mahonia.Decoder //字符串解码器
	runId        cron.EntryID
	rungoroutine int32 //运行采集携程数
	lock         sync.Mutex
	wg           sync.WaitGroup      //用于等待采集器完全关闭
	colltask     chan *file.FileMeta //文件采集任务
	collection   func()
}

func (this *Reader) Init(runner core.IRunner, reader core.IReader, meta core.IMetaerData, options core.IReaderOptions) (err error) {
	if err = this.Reader.Init(runner, reader, meta, options); err != nil {
		return
	}
	this.meta = meta.(folder.IFolderMetaData)
	this.options = options.(IOptions)
	this.colltask = make(chan *file.FileMeta)
	if this.options.GetXml_collectiontype() == LoaclCollection {
		this.collection = this.local_collection
	} else if this.options.GetXml_collectiontype() == SFTPCollection {
		this.collection = this.sftp_collection
	} else if this.options.GetXml_collectiontype() == FTPCollection {
		this.collection = this.ftp_collection
	} else {
		err = fmt.Errorf("collectiontype err:%d", this.options.GetXml_collectiontype())
	}
	return
}
func (this *Reader) Start() (err error) {
	err = this.Reader.Start()

	if this.runId, err = lgcron.AddFunc(this.options.GetXml_interval(), this.collection); err != nil {
		err = fmt.Errorf("lgcron.AddFunc corn:%s err:%v", this.options.GetXml_interval(), err)
		return
	}
	if this.options.GetXml_exec_onstart() {
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
