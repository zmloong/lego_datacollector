package gmsm2

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"lego_datacollector/modules/datacollector/core"
	"lego_datacollector/modules/datacollector/metaer/file"
	"lego_datacollector/modules/datacollector/metaer/folder"
	"lego_datacollector/modules/datacollector/reader"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ZZMarquis/gm/sm2"
	lgcron "github.com/liwei1dao/lego/sys/cron"
	"github.com/liwei1dao/lego/utils/crypto/gm_java"
	"github.com/robfig/cron/v3"
)

type Reader struct {
	reader.Reader
	options      IOptions //以接口对象传递参数 方便后期继承扩展
	meta         folder.IFolderMetaData
	runId        cron.EntryID
	rungoroutine int32               //运行采集携程数
	isend        int32               //采集任务结束
	wg           sync.WaitGroup      //用于等待采集器完全关闭
	colltask     chan *file.FileMeta //文件采集任务
	collection   func()
	privateKey   *sm2.PrivateKey
	datasourceIP string
}

func (this *Reader) Init(runner core.IRunner, reader core.IReader, meta core.IMetaerData, options core.IReaderOptions) (err error) {
	if err = this.Reader.Init(runner, reader, meta, options); err != nil {
		return
	}
	this.meta = meta.(folder.IFolderMetaData)
	this.options = options.(IOptions)
	this.colltask = make(chan *file.FileMeta)
	if this.options.GetGmsm2_collectiontype() == LoaclCollection {
		this.collection = this.local_collection
		this.datasourceIP = this.Runner.ServiceIP()
	} else if this.options.GetGmsm2_collectiontype() == SFTPCollection {
		this.collection = this.sftp_collection
		this.datasourceIP = this.options.GetGmsm2_server_addr()
	} else if this.options.GetGmsm2_collectiontype() == FTPCollection {
		this.collection = this.ftp_collection
		this.datasourceIP = this.options.GetGmsm2_server_addr()
	} else {
		err = fmt.Errorf("collectiontype err:%d", this.options.GetGmsm2_collectiontype())
		return
	}
	this.privateKey, err = gm_java.ReadPrivateKeyFormBase64(this.options.GetGmsm2_prikey())
	return
}
func (this *Reader) Start() (err error) {
	err = this.Reader.Start()
	atomic.StoreInt32(&this.rungoroutine, 0)
	atomic.StoreInt32(&this.isend, 0)
	if !this.options.GetAutoClose() {
		if this.runId, err = lgcron.AddFunc(this.options.GetGmsm2_interval(), this.collection); err != nil {
			err = fmt.Errorf("lgcron.AddFunc corn:%s err:%v", this.options.GetGmsm2_interval(), err)
			return
		}
		if this.options.GetGmsm2_exec_onstart() {
			go this.collection()
		}
	} else {
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
	if !this.options.GetAutoClose() {
		lgcron.Remove(this.runId)
	}
	this.wg.Wait()
	err = this.Reader.Close()
	return
}

///采集结束
func (this *Reader) collectionend() {
	close(this.colltask)
	this.wg.Wait()
	atomic.StoreInt32(&this.isend, 0)
	this.Runner.Debugf("Reader collection end succ!")
	if this.options.GetAutoClose() { //自动关闭
		go this.AutoClose(core.RunnerSuccAutoClose)
	}
}

func (this *Reader) AutoClose(msg string) {
	time.Sleep(time.Second * 10)
	this.Runner.Close(core.Runner_Runing, msg) //结束任务
}

///解析数据
func (this *Reader) parse(data []byte) (err error) {
	var (
		mdata    []byte
		origData []byte
		line     string
	)
	// this.Runner.Debugf("Reader parse satrt!")
	if mdata, err = base64.StdEncoding.DecodeString(string(data)); err == nil {
		if origData, err = gm_java.SM2_Decrypt(this.privateKey, mdata); err == nil {
			rbuff := bytes.NewBuffer(origData[:])
			for {
				if line, err = rbuff.ReadString('\n'); err != nil {
					if err == io.EOF && len(line) > 0 {
						// this.Runner.Debugf("Reader parse :%v", line)
						this.Input() <- core.NewCollData(this.datasourceIP, line)
					}
					err = nil
					break
				} else {
					// this.Runner.Debugf("Reader parse :%v", line)
					this.Input() <- core.NewCollData(this.datasourceIP, line)
				}
			}
		} else {
			err = fmt.Errorf("SM2_Decrypt err:%v", err)
		}
	} else {
		err = fmt.Errorf("Base64 DecodeString err:%v", err)
	}
	return
}
