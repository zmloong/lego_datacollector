package xls

import (
	"fmt"
	"lego_datacollector/modules/datacollector/core"
	"lego_datacollector/modules/datacollector/metaer/file"
	"lego_datacollector/modules/datacollector/metaer/folder"
	"lego_datacollector/modules/datacollector/reader"
	"sync"
	"time"

	"github.com/axgle/mahonia"
	lgcron "github.com/liwei1dao/lego/sys/cron"
	"github.com/robfig/cron/v3"
	"golang.org/x/crypto/ssh"
)

type Reader struct {
	reader.Reader
	options      IOptions //以接口对象传递参数 方便后期继承扩展
	meta         folder.IFolderMetaData
	sshClient    *ssh.Client
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
	if this.options.GetXls_collectiontype() == LoaclCollection {
		this.collection = this.local_collection
	} else if this.options.GetXls_collectiontype() == SFTPCollection {
		this.collection = this.sftp_collection
		clientConfig := &ssh.ClientConfig{
			User:            this.options.GetXls_server_user(),
			Auth:            []ssh.AuthMethod{ssh.Password(this.options.GetXls_server_password())},
			Timeout:         10 * time.Second,
			HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		}
		if this.sshClient, err = ssh.Dial("tcp", fmt.Sprintf("%s:%d", this.options.GetXls_server_addr(), this.options.GetXls_server_port()), clientConfig); err != nil {
			return
		}
	} else if this.options.GetXls_collectiontype() == FTPCollection {
		this.collection = this.ftp_collection
	} else {
		err = fmt.Errorf("collectiontype err:%d", this.options.GetXls_collectiontype())
	}
	return
}
func (this *Reader) Start() (err error) {
	err = this.Reader.Start()

	if this.runId, err = lgcron.AddFunc(this.options.GetXls_interval(), this.collection); err != nil {
		err = fmt.Errorf("lgcron.AddFunc corn:%s err:%v", this.options.GetXls_interval(), err)
		return
	}
	if this.options.GetXls_exec_onstart() {
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
	if this.sshClient != nil {
		this.sshClient.Close()
	}
	err = this.Reader.Close()
	return
}

///采集结束
func (this *Reader) collectionend() {
	close(this.colltask)
}
