package hdfs

/*
import (
	"bytes"
	"encoding/json"
	"fmt"
	"lego_datacollector/modules/datacollector/core"
	"lego_datacollector/modules/datacollector/sender"
	"time"

	"github.com/lestrrat-go/strftime"
	"github.com/robfig/cron/v3"
	"github.com/vladimirvivien/gowfs"
)

type Sender struct {
	sender.Sender
	options         IOptions
	runId           cron.EntryID
	pattern         *strftime.Strftime
	fs              *gowfs.FileSystem
	runname         string
	currentFilePath string
}

func (this *Sender) Init(rer core.IRunner, options core.ISenderOptions) (err error) {
	this.Sender.Init(rer, options)
	this.options = options.(IOptions)
	this.fs, err = getConnect(this.options.GetHdfs_addr(), this.options.GetHdfs_user(), this.options.GetHdfs_timeout())
	if err != nil {
		this.Runner.Debugf("Sender Init getConnect:", err)
	}
	this.runname = rer.Name()
	path := this.options.GetHdfs_path() + "/" + "_" + "/" + rer.Name() + "/" + `%Y%m%d%H%M%S.txt`
	if this.pattern, err = strftime.New(path); err != nil {
		return
	}
	return
}
func getConnect(addr string, username string, timeout int) (fs *gowfs.FileSystem, err error) {
	conf := *gowfs.NewConfiguration()
	conf.Addr = addr
	conf.User = username
	conf.ConnectionTimeout = time.Second * time.Duration(timeout)
	conf.DisableKeepAlives = false
	fs, err = gowfs.NewFileSystem(conf)
	if err != nil {
		return
	}
	return
}

//关闭
func (this *Sender) Close() (err error) {
	this.fs = &gowfs.FileSystem{}
	return
}

func (this *Sender) Send(bucket core.ICollDataBucket) {
	var (
		databytes []byte
		err       error
	)

	if databytes, err = json.Marshal(bucket); err != nil {
		bucket.SetError(err)
		return
	}
	if this.currentFilePath == "" {
		if err = this.createFile(databytes); err != nil {
			bucket.SetError(err)
			return
		}
	} else {
		size, err := this.getFileStatus()
		//134217728
		if size >= 800 {
			if err = this.createFile(databytes); err != nil {
				bucket.SetError(err)
				return
			}
		} else {
			if err = this.appendFile(databytes); err != nil {
				bucket.SetError(err)
			}
		}

	}
	return
}
func (this *Sender) createFile(data []byte) (err error) {
	path0 := this.options.GetHdfs_path() + "/" + time.Now().Format("2006-01-02") + "/" + this.runname + "/" + `%Y%m%d%H%M%S.log`
	if this.pattern, err = strftime.New(path0); err != nil {
		return
	}
	filename := this.pattern.FormatString(time.Now())
	path := gowfs.Path{Name: filename}
	ok, err := this.fs.Create(bytes.NewBuffer(data), path,
		false,
		0,
		0,
		0755,
		0,
		"application/octet-stream",
	)
	if err != nil || !ok {
		return fmt.Errorf("hdfs createFile err: %v", err)
	}
	this.currentFilePath = filename
	return
}
func (this *Sender) appendFile(data []byte) (err error) {
	path := gowfs.Path{Name: this.currentFilePath}
	ok, err := this.fs.Append(bytes.NewBuffer(data), path,
		0,
		"application/octet-stream",
	)
	if err != nil || !ok {
		return fmt.Errorf("hdfs appendFile err: %v", err)
	}
	return
}
func (this *Sender) getFileStatus() (size int64, err error) {
	path := gowfs.Path{Name: this.currentFilePath}
	gfs, err := this.fs.GetFileStatus(path)
	if err != nil {
		this.Runner.Debugf("getFileStatus err:", err)
	}
	size = gfs.Length
	return
}
*/
