package xls

import (
	"fmt"
	"io"
	"io/fs"
	"io/ioutil"
	"lego_datacollector/modules/datacollector/core"
	"lego_datacollector/modules/datacollector/metaer/file"
	"path"
	"regexp"
	"sync/atomic"

	"github.com/extrame/xls"
	"github.com/liwei1dao/lego"
)

func (this *Reader) local_collection() {
	this.lock.Lock()
	defer this.lock.Unlock()
	var (
		files           = make(map[string]fs.FileInfo)
		collectionfiles = make([]*file.FileMeta, 0)
		f               *file.FileMeta
		ok              bool
	)
	if rungoroutine := atomic.LoadInt32(&this.rungoroutine); rungoroutine != 0 {
		this.Runner.Debugf("Reader is collectioning rungoroutine:%d", rungoroutine)
		this.SyncMeta()
		return
	}
	if err := this.local_scanDirectory(this.options.GetXls_directory(), files); err == nil {
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
				go this.local_asyncollection(this.colltask)
			}
		}
	} else {
		this.Runner.Infof("Reader folder Scan Directory err:%v", err)
	}
}
func (this *Reader) local_asyncollection(fmeta <-chan *file.FileMeta) {
	defer lego.Recover(fmt.Sprintf("%s Reader", this.Runner.Name()))
	defer atomic.AddInt32(&this.rungoroutine, -1)
	defer this.wg.Done()
	var (
		f      *xls.WorkBook
		closer io.Closer
		err    error
	)
locp:
	for v := range fmeta {
		if f, closer, err = xls.OpenWithCloser(v.FileName, "utf-8"); err == nil {
		clocp:
			for {
				if ok, err := this.local_collection_file(v, f); ok || err != nil {
					this.Runner.Debugf("Reader asyncollection ok:%v,err:%v", ok, err)
					closer.Close()
					break clocp
				}
				if this.Runner.GetRunnerState() == core.Runner_Stoping { //采集器进入停止过程中
					this.Runner.Debugf("Reader asyncollection exit")
					closer.Close()
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
	if atomic.LoadInt32(&this.rungoroutine) == 1 { //最后一个任务已经完成
		this.collectionend()
	}
	this.Runner.Debugf("Reader asyncollection exit succ!")
}

//--------------------------------------------------------------------------------------------------------
func (this *Reader) local_scanDirectory(_path string, files map[string]fs.FileInfo) (err error) {
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
				err = this.local_scanDirectory(fp, files)
			} else {
				if len(files) >= this.options.GetXls_max_collection_num() {
					this.Runner.Debugf("Reader scanDirectory up top")
					return
				} else if v.Size() > this.options.GetXls_max_collection_size() {
					this.Runner.Debugf("Reader file:%s size：%d up top:%d", v.Name(), v.Size(), this.options.GetXls_max_collection_size())
				} else {
					if matched, err = regexp.MatchString(this.options.GetXls_regularrules(), v.Name()); err == nil && matched {
						files[fp] = v
					} else {
						this.Runner.Debugf("Reader regular:%s file:%s matched:%v err:%v", this.options.GetXls_regularrules(), v.Name(), matched, err)
					}
				}
			}

		}
	}
	return
}

///采集文件
func (this *Reader) local_collection_file(fmeta *file.FileMeta, f *xls.WorkBook) (iscollend bool, err error) {
	var (
		sheet *xls.WorkSheet
		keys  *xls.Row
		key   string
	)

	for i := 0; i < f.NumSheets(); i++ {
		if sheet = f.GetSheet(i); sheet.MaxRow == 0 {
			this.Runner.Errorf("reader sheet GetSheet:%d MaxRow:%v", i, sheet.MaxRow)
			continue
		}
		this.Runner.Debugf("reader xls sheet GetSheet:%d MaxRow:%v", i, sheet.MaxRow)
		if int(sheet.MaxRow) > this.options.GetXls_key_line() {
			keys = sheet.Row(this.options.GetXls_key_line())
			if keys.LastCol() == 0 {
				this.Runner.Errorf("reader key_line:%d last_col:%d", this.options.GetXls_key_line(), int(keys.LastCol()))
				continue
			}
		} else {
			this.Runner.Errorf("reader key_line:%d rows_counr:%d", this.options.GetXls_key_line(), int(sheet.MaxRow))
			continue
		}
		for n := 0; n < int(sheet.MaxRow)+1; n++ {
			if n >= this.options.GetXls_data_line() {
				row := sheet.Row(n)
				data := make(map[string]string)
				if row.LastCol() == keys.LastCol() {
					for j := 0; j < row.LastCol(); j++ {
						key = keys.Col(j)
						data[key] = row.Col(j)
					}
					this.Runner.Debugf("xls data:%v", data)
					this.Input() <- core.NewCollData(this.Runner.ServiceIP(), data)
				} else {
					this.Runner.Errorf("reader xls keys_count:%d row_count:%d", keys.LastCol(), row.LastCol())
				}
			}
		}
	}
	fmeta.FileAlreadyReadOffset = fmeta.FileSize
	return true, nil
}
