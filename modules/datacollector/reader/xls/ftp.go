package xls

import (
	"bytes"
	"fmt"
	"io"
	"lego_datacollector/modules/datacollector/core"
	"lego_datacollector/modules/datacollector/metaer/file"
	"regexp"
	"sync/atomic"
	"time"

	"github.com/jlaffaye/ftp"
	"github.com/liwei1dao/lego"
	"github.com/xuri/excelize/v2"
)

//高并发采集服务
func (this *Reader) ftp_collection() {
	this.lock.Lock()
	defer this.lock.Unlock()
	var (
		cfiles map[string]*file.FileMeta
	)
	if rungoroutine := atomic.LoadInt32(&this.rungoroutine); rungoroutine != 0 {
		this.Runner.Debugf("Reader is collectioning rungoroutine:%d", rungoroutine)
		this.SyncMeta()
		return
	}
	cfiles = make(map[string]*file.FileMeta)
	if conn, err := this.ftp_connectFtp(); err != nil {
		this.Runner.Errorf("Reader Dial Ftp err:%v", err)
		return
	} else {
		fmetas := this.meta.GetFiles()
		this.ftp_scandirectory(conn, this.options.GetXls_directory(), fmetas, cfiles)
		conn.Quit()
	}
	this.Runner.Debugf("Reader collection filenum:%d", len(cfiles))
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
}

func (this *Reader) ftp_asyncollection(conn *ftp.ServerConn, fmeta <-chan *file.FileMeta) {
	defer lego.Recover(fmt.Sprintf("%s Reader", this.Runner.Name()))
	defer atomic.AddInt32(&this.rungoroutine, -1)
	defer this.wg.Done()
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
	if atomic.LoadInt32(&this.rungoroutine) == 1 { //最后一个任务已经完成
		this.collectionend()
	}
	this.Runner.Debugf("Reader asyncollection exit")
}

//业务接口------------------------------------------------------------------------------------------------------------------
func (this *Reader) ftp_connectFtp() (conn *ftp.ServerConn, err error) {
	if conn, err = ftp.Dial(fmt.Sprintf("%s:%d", this.options.GetXls_server_addr(), this.options.GetXls_server_port()), ftp.DialWithTimeout(5*time.Second)); err == nil {
		err = conn.Login(this.options.GetXls_server_user(), this.options.GetXls_server_password())
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
					if len(cfile) >= this.options.GetXls_max_collection_num() {
						this.Runner.Debugf("Reader scanDirectory up top")
						return
					} else if int64(entrie.Size) > this.options.GetXls_max_collection_size() {
						this.Runner.Debugf("Reader file:%s size up top", entrie.Name)
					} else {
						if matched, err = regexp.MatchString(this.options.GetXls_regularrules(), entrie.Name); err == nil && matched {
							cfile[filepath] = f
						} else {
							this.Runner.Debugf("Reader regular:%s file:%s matched:%v err:%v", this.options.GetXls_regularrules(), entrie.Name, matched, err)
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
		buff = make([]byte, this.options.GetXls_max_collection_size())
		xf   *excelize.File
		rows [][]string
		keys []string
		key  string
		n    int
	)
	if resp, err = conn.RetrFrom(f.FileName, 0); err == nil {
		defer resp.Close()
		if n, err = resp.Read(buff); err != nil && err != io.EOF {
			return
		}
	}
	if n > 0 {
		reader := bytes.NewReader(buff[0:n])
		if xf, err = excelize.OpenReader(reader); err == nil {
			for _, v := range xf.GetSheetList() {
				if rows, err = xf.GetRows(v); err != nil {
					this.Runner.Errorf("reader GetRows:%s err:%v", v, err)
					continue
				}
				if len(rows) > this.options.GetXls_key_line() {
					keys = rows[this.options.GetXls_key_line()]
				} else {
					this.Runner.Errorf("reader key_line:%d rows_counr:%d", this.options.GetXls_key_line(), len(rows))
					continue
				}
				for i, row := range rows {
					if i >= this.options.GetXls_data_line() {
						if len(keys) == len(row) {
							data := make(map[string]string)
							for ii, vv := range row {
								key = keys[ii]
								data[key] = vv
							}
							this.Runner.Debugf("xls data:%v", data)
							this.Input() <- core.NewCollData(this.Runner.ServiceIP(), data)
						} else {
							this.Runner.Errorf("reader xls keys_count:%d row_count:%d", len(keys), len(row))
						}
					}
				}
			}
		}
	}
	f.FileAlreadyReadOffset = f.FileSize
	return true, nil
}
