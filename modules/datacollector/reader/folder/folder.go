package folder

import (
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"io/ioutil"
	"lego_datacollector/comm"
	"lego_datacollector/modules/datacollector/core"
	"lego_datacollector/modules/datacollector/metaer/file"
	"lego_datacollector/modules/datacollector/metaer/folder"
	"lego_datacollector/modules/datacollector/reader"
	"os"
	"path"
	"regexp"
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
	meta         folder.IFolderMetaData //愿数据
	options      IOptions               //以接口对象传递参数 方便后期继承扩展
	runId        cron.EntryID
	rungoroutine int32 //运行采集携程数
	isend        int32 //采集任务结束
	// lock         sync.Mutex
	wg       sync.WaitGroup      //用于等待采集器完全关闭
	colltask chan *file.FileMeta //文件采集任务
}

func (this *Reader) GetMetaerData() (meta core.IMetaerData) {
	return this.meta
}

func (this *Reader) Init(runner core.IRunner, reader core.IReader, meta core.IMetaerData, options core.IReaderOptions) (err error) {
	if err = this.Reader.Init(runner, reader, meta, options); err != nil {
		return
	}
	this.meta = meta.(folder.IFolderMetaData)
	this.options = options.(IOptions)
	this.colltask = make(chan *file.FileMeta)
	return
}
func (this *Reader) Start() (err error) {
	err = this.Reader.Start()
	atomic.StoreInt32(&this.rungoroutine, 0)
	atomic.StoreInt32(&this.isend, 0)
	if !this.options.GetAutoClose() {
		if this.runId, err = lgcron.AddFunc(this.options.GetFolder_scan_initial(), this.collection); err != nil {
			err = fmt.Errorf("lgcron.AddFunc corn:%s err:%v", this.options.GetFolder_scan_initial(), err)
			return
		}
		if this.options.GetFolder_exec_onstart() {
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
	// this.lock.Lock()
	// defer this.lock.Unlock()
	var (
		files           = make(map[string]fs.FileInfo)
		collectionfiles = make([]*file.FileMeta, 0)
		f               *file.FileMeta
		ok              bool
	)
	if !atomic.CompareAndSwapInt32(&this.isend, 0, 1) {
		this.Runner.Debugf("Reader is collectioning rungoroutine:%d", atomic.LoadInt32(&this.rungoroutine))
		this.SyncMeta()
		return
	}
	this.Runner.Debugf("Reader scan start!")
	if err := this.scanDirectory(this.options.GetFolder_directory(), files); err == nil {
		for k, v := range files {
			if f, ok = this.meta.GetFile(k); !ok {
				f = &file.FileMeta{
					FileName:           k,
					FileSize:           uint64(v.Size()),
					FileLastModifyTime: v.ModTime().Unix(),
					FileCacheData:      make([]byte, 0),
				}
				this.meta.SetFile(k, f)
			}
			if f.FileAlreadyReadOffset < uint64(v.Size()) {
				f.FileSize = uint64(v.Size())
				collectionfiles = append(collectionfiles, f)
			}
		}

		if len(collectionfiles) > 0 {
			procs := this.Runner.MaxProcs()
			if len(collectionfiles) < this.Runner.MaxProcs() {
				procs = len(collectionfiles)
			}
			this.colltask = make(chan *file.FileMeta, len(collectionfiles))
			for _, v := range collectionfiles {
				this.colltask <- v
			}
			atomic.AddInt32(&this.rungoroutine, int32(procs))
			this.wg.Add(procs)
			for i := 0; i < procs; i++ {
				go this.asyncollection(this.colltask)
			}
		} else {
			this.Runner.Debugf("Reader scan no found new data!")
			if this.options.GetAutoClose() { //自动关闭
				event.TriggerEvent(comm.Event_WriteLog, this.Runner.Name(), this.Runner.InstanceId(), "未扫描到新的可采集数据!", 1, time.Now().Unix())
				go this.AutoClose(core.RunnerFailAutoClose)
			} else {
				atomic.StoreInt32(&this.isend, 0)
			}
		}
	} else {
		this.Runner.Infof("Reader folder Scan Directory err:%v", err)
		if this.options.GetAutoClose() { //自动关闭
			event.TriggerEvent(comm.Event_WriteLog, this.Runner.Name(), this.Runner.InstanceId(), fmt.Sprintf("扫描目标目录失败 err:%v", err), 1, time.Now().Unix())
			go this.AutoClose(core.RunnerFailAutoClose)
		} else {
			atomic.StoreInt32(&this.isend, 0)
		}
	}
}
func (this *Reader) asyncollection(fmeta <-chan *file.FileMeta) {
	defer lego.Recover(fmt.Sprintf("%s Reader", this.Runner.Name()))
	defer atomic.AddInt32(&this.rungoroutine, -1)
	var (
		f   *os.File
		err error
	)
locp:
	for v := range fmeta {
		if f, err = os.Open(v.FileName); err == nil {
		clocp:
			for {
				if ok, err := this.collection_file(v, f); ok || err != nil {
					this.Runner.Debugf("Reader asyncollection ok:%v,err:%v", ok, err)
					f.Close()
					break clocp
				}
				if this.Runner.GetRunnerState() == core.Runner_Stoping { //采集器进入停止过程中
					this.Runner.Debugf("Reader asyncollection exit")
					f.Close()
					break locp
				}
			}
			this.Runner.Debugf("Reader SyncMeta")
			this.SyncMeta()
		} else {
			this.Runner.Debugf("Reader asyncollection FileName:%s Open err:%v", v.FileName, err)
		}
		this.Runner.Debugf("Reader SyncMeta")
		if this.Runner.GetRunnerState() == core.Runner_Stoping || len(fmeta) == 0 {
			break locp
		} else {
			this.SyncMeta()
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

//--------------------------------------------------------------------------------------------------------
func (this *Reader) scanDirectory(_path string, files map[string]fs.FileInfo) (err error) {
	var (
		fs      []fs.FileInfo
		matched bool
		fp      string
	)

	if fs, err = ioutil.ReadDir(_path); err == nil {
		for _, v := range fs {
			fp = path.Join(_path, v.Name())
			// this.Runner.Debugf("Reader scanDirectory v:%v", v)
			if v.IsDir() { //是目录
				err = this.scanDirectory(fp, files)
			} else {
				if len(files) >= this.options.GetFolder_max_collection_num() {
					this.Runner.Debugf("Reader scanDirectory up top")
					return
				} else {
					if matched, err = regexp.MatchString(this.options.GetFolder_regular_filter(), v.Name()); err == nil && matched {
						files[fp] = v
					} else {
						this.Runner.Debugf("Reader regular:%s file:%s matched:%v err:%v", this.options.GetFolder_regular_filter(), v.Name(), matched, err)
					}
				}
			}

		}
	}
	return
}

///采集文件
func (this *Reader) collection_file(fmeta *file.FileMeta, f *os.File) (iscollend bool, err error) {
	var (
		buff  = make([]byte, this.options.GetFolder_readerbuf_size())
		rbuff *bytes.Buffer
		n     int
	)
	if n, err = f.ReadAt(buff, int64(fmeta.FileAlreadyReadOffset)+int64(len(fmeta.FileCacheData))); err != nil && err != io.EOF {
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
				if err == io.EOF {
					if int64(len(line))+int64(len(fmeta.FileCacheData))+int64(fmeta.FileAlreadyReadOffset) >= int64(fmeta.FileSize) { //读完了
						iscollend = this.writeDataChan(fmeta, line)
					} else {
						this.writeDataCache(fmeta, line)
					}
				}
				break
			} else {
				iscollend = this.writeDataChan(fmeta, line)
			}
		}
	} else {
		iscollend = true
	}
	return
}

//发送数据 数据量打的时候会有堵塞
func (this *Reader) writeDataChan(fmeta *file.FileMeta, line string) (isend bool) {
	var (
		value  = line
		cache  = fmeta.FileCacheData
		offset = fmeta.FileAlreadyReadOffset
	)
	if cache != nil && len(cache) != 0 {
		var buffer bytes.Buffer
		buffer.Write(cache)
		buffer.WriteString(line)
		value = buffer.String()
		fmeta.FileCacheData = make([]byte, 0)
	}
	this.Input() <- core.NewCollData(this.Runner.ServiceIP(), value)
	offset += uint64(len(value))
	fmeta.FileAlreadyReadOffset = offset
	if offset >= uint64(fmeta.FileSize) { //文件读取完毕
		fmeta.FileLastCollectionTime = time.Now().Unix()
		isend = true
	}
	return
}

func (this *Reader) writeDataCache(fmeta *file.FileMeta, line string) {
	var (
		cache  = fmeta.FileCacheData
		buffer bytes.Buffer
	)
	if len(cache) > 0 {
		buffer.Write(cache)
	}
	buffer.WriteString(line)
	fmeta.FileCacheData = buffer.Bytes()
	if uint64(len(fmeta.FileCacheData)) > this.Runner.MaxCollDataSzie() { //大于最大采集数据 直接丢弃
		this.Runner.Errorf("清理缓存%s", buffer.String())
		fmeta.FileAlreadyReadOffset += uint64(len(fmeta.FileCacheData))
		fmeta.FileCacheData = make([]byte, 0)
	}
}
