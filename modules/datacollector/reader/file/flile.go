package file

import (
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"lego_datacollector/comm"
	"lego_datacollector/modules/datacollector/core"
	"lego_datacollector/modules/datacollector/metaer/file"
	"lego_datacollector/modules/datacollector/reader"
	"lego_datacollector/utils/fileutils"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/liwei1dao/lego"
	lgcron "github.com/liwei1dao/lego/sys/cron"
	"github.com/liwei1dao/lego/sys/event"
	"github.com/robfig/cron/v3"
)

type Reader struct {
	reader.Reader
	options      IOptions           //以接口对象传递参数 方便后期继承扩展
	meta         file.IFileMetaData //愿数据
	rungoroutine int32              //运行采集携程数
	isend        int32              //采集任务结束
	wg           sync.WaitGroup
	runId        cron.EntryID
}

func (this *Reader) GetMetaerData() (meta core.IMetaerData) {
	return this.meta
}
func (this *Reader) Init(runner core.IRunner, reader core.IReader, meta core.IMetaerData, options core.IReaderOptions) (err error) {
	if err = this.Reader.Init(runner, reader, meta, options); err != nil {
		return
	}
	this.meta = meta.(file.IFileMetaData)
	this.options = options.(IOptions)
	if _, _, err = fileutils.GetRealPath(this.options.GetFile_path()); err != nil {
		return
	}
	return
}
func (this *Reader) Start() (err error) {
	err = this.Reader.Start()
	atomic.StoreInt32(&this.rungoroutine, 0)
	atomic.StoreInt32(&this.isend, 0)
	if !this.options.GetAutoClose() {
		if this.runId, err = lgcron.AddFunc(this.options.GetFile_scan_initial(), this.collection); err != nil {
			err = fmt.Errorf("lgcron.AddFunc corn:%s err:%v", this.options.GetFile_scan_initial(), err)
			return
		}
		if this.options.GetFile_exec_onstart() {
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

func (this *Reader) collection() {
	if !atomic.CompareAndSwapInt32(&this.isend, 0, 1) {
		this.Runner.Debugf("Reader is collectioning rungoroutine:%d", atomic.LoadInt32(&this.rungoroutine))
		this.SyncMeta()
		return
	}
	this.Runner.Debugf("Reader scan start!")
	if fp, ok := this.checkcollection(); ok { //校验是否需要采集
		this.Runner.Debugf("start collection fp:%s ok:%v", fp, ok)
		if f, err := os.Open(fp); err != nil {
			this.Runner.Debugf("start collection fp:%s Open err:%v", fp, err)
			if this.options.GetAutoClose() { //自动关闭
				event.TriggerEvent(comm.Event_WriteLog, this.Runner.Name(), this.Runner.InstanceId(), fmt.Sprintf("采集数据失败 错误:%v", err), 1, time.Now().Unix())
				go this.AutoClose(core.RunnerFailAutoClose)
			} else {
				atomic.StoreInt32(&this.isend, 0)
			}
			return
		} else {
			atomic.AddInt32(&this.rungoroutine, 1)
			this.wg.Add(1)
			go this.asyncollection(fp, f)
		}
	} else {
		this.Runner.Infof("Reader Scan file no found!")
		if this.options.GetAutoClose() { //自动关闭
			event.TriggerEvent(comm.Event_WriteLog, this.Runner.Name(), this.Runner.InstanceId(), "未找到目标数据表!", 1, time.Now().Unix())
			go this.AutoClose(core.RunnerFailAutoClose)
		} else {
			atomic.StoreInt32(&this.isend, 0)
		}
	}
}

func (this *Reader) asyncollection(fp string, f *os.File) {
	defer lego.Recover(fmt.Sprintf("%s Reader", this.Runner.Name()))
	defer atomic.AddInt32(&this.rungoroutine, -1)
clocp:
	for {
		if end, err := this.collection_file(fp, f); end || err != nil {
			this.Runner.Debugf("Reader collection_file ok:%v,err:%v", end, err)
			f.Close()
			break clocp
		}
		if this.Runner.GetRunnerState() == core.Runner_Stoping { //采集器进入停止过程中
			f.Close()
			break clocp
		}
	}
	this.SyncMeta()
	this.wg.Done()
	if atomic.CompareAndSwapInt32(&this.isend, 1, 2) { //最后一个任务已经完成
		this.collectionend()
	}
	this.Runner.Debugf("Reader asyncollection exit succ!")
}

///采集结束
func (this *Reader) collectionend() {
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

//----------------------------------------------------------------------------------------
func (this *Reader) checkcollection() (fp string, ok bool) {
	var (
		err   error
		finfo fs.FileInfo
	)
	if fp, finfo, err = fileutils.GetRealPath(this.options.GetFile_path()); err != nil {
		ok = false
		return
	}
	if uint64(finfo.Size()) > this.meta.GetFileAlreadyReadOffset() { //有新内容
		this.meta.SetFileName(fp)
		this.meta.SetFileSize(uint64(finfo.Size()))
		this.meta.SetFileLastModifyTime(finfo.ModTime().Unix())
		ok = true
	}
	return
}

func (this *Reader) collection_file(fp string, f *os.File) (isend bool, err error) {
	var (
		buff  = make([]byte, this.options.GetFile_readbuff_size())
		rbuff *bytes.Buffer
		n     int
	)
	if n, err = f.ReadAt(buff, int64(this.meta.GetFileAlreadyReadOffset())+int64(len(this.meta.GetFileCacheData()))); err != nil && err != io.EOF {
		return
	}
	if n > 0 {
		rbuff = bytes.NewBuffer(buff[:n])
		var (
			line string
			err  error
		)
		for {
			if line, err = rbuff.ReadString('\n'); err != nil {
				// this.Runner.Debugf("collection fp:%s lien:%s", fp, line)
				if err == io.EOF {
					if int64(len(line))+int64(len(this.meta.GetFileCacheData()))+int64(this.meta.GetFileAlreadyReadOffset()) >= int64(this.meta.GetFileSize()) { //读完了
						isend = this.writeDataChan(line)
					} else {
						this.writeDataCache(line)
					}
				}
				break
			} else {
				isend = this.writeDataChan(line)
			}
		}
	} else {
		isend = true
	}
	return
}

//发送数据 数据量打的时候会有堵塞
func (this *Reader) writeDataChan(line string) (isend bool) {
	var (
		value  = line
		cache  = this.meta.GetFileCacheData()
		offset = this.meta.GetFileAlreadyReadOffset()
	)
	if cache != nil && len(cache) != 0 {
		var buffer bytes.Buffer
		buffer.Write(cache)
		buffer.WriteString(line)
		value = buffer.String()
		this.meta.ResetCacheData()
	}
	this.Input() <- core.NewCollData(this.Runner.ServiceIP(), value)
	offset += uint64(len(value))
	this.meta.SetFileAlreadyReadOffset(offset)
	if offset >= this.meta.GetFileSize() { //文件读取完毕
		this.meta.SetFileLastCollectionTime(time.Now().Unix())
		isend = true
	}
	return
}

func (this *Reader) writeDataCache(line string) {
	var (
		cache      = this.meta.GetFileCacheData()
		buffer     bytes.Buffer
		offset     uint64
		cacheCount uint64
	)
	if len(cache) > 0 {
		buffer.Write(cache)
	}
	buffer.WriteString(line)
	cacheCount = this.meta.SetFileCacheData(buffer.Bytes())
	if cacheCount > this.Runner.MaxCollDataSzie() { //大于最大采集数据 直接丢弃
		this.Runner.Errorf("清理缓存%s", buffer.String())
		offset = this.meta.GetFileAlreadyReadOffset() + cacheCount
		this.meta.SetFileAlreadyReadOffset(offset)
		this.meta.ResetCacheData()
	}
}
