package binary

import (
	"fmt"
	"io"
	"lego_datacollector/comm"
	"lego_datacollector/modules/datacollector/core"
	"lego_datacollector/modules/datacollector/metaer/file"
	"regexp"
	"sync/atomic"
	"time"

	"github.com/jlaffaye/ftp"
	"github.com/liwei1dao/lego"
	"github.com/liwei1dao/lego/sys/event"
)

//高并发采集服务
func (this *Reader) ftp_collection() {
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
	if conn, err := this.ftp_connectFtp(); err != nil {
		this.Runner.Errorf("Reader Dial Ftp err:%v", err)
		if this.options.GetAutoClose() { //自动关闭
			event.TriggerEvent(comm.Event_WriteLog, this.Runner.Name(), this.Runner.InstanceId(), fmt.Sprintf("连接目标服务器失败:%v", err), 1, time.Now().Unix())
			go this.AutoClose(core.RunnerFailAutoClose)
		} else {
			atomic.StoreInt32(&this.isend, 0)
		}
		return
	} else {
		fmetas := this.meta.GetFiles()
		this.ftp_scandirectory(conn, this.options.GetBinary_directory(), fmetas, cfiles)
		conn.Quit()
	}
	this.Runner.Debugf("Reader collection filenum:%d", len(cfiles))
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
			if conn, err := this.ftp_connectFtp(); err == nil {
				atomic.AddInt32(&this.rungoroutine, 1)
				this.wg.Add(1)
				go this.ftp_asyncollection(conn, this.colltask)
			} else {
				this.Runner.Errorf("Reader Ftp collection err:%v", err)
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
			event.TriggerEvent(comm.Event_WriteLog, this.Runner.Name(), this.Runner.InstanceId(), "未找到目标数据文件!", 1, time.Now().Unix())
			go this.AutoClose(core.RunnerFailAutoClose)
		} else {
			atomic.StoreInt32(&this.isend, 0)
		}
	}
}

func (this *Reader) ftp_asyncollection(conn *ftp.ServerConn, fmeta <-chan *file.FileMeta) {
	defer lego.Recover(fmt.Sprintf("%s Reader", this.Runner.Name()))
	defer atomic.AddInt32(&this.rungoroutine, -1)
	defer conn.Quit()
locp:
	for v := range fmeta {
	clocp:
		for {
			if ok, err := this.ftp_collection_file(conn, v); ok || err != nil {
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
	this.Runner.Debugf("Reader asyncollection exit")
}

//业务接口------------------------------------------------------------------------------------------------------------------
func (this *Reader) ftp_connectFtp() (conn *ftp.ServerConn, err error) {
	if conn, err = ftp.Dial(fmt.Sprintf("%s:%d", this.options.GetBinary_server_addr(), this.options.GetBinary_server_port()), ftp.DialWithTimeout(5*time.Second)); err == nil {
		err = conn.Login(this.options.GetBinary_server_user(), this.options.GetBinary_server_password())
	}
	return
}

///扫描目录
func (this *Reader) ftp_scandirectory(conn *ftp.ServerConn, dir string, fmetas map[string]*file.FileMeta, cfile map[string]*file.FileMeta) (err error) {
	var (
		entries []*ftp.Entry
		entrie  *ftp.Entry
		matched bool
	)
	if entries, err = conn.List(dir); err == nil {
		for _, entrie = range entries {
			filepath := dir + "/" + entrie.Name
			if f, ok := fmetas[filepath]; !ok || f.FileLastCollectionTime < entrie.Time.Unix() { //文件有修改
				if !ok {
					f = &file.FileMeta{
						FileName:              filepath,
						FileSize:              entrie.Size,
						FileLastModifyTime:    entrie.Time.Unix(), //只有采集完毕才同步时间
						FileAlreadyReadOffset: 0,
						FileCacheData:         make([]byte, 0),
					}
					this.meta.SetFile(filepath, f)
				} else {
					f.FileLastModifyTime = entrie.Time.Unix()
					f.FileSize = entrie.Size
				}
				if entrie.Type == ftp.EntryTypeFolder {
					err = this.ftp_scandirectory(conn, filepath, fmetas, cfile)
				} else if entrie.Type == ftp.EntryTypeFile && f.FileAlreadyReadOffset < entrie.Size {
					if len(cfile) >= this.options.GetBinary_max_collection_num() {
						this.Runner.Debugf("Reader scanDirectory up top")
						return
					} else {
						if matched, err = regexp.MatchString(this.options.GetBinary_regularrules(), entrie.Name); err == nil && matched {
							cfile[filepath] = f
						} else {
							this.Runner.Debugf("Reader regular:%s file:%s matched:%v err:%v", this.options.GetBinary_regularrules(), entrie.Name, matched, err)
						}
					}
				}
			}
		}
	}
	return
}

//采集文件
func (this *Reader) ftp_collection_file(conn *ftp.ServerConn, f *file.FileMeta) (isend bool, err error) {
	var (
		resp *ftp.Response
		buff = make([]byte, this.options.GetBinary_readerbuf_size())
		n    int
	)
	if resp, err = conn.RetrFrom(f.FileName, f.FileAlreadyReadOffset+uint64(len(f.FileCacheData))); err == nil || err == io.EOF {
		defer resp.Close()
		if n, err = resp.Read(buff); err != nil && err != io.EOF {
			return
		}
	}
	if n > 0 {
		if err = this.parse(f, buff[:n]); err == nil {
			f.FileAlreadyReadOffset += uint64(n)
		} else {
			return
		}
	}
	if f.FileAlreadyReadOffset >= f.FileSize {
		return true, nil
	} else {
		return false, nil
	}
}
