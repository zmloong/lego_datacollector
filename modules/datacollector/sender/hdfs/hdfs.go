package hdfs

import (
	"fmt"
	"lego_datacollector/modules/datacollector/core"
	"lego_datacollector/modules/datacollector/sender"
	"reflect"
	"strings"
	"sync"
	"time"
	"unsafe"

	"github.com/colinmarc/hdfs/v2"
	"github.com/colinmarc/hdfs/v2/hadoopconf"
	"github.com/jcmturner/gokrb5/v8/client"
	"github.com/jcmturner/gokrb5/v8/config"
	"github.com/jcmturner/gokrb5/v8/keytab"
	"github.com/lestrrat-go/strftime"
	"github.com/robfig/cron/v3"
)

type Sender struct {
	sender.Sender
	options  IOptions
	runId    cron.EntryID
	pattern  *strftime.Strftime
	fsc      map[int]*fsCM
	runname  string
	dataChan chan []byte
	wg       *sync.WaitGroup
}
type fsCM struct {
	id              string
	client          *hdfs.Client
	addr            string
	username        string
	currentFilePath string
	currentSize     uint64
}

func (this *Sender) Init(rer core.IRunner, ser core.ISender, options core.ISenderOptions) (err error) {
	this.Sender.Init(rer, ser, options)
	this.options = options.(IOptions)
	this.fsc = make(map[int]*fsCM)
	err = this.getConnectMap(this.options.GetHdfs_addr(), this.options.GetHdfs_user(), this.Runner.MaxProcs())
	if err != nil {
		this.Runner.Debugf("Sender Init getConnect:", err)
	}
	this.runname = rer.Name()
	return
}
func (this *Sender) getConnectMap(addr string, username string, num int) (err error) {
	for i := 1; i <= num; i++ {
		var fsc fsCM
		var client *hdfs.Client
		client, err = this.getConnect(addr, username)
		if err != nil {
			return
		}
		fsc.id = fmt.Sprintf("%d", i)
		fsc.client = client
		fsc.addr = addr
		fsc.username = username
		this.fsc[i] = &fsc
	}
	return
}
func (this *Sender) writeHdfs(pipeId int, databytes []byte) (err error) {

	fsc := this.fsc[pipeId+1]
	size := fsc.currentSize
	if fsc.currentFilePath == "" || size >= uint64(this.options.GetHdfs_size()) {
		//第一次创建文件
		for {
			if err := this.createFile(fsc, databytes); err != nil && this.Runner.GetRunnerState() == core.Runner_Runing {
				this.Runner.Debugf("hdfs createFile:", err)
				if strings.Contains(err.Error(), "org.apache.hadoop.fs.FileAlreadyExistsException") {
					this.Runner.Errorf("hdfs createFile err fs.FileAlreadyExistsException break！")
					break
				}
				time.Sleep(time.Second)
				continue
			}
			break
		}

	}

	for {

		if err = this.appendFile(fsc, databytes); err != nil && this.Runner.GetRunnerState() == core.Runner_Runing {

			if strings.Contains(err.Error(), "org.apache.hadoop.hdfs.protocol.AlreadyBeingCreatedException") {
				this.Runner.Debugf("hdfs appendFile :", err)
				//标记文件写满
				fsc.currentSize = uint64(this.options.GetHdfs_size()) + 1
				return
			}
			time.Sleep(time.Second)
			continue
		}
		return
	}
}

//关闭
func (this *Sender) Close() (err error) {
	err = this.Sender.Close()
	if err != nil {
		this.Runner.Errorf("hdfs Sender.Close err :%v", err)
	}
	for _, v := range this.fsc {
		err = v.client.Close()
	}

	return
}

func (this *Sender) Send(pipeId int, bucket core.ICollDataBucket, params ...interface{}) {
	var (
		databytes []byte
		datastr   string
		err       error
	)
	if datastr, err = bucket.ToString(true); err != nil {
		bucket.SetError(err)
		return
	}
	datastr = datastr + "\n"
	databytes = String2Bytes(datastr)
	err = this.writeHdfs(pipeId, databytes)

	if err != nil && strings.Contains(err.Error(), "org.apache.hadoop.hdfs.protocol.AlreadyBeingCreatedException") {
		this.Runner.Debugf("AlreadyBeingCreatedException pipeid:%v", pipeId)
		go this.Runner.Push_SenderPipe(bucket)
		return
	}
	this.Sender.Send(pipeId, bucket)
	return
}
func (this *fsCM) fsCMconnect(s *Sender) (err error) {
	var client *hdfs.Client
	client, err = s.getConnect(this.addr, this.username)
	if err != nil {
		return
	}
	this.client = client
	return
}
func (this *Sender) getConnect(addr string, username string) (conn *hdfs.Client, err error) {
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
		Addresses: GetStringList(addr),
		User:      username,
	}
	conn, err = hdfs.NewClient(conf)
	if err != nil {
		return
	}
	return
}
func (this *Sender) createFile(fsc *fsCM, data []byte) (err error) {
	tn := fmt.Sprintf("%v", time.Now().UnixMilli())
	path0 := this.options.GetHdfs_path() + "/" + time.Now().Format("2006-01-02") + "/" + this.runname + "/" + `%Y%m%d%H%M%S` + "_" + tn + "_p" + fsc.id + `.log`
	if this.pattern, err = strftime.New(path0); err != nil {
		return
	}
	filename := this.pattern.FormatString(time.Now())
	err = this.createDir(fsc, filename)
	if err != nil {
		this.Runner.Errorf("createDir err:%v", err)
		return
	}
	fs, err := fsc.client.CreateFile(filename,
		1,
		134217728,
		0755,
	)
	if err != nil {
		this.Runner.Errorf("createFile err:%v", err)
		// fs.Close()
		return
	}
	// _, err = fs.Write(data)
	// if err != nil {
	// 	this.Runner.Errorf("createFile Write err:%v", err)
	// 	return
	// }
	err = fs.Close()
	if err != nil {
		return
	}
	fsc.currentFilePath = filename
	fsc.currentSize = 0
	err = fsc.fsCMconnect(this)
	if err != nil {
		return
	}
	return
}
func (this *Sender) appendFile(fsc *fsCM, data []byte) (err error) {
	fs, err := fsc.client.Append(fsc.currentFilePath)
	if err != nil {
		//this.Runner.Errorf("appendFile err:%v", err)
		// fs.Close()
		return
	}
	num, err := fs.Write(data)
	if err != nil {
		this.Runner.Errorf("appendFile Write %v err:%v", num, err)
		// fs.Close()
		return
	}
	fsc.currentSize += uint64(len(data))
	err = fs.Close()
	if err != nil {
		return
	}
	return
}
func (this *Sender) getFileStatus(fsc *fsCM) (size int64, err error) {
	fsinfo, err := fsc.client.Stat(fsc.currentFilePath)
	if err != nil {
		this.Runner.Errorf("getFileStatus err:%v", err)
		return
	}
	size = fsinfo.Size()
	return
}

func (this *Sender) createDir(fs *fsCM, hdfsPath string) (err error) {
	//去掉路径最后文件名 xxx.log
	index := strings.LastIndex(hdfsPath, "/")
	err = fs.client.MkdirAll(hdfsPath[:index], 0755)
	if err != nil {
		this.Runner.Errorf("createDir path:%s,err:%v", hdfsPath, err)
		return
	}
	return
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
func String2Bytes(s string) []byte {
	sh := (*reflect.StringHeader)(unsafe.Pointer(&s))
	bh := reflect.SliceHeader{
		Data: sh.Data,
		Len:  sh.Len,
		Cap:  sh.Len,
	}
	return *(*[]byte)(unsafe.Pointer(&bh))
}

func Bytes2String(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}
