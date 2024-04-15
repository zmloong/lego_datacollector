package binary

import (
	"fmt"
	"io"
	"io/fs"
	"io/ioutil"
	"lego_datacollector/modules/datacollector/core"
	"lego_datacollector/modules/datacollector/metaer/file"
	"os"
	"path"
	"regexp"
	"sync/atomic"

	"github.com/liwei1dao/lego"
)

func (this *Reader) local_collection() {

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
	if err := this.local_scanDirectory(this.options.GetBinary_directory(), files); err == nil {
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
		} else {
			this.Runner.Debugf("Reader scan no found new data!")
			atomic.StoreInt32(&this.isend, 0)
		}
	} else {
		this.Runner.Infof("Reader folder Scan Directory err:%v", err)
		atomic.StoreInt32(&this.isend, 0)
	}
}
func (this *Reader) local_asyncollection(fmeta <-chan *file.FileMeta) {
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
				// this.Runner.Debugf("Reader collection_file flile:%v", v.FileName)
				if ok, err := this.local_collection_file(v, f); ok || err != nil {
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
				if len(files) >= this.options.GetBinary_max_collection_num() {
					this.Runner.Debugf("Reader scanDirectory up top")
					return
				} else {
					if matched, err = regexp.MatchString(this.options.GetBinary_regularrules(), v.Name()); err == nil && matched {
						files[fp] = v
					} else {
						this.Runner.Debugf("Reader regular:%s file:%s matched:%v err:%v", this.options.GetBinary_regularrules(), v.Name(), matched, err)
					}
				}
			}

		}
	}
	return
}

///采集文件
func (this *Reader) local_collection_file(fmeta *file.FileMeta, f *os.File) (iscollend bool, err error) {
	var (
		buff = make([]byte, this.options.GetBinary_readerbuf_size())
		n    int
	)
	if n, err = f.ReadAt(buff, int64(fmeta.FileAlreadyReadOffset)); err != nil && err != io.EOF {
		return
	}
	if n > 0 {
		if err = this.parse(fmeta, buff[:n]); err == nil {
			fmeta.FileAlreadyReadOffset += uint64(n)
		} else {
			return
		}
	}
	if fmeta.FileAlreadyReadOffset >= fmeta.FileSize {
		return true, nil
	} else {
		return false, nil
	}
}
