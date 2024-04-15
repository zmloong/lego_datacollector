package mysql

import (
	"fmt"
	"lego_datacollector/modules/datacollector/core"
	"lego_datacollector/modules/datacollector/sender"
	"time"

	"github.com/liwei1dao/lego/sys/log"
	lgsql "github.com/liwei1dao/lego/sys/sql"
)

type Sender struct {
	sender.Sender
	options IOptions
	sql     lgsql.ISys
	keys    []string
}

func (this *Sender) Init(rer core.IRunner, ser core.ISender, options core.ISenderOptions) (err error) {
	this.Sender.Init(rer, ser, options)
	this.options = options.(IOptions)
	for _, v := range this.options.GetMysql_mapping() {
		this.keys = append(this.keys, v)
	}
	this.sql, err = lgsql.NewSys(lgsql.SetSqlType(lgsql.MySql), lgsql.SetSqlUrl(fmt.Sprintf("%s/%s", this.options.GetMysql_datasource(), this.options.GetMysql_database())))
	return
}
func (this *Sender) Start() (err error) {
	err = this.Sender.Start()
	return
}

//关闭
func (this *Sender) Close() (err error) {
	if err = this.Sender.Close(); err != nil {
		log.Errorf("Sender Close err:%v", err)
	}
	err = this.sql.Close()
	return
}

func (this *Sender) Send(pipeId int, bucket core.ICollDataBucket, params ...interface{}) {
	sqlstr := "INSERT INTO `" + this.options.GetMysql_table() + "`("
	datastr := ""

	for _, v := range this.keys {
		sqlstr += "`" + v + "`" + ","
	}
	sqlstr = sqlstr[0 : len(sqlstr)-1]
	sqlstr += ")VALUES"

	//在写入数据
	for _, v := range bucket.SuccItems() {
		data := this.getemptycontainer()
		for k, v := range v.GetData() {
			if v1, ok := this.options.GetMysql_mapping()[k]; ok { //在映射表中提取
				data[v1] = v
			}
		}
		if m, ok := v.GetValue().(map[string]interface{}); ok {
			for k, v := range m {
				if v1, ok := this.options.GetMysql_mapping()[k]; ok { //在映射表中提取
					data[v1] = v
				}
			}
		}
		datastr += this.transIntface(data) + ","
	}
	datastr = datastr[0 : len(datastr)-1]
	// this.Runner.Debugf("sql:%s", sqlstr+datastr)
locp:
	for {
		if _, err := this.sql.Exec(sqlstr + datastr); err != nil {
			this.Runner.Errorf("Send Mysql err:%v", err)
			time.Sleep(time.Millisecond * 100)
			// bucket.SetError(err)
			// return
		} else {
			break locp
		}
		if this.Runner.GetRunnerState() == core.Runner_Stoping { //采集器进入停止过程中
			this.Runner.Debugf("Send asynpipe:%d exit", pipeId)
			break locp
		}
	}
	this.Sender.Send(pipeId, bucket)
	return
}

func (this *Sender) getemptycontainer() (result map[string]interface{}) {
	result = make(map[string]interface{})
	for _, v := range this.options.GetMysql_mapping() {
		result[v] = nil
	}
	return
}

func (this *Sender) transIntface(data map[string]interface{}) (result string) {
	result = "("
	for _, v := range this.keys {
		switch data[v].(type) {
		case string:
			result += fmt.Sprintf("'%s',", data[v])
			break
		case time.Time:
			result += fmt.Sprintf("'%s',", data[v].(time.Time).Format("2006-01-02 15:04:05"))
			break
		default:
			result += fmt.Sprintf("%v,", data[v])
			break
		}
	}
	result = result[0 : len(result)-1]
	result += ")"
	return
}
