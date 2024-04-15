package hdfs

import (
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"lego_datacollector/comm"
	"lego_datacollector/modules/datacollector/core"
	"lego_datacollector/modules/datacollector/metaer/file"
	"lego_datacollector/modules/datacollector/metaer/folder"
	"lego_datacollector/modules/datacollector/reader"
	"regexp"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/axgle/mahonia"
	"github.com/colinmarc/hdfs/v2"
	"github.com/colinmarc/hdfs/v2/hadoopconf"
	"github.com/jcmturner/gokrb5/v8/client"
	"github.com/jcmturner/gokrb5/v8/config"
	"github.com/jcmturner/gokrb5/v8/keytab"
	"github.com/liwei1dao/lego"
	lgcron "github.com/liwei1dao/lego/sys/cron"
	"github.com/liwei1dao/lego/sys/event"
	"github.com/robfig/cron/v3"
)

type Reader struct {
	reader.Reader
	options      IOptions
	meta         folder.IFolderMetaData
	decoder      mahonia.Decoder //字符串解码器
	runId        cron.EntryID
	rungoroutine int32 //运行采集携程数
	lock         sync.Mutex
	isend        int32               //采集任务结束
	wg           sync.WaitGroup      //用于等待采集器完全关闭
	colltask     chan *file.FileMeta //文件采集任务
}

func (this *Reader) Init(runner core.IRunner, reader core.IReader, meta core.IMetaerData, options core.IReaderOptions) (err error) {
	this.meta = meta.(folder.IFolderMetaData)
	this.options = options.(IOptions)
	this.colltask = make(chan *file.FileMeta)
	if conn, err := this.connectHdfs(); err != nil {
		return err
	} else {
		conn.Close()
	}

	err = this.Reader.Init(runner, reader, meta, options)
	return
}
func (this *Reader) Start() (err error) {
	err = this.Reader.Start()
	atomic.StoreInt32(&this.rungoroutine, 0)
	atomic.StoreInt32(&this.isend, 0)

	if !this.options.GetAutoClose() {
		if this.runId, err = lgcron.AddFunc(this.options.GetHdfs_interval(), this.collection); err != nil {
			err = fmt.Errorf("lgcron.AddFunc corn:%s err:%v", this.options.GetHdfs_interval(), err)
			return
		}
		if this.options.GetHdfs_exec_onstart() {
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

//高并发采集服务
func (this *Reader) collection() {
	// this.lock.Lock()
	// defer this.lock.Unlock()
	var (
		cfiles map[string]*file.FileMeta
	)
	if !atomic.CompareAndSwapInt32(&this.isend, 0, 1) {
		this.Runner.Debugf("Reader is collectioning rungoroutine:%d", atomic.LoadInt32(&this.rungoroutine))
		this.SyncMeta()
		return
	}
	this.Runner.Debugf("Reader scan start!")
	cfiles = make(map[string]*file.FileMeta)
	if conn, err := this.connectHdfs(); err != nil {
		atomic.StoreInt32(&this.isend, 0)
		this.Runner.Debugf("Reader Dial hdfs err:%v", err)
		return
	} else {
		defer conn.Close()
		fmetas := this.meta.GetFiles()
		this.scandirectory(conn, this.options.GetHdfs_directory(), fmetas, cfiles)
	}
	if len(cfiles) > 0 {
		procs := this.Runner.MaxProcs()
		if len(cfiles) < this.Runner.MaxProcs() {
			procs = len(cfiles)
		}
		this.colltask = make(chan *file.FileMeta, len(cfiles))
		for _, v := range cfiles {
			this.colltask <- v
		}

		for i := 0; i < procs; i++ {
			if conn, err := this.connectHdfs(); err == nil {
				atomic.AddInt32(&this.rungoroutine, 1)
				this.wg.Add(1)
				go this.asyncollection(conn, this.colltask)
			} else {
				this.Runner.Errorf("Reader hdfs collection err:%v", err)
			}
		}
		if atomic.LoadInt32(&this.rungoroutine) == 0 { // 没有任务
			if this.options.GetAutoClose() { //自动关闭
				event.TriggerEvent(comm.Event_WriteLog, this.Runner.Name(), this.Runner.InstanceId(), "采集协程未能正常启动!", 1, time.Now().Unix())
				go this.AutoClose(core.RunnerFailAutoClose)
			} else {
				atomic.StoreInt32(&this.isend, 0)
			}
		}
	} else {
		this.Runner.Debugf("Reader scan no found table!")
		if this.options.GetAutoClose() { //自动关闭
			event.TriggerEvent(comm.Event_WriteLog, this.Runner.Name(), this.Runner.InstanceId(), "未找到目标数据表!", 1, time.Now().Unix())
			go this.AutoClose(core.RunnerFailAutoClose)
		} else {
			atomic.StoreInt32(&this.isend, 0)
		}
	}

}
func (this *Reader) asyncollection(conn *hdfs.Client, fmeta <-chan *file.FileMeta) {
	defer lego.Recover(fmt.Sprintf("%s Reader", this.Runner.Name()))
	defer atomic.AddInt32(&this.rungoroutine, -1)
	defer conn.Close()
locp:
	for v := range fmeta {
	clocp:
		for {
			if ok, err := this.collection_file(conn, v); ok || err != nil {
				this.Runner.Debugf("Reader asyncollection ok:%v,err:%v", ok, err)
				break clocp
			}
			if this.Runner.GetRunnerState() == core.Runner_Stoping { //采集器进入停止过程中
				break locp
			}
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

//---------------------------------------------------------------------------------------------------------------------------
func (this *Reader) connectHdfs() (conn *hdfs.Client, err error) {

	if this.options.GetHdfs_Kerberos_Enable() {
		var (
			hcf hadoopconf.HadoopConf
			kt  *keytab.Keytab
			cf  *config.Config
		)
		//*从xml导入hadoop配置
		hcf, err = hadoopconf.Load(this.options.GetHdfs_Kerberos_HadoopConfPath())
		if err != nil {
			return
		}
		options := hdfs.ClientOptionsFromConf(hcf)
		//*导入keytab和conf
		kt, err = keytab.Load(this.options.GetHdfs_Kerberos_KeyTabPath())
		if err != nil {
			return
		}
		cf, err = config.Load(this.options.GetHdfs_Kerberos_KerberosConfigPath())
		if err != nil {
			return
		}
		options.KerberosServicePrincipleName = this.options.GetHdfs_Kerberos_ServiceName()
		options.KerberosClient = client.NewWithKeytab(this.options.GetHdfs_Kerberos_Username(), this.options.GetHdfs_Kerberos_Realm(), kt, cf)
		conn, err = hdfs.NewClient(options)
		if err != nil {
			return
		}
		return
	}
	conf := hdfs.ClientOptions{
		Addresses: GetStringList(this.options.GetHdfs_addr()),
		User:      this.options.GetHdfs_user(),
	}
	conn, err = hdfs.NewClient(conf)
	if err != nil {
		return
	}
	return
}

///扫描目录
func (this *Reader) scandirectory(conn *hdfs.Client, dir string, fmetas map[string]*file.FileMeta, cfile map[string]*file.FileMeta) (err error) {
	var (
		entries []fs.FileInfo
		entrie  fs.FileInfo
		matched bool
	)
	if entries, err = conn.ReadDir(dir); err == nil {
		for _, entrie = range entries {
			filepath := dir + "/" + entrie.Name()
			if f, ok := fmetas[filepath]; !ok || f.FileLastCollectionTime < entrie.ModTime().Unix() { //文件有修改
				if !ok {
					f = &file.FileMeta{
						FileName:              filepath,
						FileSize:              uint64(entrie.Size()),
						FileLastModifyTime:    entrie.ModTime().Unix(), //只有采集完毕才同步时间
						FileAlreadyReadOffset: 0,
						FileCacheData:         make([]byte, 0),
					}
					this.meta.SetFile(filepath, f)
				} else {
					f.FileLastModifyTime = entrie.ModTime().Unix()
					f.FileSize = uint64(entrie.Size())
				}
				if entrie.IsDir() {
					err = this.scandirectory(conn, filepath, fmetas, cfile)
				} else if f.FileAlreadyReadOffset < uint64(entrie.Size()) {
					if len(cfile) >= this.options.GetHdfs_max_collection_num() {
						this.Runner.Warnf("Reader scanDirectory up top")
						return
					} else {
						if matched, err = regexp.MatchString(this.options.GetHdfs_regularrules(), entrie.Name()); err == nil && matched {
							cfile[filepath] = f
						} else {
							this.Runner.Errorf("Reader entrie:%s matched:%v err:%v", entrie.Name(), matched, err)
						}
					}
				}
			}
		}
	}
	return
}

//采集文件
func (this *Reader) collection_file(client *hdfs.Client, f *file.FileMeta) (isend bool, err error) {
	var (
		file  *hdfs.FileReader
		buff  []byte
		leng  int
		rbuff *bytes.Buffer
	)
	// defer this.Runner.Debugf("reder hdfs end file:%s", f.FileName)
	// this.Runner.Debugf("reder hdfs start file:%s", f.FileName)
	if file, err = client.Open(f.FileName); err == nil {
		defer file.Close()
		buff = make([]byte, this.options.GetHdfs_read_buffer_size())
		if leng, err = file.ReadAt(buff, int64(f.FileAlreadyReadOffset)+int64(len(f.FileCacheData))); (err == nil || err == io.EOF) && leng > 0 {
			rbuff = bytes.NewBuffer(buff[:leng])
			var (
				line string
				err  error
			)
			for {
				if line, err = rbuff.ReadString('\n'); err != nil {
					// this.Runner.Debugf("reder hdfs file:%s line:%s", f.FileName, line)
					if err == io.EOF {
						if uint64(len(line))+uint64(len(f.FileCacheData))+f.FileAlreadyReadOffset >= f.FileSize { //读完了
							isend = this.writeDataChan(f, line)
						} else {
							this.writeDataCache(f, line)
						}
					}
					break
				} else {
					// this.Runner.Debugf("reder hdfs file:%s line:%s", f.FileName, line)
					isend = this.writeDataChan(f, line)
				}
			}
		} else {
			this.Runner.Errorf("reder hdfs f:%v read err:%v buff:%d leng:%d", f.FileName, err, len(buff), leng)
		}
	} else {
		this.Runner.Errorf("reder hdfs f:%v read err:%v", f.FileName, err)
	}
	return
}

//发送数据 数据量打的时候会有堵塞
func (this *Reader) writeDataChan(f *file.FileMeta, line string) (isend bool) {
	// this.Runner.Debugf("reder ftp file:%s writeDataChan start", f.FileName)
	// defer this.Runner.Debugf("reder ftp file:%s writeDataChan end", f.FileName)
	var value = line
	if f.FileCacheData != nil && len(f.FileCacheData) != 0 {
		var buffer bytes.Buffer
		buffer.Write(f.FileCacheData)
		buffer.WriteString(line)
		value = buffer.String()
		f.FileCacheData = nil
	}
	this.Input() <- core.NewCollData(this.options.GetHdfs_ip(), value)
	f.FileAlreadyReadOffset += uint64(len(value))
	if f.FileAlreadyReadOffset >= f.FileSize {
		f.FileLastCollectionTime = f.FileLastModifyTime
		isend = true
	}
	return
}

func (this *Reader) writeDataCache(f *file.FileMeta, line string) {
	var buffer bytes.Buffer //这种方式对内存优化比较好
	if len(f.FileCacheData) > 0 {
		buffer.Write(f.FileCacheData)
	}
	buffer.WriteString(line)
	f.FileCacheData = buffer.Bytes()
	if uint64(len(f.FileCacheData)) > this.Runner.MaxCollDataSzie() { //大于最大采集数据 直接丢弃
		this.Runner.Errorf("hdfs 清理缓存 超过消息体最大限制！")
		f.FileAlreadyReadOffset += uint64(len(f.FileCacheData))
		f.FileCacheData = make([]byte, 0)
	}
}
func GetStringList(value string) []string {
	v := strings.Split(value, ",")
	var newV []string
	for _, i := range v {
		trimI := strings.TrimSpace(i)
		if len(trimI) > 0 {
			newV = append(newV, trimI)
		}
	}
	return newV
}
