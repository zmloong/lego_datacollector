package hive

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"lego_datacollector/modules/datacollector/core"
	"lego_datacollector/modules/datacollector/sender"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"
	"unsafe"

	"github.com/beltran/gohive"
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
	connHive *gohive.Connection
	column   []string
}
type fsCM struct {
	id              string
	client          *hdfs.Client
	connHive        *gohive.Connection
	addr            string
	username        string
	currentFilePath string
	currentSize     uint64
	hasLoadHive     bool //标记文件已上传
}

func (this *Sender) Init(rer core.IRunner, ser core.ISender, options core.ISenderOptions) (err error) {
	this.Sender.Init(rer, ser, options)
	this.options = options.(IOptions)
	this.fsc = make(map[int]*fsCM)
	err = this.getConnectMap(this.options.GetHdfs_addr(), this.options.GetHdfs_user(), this.Runner.MaxProcs())
	if err != nil {
		this.Runner.Debugf("Sender Init get HDFS Connect err:%v", err)
	}
	this.runname = rer.Name()
	this.connHive, err = this.getConnectHive()
	if err != nil {
		this.Runner.Debugf("Sender Init get HIVE Connect err:%v", err)
	}
	if this.options.GetHive_sendertype() == CollectionSender { //采集输出类型 需要建表
		err = this.create_table()
	} else { //自定义表格 需要获取到表列名
		err = this.getHiveSqlColumns()
	}
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
		var conn_hive *gohive.Connection
		conn_hive, err = this.getConnectHive()
		if err != nil {
			return
		}
		fsc.id = fmt.Sprintf("%d", i)
		fsc.client = client
		fsc.connHive = conn_hive
		fsc.addr = addr
		fsc.username = username
		this.fsc[i] = &fsc
	}
	return
}
func (this *Sender) writeHdfs(pipeId int, databytes []byte) (err error) {
	fsc := this.fsc[pipeId+1]
	if fsc.currentFilePath == "" || fsc.currentSize >= uint64(this.options.GetHdfs_size()) {
		//第一次创建文件
		for {
			if err := this.createFile(fsc); err != nil && this.Runner.GetRunnerState() == core.Runner_Runing {
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
				//*标记文件写满
				fsc.currentSize = uint64(this.options.GetHdfs_size()) + 1
				//*报错后文件不为空导入hive
				if fsc.currentSize != 0 && !fsc.hasLoadHive {
					sqlstr := "LOAD DATA INPATH '" + fsc.currentFilePath + "' INTO TABLE " + this.options.GetHive_dbname() + "." + this.options.GetHive_tableName()
					if err := fsc.sqlHive(sqlstr); err != nil {
						this.Runner.Errorf("hive load  err:%v", err)
					}
					fsc.hasLoadHive = true
				}
				return
			}
			time.Sleep(time.Second)
			continue
		}
		//*文件写满导入hive
		if fsc.currentSize >= uint64(this.options.GetHdfs_size()) && !fsc.hasLoadHive {
			sqlstr := "LOAD DATA INPATH '" + fsc.currentFilePath + "' INTO TABLE " + this.options.GetHive_dbname() + "." + this.options.GetHive_tableName()
			if err := fsc.sqlHive(sqlstr); err != nil {
				this.Runner.Errorf("hive load hdfs err:%v", err)
			}
			fsc.hasLoadHive = true
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
		err = v.connHive.Close()
	}
	this.connHive.Close()
	return
}
func (this *Sender) Send(pipeId int, bucket core.ICollDataBucket, params ...interface{}) {
	buffer := new(bytes.Buffer)
	for _, v := range bucket.SuccItems() {
		if this.options.GetHive_sendertype() == CollectionSender {
			timestr := time.UnixMilli(v.GetData()[core.KeyTimestamp].(int64)).Format("2006-01-02 15:04:05")
			collect_ip := ""
			if v, ok := v.GetData()[core.KeyIdsCollecip]; ok {
				collect_ip = v.(string)
			}
			eqpt_ip := ""
			if v, ok := v.GetData()[core.KeyEqptip]; ok {
				eqpt_ip = v.(string)
			}
			expmatc := ""
			if v, ok := v.GetData()[core.KeyIdsCollecexpmatch]; ok {
				if v.(bool) {
					expmatc = "1"
				} else {
					expmatc = "0"
				}
			}
			message := v.GetData()[core.KeyRaw]
			messagebt, err := json.Marshal(message)

			if err != nil {
				this.Runner.Errorf("hive CollectionSender message map to json err:%v", err)
			}
			// if collect_ip == "" {
			// 	collect_ip = "NULL"
			// }
			// if eqpt_ip == "" {
			// 	eqpt_ip = "NULL"
			// }
			// if expmatc == "" {
			// 	expmatc = "NULL"
			// }
			if this.options.GetHive_json() {
				jsonmap := make(map[string]interface{}, 0)
				jsonmap["idss_collect_time"] = timestr
				jsonmap["idss_collect_ip"] = collect_ip
				jsonmap["eqpt_ip"] = eqpt_ip
				jsonmap["idss_collect_expmatch"] = expmatc
				jsonmap["message"] = message

				jsonbt, err := json.Marshal(jsonmap)
				if err != nil {
					this.Runner.Errorf("hive CollectionSender Hive_json map to json err:%v", err)
				}
				buffer.WriteString(fmt.Sprintf("%s\n", jsonbt))
			} else {
				buffer.WriteString(fmt.Sprintf("%s%s%s%s%s%s%s%s%s\n", timestr, this.options.GetHive_column_separator(), collect_ip, this.options.GetHive_column_separator(), eqpt_ip, this.options.GetHive_column_separator(), expmatc, this.options.GetHive_column_separator(), messagebt))
			}

		} else {
			str := ""
			for _, v1 := range this.column {
				found := false
				for k2, v2 := range this.options.GetHive_mapping() {
					if v1 == v2 {
						found = true
						if v3, ok := v.GetData()[k2]; ok {
							str += fmt.Sprintf("%v%s", v3, this.options.GetHive_column_separator())
						} else if m, ok := v.GetValue().(map[string]interface{}); ok {
							if v2, ok := m[k2]; ok {
								str += fmt.Sprintf("%v%s", v2, this.options.GetHive_column_separator())
							} else {
								str += this.options.GetHive_column_separator()
							}
						} else {
							str += this.options.GetHive_column_separator()
						}
					}
				}
				if !found {
					str += this.options.GetHive_column_separator()
				}
			}
			str = str[0 : len(str)-1]
			str += "\n"
			buffer.WriteString(str)
		}
	}

	err := this.writeHdfs(pipeId, buffer.Bytes())

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
func (this *Sender) getConnectHive() (connection *gohive.Connection, err error) {
	configuration := gohive.NewConnectConfiguration()
	configuration.Username = this.options.GetHive_username()
	configuration.Password = this.options.GetHive_password()
	hopt := strings.Split(this.options.GetHive_addr(), ":")
	port, _ := strconv.Atoi(hopt[1])
	connection, errConn := gohive.Connect(hopt[0], port, this.options.GetHive_auth(), configuration)
	if errConn != nil {
		return
	}
	return
}
func (this *Sender) createFile(fsc *fsCM) (err error) {
	filetail := ".log"
	if this.options.GetHive_json() {
		filetail = ".json"
	}
	tn := fmt.Sprintf("%v", time.Now().UnixMilli())
	path0 := this.options.GetHdfs_path() + "/" + time.Now().Format("2006-01-02") + "/" + this.runname + "/" + `%Y%m%d%H%M%S` + "_" + tn + "_p" + fsc.id + filetail
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
	fsc.hasLoadHive = false
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
func (this *Sender) create_table() (err error) {
	sqlstr1 := fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s", this.options.GetHive_dbname())
	err = this.sqlState(this.connHive, sqlstr1)
	if err != nil {
		return
	}
	var sqltable string
	if this.options.GetHive_json() {
		sqltable = fmt.Sprintf("CREATE external TABLE IF NOT EXISTS datacollector.%s (jsonObject string) Location '"+this.options.GetHdfs_path()+"/hive'", this.options.GetHive_tableName())
	} else {
		sqltable = fmt.Sprintf("CREATE TABLE IF NOT EXISTS datacollector.%s "+"("+"idss_collect_time STRING,"+"idss_collect_ip STRING,"+"eqpt_ip STRING,"+"idss_collect_expmatch STRING,"+"message STRING"+")"+"ROW FORMAT DELIMITED FIELDS TERMINATED BY '%s' STORED AS TEXTFILE", this.options.GetHive_tableName(), this.options.GetHive_column_separator())
	}

	err = this.sqlState(this.connHive, sqltable)
	if err != nil {
		return
	}
	return
}
func (this *Sender) getHiveSqlColumns() (err error) {

	// connHive := getconnectTest()
	// this.connHive, err = this.getConnectHive()
	// if err != nil {
	// 	this.Runner.Debugf("Sender Init get HIVE Connect err:%v", err)
	// }
	cursor := this.connHive.Cursor()
	ctx := context.Background()
	var sqlstr string
	if this.options.GetHive_dbname() != "" {
		sqlstr = "desc " + this.options.GetHive_dbname() + "." + this.options.GetHive_tableName()
	} else {
		sqlstr = "desc " + this.options.GetHive_tableName()
	}

	cursor.Exec(ctx, sqlstr)
	// cursor.WaitForCompletion(ctx)
	// cursor.Finished()
	//*验证状态
	if cursor.Err != nil {
		err = cursor.Err
		return
	}
	//*获取返回值
	for cursor.HasMore(ctx) {
		if cursor.Err != nil {
			err = cursor.Err
			return
		}
		m := cursor.RowMap(ctx)
		this.column = append(this.column, m["col_name"].(string))
	}
	cursor.Close()
	//connection.Close()
	return
}
func (this *Sender) sqlState(connection *gohive.Connection, sqlstr string) (err error) {
	cursor := connection.Cursor()
	ctx := context.Background()

	this.Runner.Debugf("sqlState start: [ %v ]", sqlstr)
	cursor.Exec(ctx, sqlstr)
	//*验证状态
	if cursor.Err != nil {
		err = cursor.Err
		return
	}
	// cursor.WaitForCompletion(ctx)
	// stat = cursor.Poll(true)
	cursor.Close()
	//connection.Close()
	fmt.Printf("sqlState end Close!")
	return
}
func (this *fsCM) sqlHive(sqlstr string) (err error) {
	cursor := this.connHive.Cursor()
	ctx := context.Background()
	fmt.Printf("sqlHive start: [ %v ]", sqlstr)
	cursor.Exec(ctx, sqlstr)
	if cursor.Err != nil {
		cursor.WaitForCompletion(ctx)
		stat := cursor.Poll(false)
		fmt.Println(stat.Status)
		err = cursor.Err
		return
	}
	fmt.Printf("sqlHive over: [ %v ]", sqlstr)
	cursor.WaitForCompletion(ctx)
	stat := cursor.Poll(true)
	fmt.Println(stat.Status)

	cursor.Close()
	// connection.Close()
	fmt.Printf("sqlHive end Close!")
	return
}
