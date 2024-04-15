package xml

import (
	"fmt"
	"io"
	"io/fs"
	"lego_datacollector/modules/datacollector/core"
	"lego_datacollector/modules/datacollector/metaer/file"
	"regexp"
	"sync/atomic"
	"time"

	"github.com/beevik/etree"
	"github.com/liwei1dao/lego"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

//高并发采集服务
func (this *Reader) sftp_collection() {
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
	if conn, err := this.sftp_connectSftp(); err != nil {
		this.Runner.Errorf("Reader Dial sftp err:%v", err)
		return
	} else {
		defer conn.Close()
		fmetas := this.meta.GetFiles()
		this.sftp_scandirectory(conn, this.options.GetXml_directory(), fmetas, cfiles)
	}

	procs := this.Runner.MaxProcs()
	if len(cfiles) < this.Runner.MaxProcs() {
		procs = len(cfiles)
	}
	this.colltask = make(chan *file.FileMeta, len(cfiles))
	for _, v := range cfiles {
		this.colltask <- v
	}

	for i := 0; i < procs; i++ {
		if conn, err := this.sftp_connectSftp(); err == nil {
			atomic.AddInt32(&this.rungoroutine, 1)
			this.wg.Add(1)
			go this.sftp_asyncollection(conn, this.colltask)
		} else {
			this.Runner.Errorf("Reader sftp collection err:%v", err)
		}
	}
}
func (this *Reader) sftp_asyncollection(conn *sftp.Client, fmeta <-chan *file.FileMeta) {
	defer lego.Recover(fmt.Sprintf("%s Reader", this.Runner.Name()))
	defer atomic.AddInt32(&this.rungoroutine, -1)
	defer this.wg.Done()
	defer conn.Close()
locp:
	for v := range fmeta {
	clocp:
		for {
			if ok, err := this.sftp_collection_file(conn, v); ok || err != nil {
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

//---------------------------------------------------------------------------------------------------------------------------
func (this *Reader) sftp_connectSftp() (conn *sftp.Client, err error) {
	var (
		sshClient *ssh.Client
	)
	clientConfig := &ssh.ClientConfig{
		User:            this.options.GetXml_server_user(),
		Auth:            []ssh.AuthMethod{ssh.Password(this.options.GetXml_server_password())},
		Timeout:         10 * time.Second,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	if sshClient, err = ssh.Dial("tcp", fmt.Sprintf("%s:%d", this.options.GetXml_server_addr(), this.options.GetXml_server_port()), clientConfig); err != nil {
		return
	}
	if conn, err = sftp.NewClient(sshClient); err != nil {
		return
	}
	return
}

///扫描目录
func (this *Reader) sftp_scandirectory(conn *sftp.Client, dir string, fmetas map[string]*file.FileMeta, cfile map[string]*file.FileMeta) (err error) {
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
					err = this.sftp_scandirectory(conn, filepath, fmetas, cfile)
				} else if f.FileAlreadyReadOffset < uint64(entrie.Size()) {
					if len(cfile) >= this.options.GetXml_max_collection_num() {
						this.Runner.Debugf("Reader scanDirectory up top")
						return
					} else if entrie.Size() > this.options.GetXml_max_collection_size() {
						this.Runner.Debugf("Reader file:%s size up top", entrie.Name())
					} else {
						if matched, err = regexp.MatchString(this.options.GetXml_regularrules(), entrie.Name()); err == nil && matched {
							cfile[filepath] = f
						} else {
							this.Runner.Debugf("Reader regular:%s file:%s matched:%v err:%v", this.options.GetXml_regularrules(), entrie.Name(), matched, err)
						}
					}
				}
			}
		}
	}
	return
}

//采集文件
func (this *Reader) sftp_collection_file(client *sftp.Client, f *file.FileMeta) (isend bool, err error) {
	var (
		_file *sftp.File
		buff  = make([]byte, this.options.GetXml_max_collection_size())
		doc   *etree.Document
		n     int
	)
	if _file, err = client.Open(f.FileName); err == nil {
		defer _file.Close()
		if n, err = _file.ReadAt(buff, 0); err != nil && err != io.EOF {
			return
		}
	}
	if n > 0 {
		doc = etree.NewDocument()
		if err = doc.ReadFromBytes(buff[0:n]); err != nil {
			return
		}
		root := doc.FindElement(this.options.GetXml_node_root())
		if root != nil {
			this.Runner.Debugf("ROOT element:", root.Tag)
			for _, book := range root.SelectElements(this.options.GetXml_node_data()) {
				this.Runner.Debugf("CHILD element:", book.Tag)
				data := make(map[string]string)
				for _, v := range book.ChildElements() {
					data[v.Tag] = v.Text()
				}
				this.Input() <- core.NewCollData(this.options.GetXml_server_addr(), data)
			}
		} else {
			this.Runner.Errorf("ROOT element Is Null")
		}

	}
	f.FileAlreadyReadOffset = f.FileSize
	return true, nil
}
