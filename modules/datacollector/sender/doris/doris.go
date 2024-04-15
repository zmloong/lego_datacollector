package doris

import (
	"bytes"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"lego_datacollector/modules/datacollector/core"
	"lego_datacollector/modules/datacollector/sender"
	"net/http"
	"time"

	lgsql "github.com/liwei1dao/lego/sys/sql"
	"github.com/liwei1dao/lego/utils/container/id"
)

type (
	//Stream load返回消息结构体
	ResponseBody struct {
		TxnID                  int    `json:"TxnId"`
		Label                  string `json:"Label"`
		Status                 string `json:"Status"`
		Message                string `json:"Message"`
		NumberTotalRows        int    `json:"NumberTotalRows"`
		NumberLoadedRows       int    `json:"NumberLoadedRows"`
		NumberFilteredRows     int    `json:"NumberFilteredRows"`
		NumberUnselectedRows   int    `json:"NumberUnselectedRows"`
		LoadBytes              int    `json:"LoadBytes"`
		LoadTimeMs             int    `json:"LoadTimeMs"`
		BeginTxnTimeMs         int    `json:"BeginTxnTimeMs"`
		StreamLoadPutTimeMs    int    `json:"StreamLoadPutTimeMs"`
		ReadDataTimeMs         int    `json:"ReadDataTimeMs"`
		WriteDataTimeMs        int    `json:"WriteDataTimeMs"`
		CommitAndPublishTimeMs int    `json:"CommitAndPublishTimeMs"`
		ErrorURL               string `json:"ErrorURL"`
	}
	Sender struct {
		sender.Sender
		options IOptions
		column  []string
	}
)

//Sender----------------------------------------------------------------------------------------------------------------
func (this *Sender) Type() string {
	return SenderType
}

func (this *Sender) Init(rer core.IRunner, ser core.ISender, options core.ISenderOptions) (err error) {
	this.Sender.Init(rer, ser, options)
	this.options = options.(IOptions)
	if this.options.GetDoris_sendertype() == CollectionSender { //采集输出类型 需要建表
		err = this.create_table()
	} else { //自定义表格 需要获取到表列名
		err = this.gettablecolumn()
	}
	return
}

func (this *Sender) Send(pipeId int, bucket core.ICollDataBucket, params ...interface{}) {
	buffer := new(bytes.Buffer)
	for _, v := range bucket.SuccItems() {
		if this.options.GetDoris_sendertype() == CollectionSender {
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
			buffer.WriteString(fmt.Sprintf("%s%s%s%s%s%s%s%s%s\n", timestr, this.options.GetDoris_column_separator(), collect_ip, this.options.GetDoris_column_separator(), eqpt_ip, this.options.GetDoris_column_separator(), expmatc, this.options.GetDoris_column_separator(), message))
		} else {
			str := ""
			for _, v1 := range this.column {
				found := false
				for k2, v2 := range this.options.GetDoris_mapping() {
					if v1 == v2 {
						found = true
						if v3, ok := v.GetData()[k2]; ok {
							str += fmt.Sprintf("%v%s", v3, this.options.GetDoris_column_separator())
						} else if m, ok := v.GetValue().(map[string]interface{}); ok {
							if v2, ok := m[k2]; ok {
								str += fmt.Sprintf("%v%s", v2, this.options.GetDoris_column_separator())
							} else {
								str += this.options.GetDoris_column_separator()
							}
						} else {
							str += this.options.GetDoris_column_separator()
						}
					}
				}
				if !found {
					str += this.options.GetDoris_column_separator()
				}
			}
			str = str[0 : len(str)-1]
			str += "\n"
			buffer.WriteString(str)
		}
	}
	// this.Runner.Debugf("Sender data:%s", buffer.String())
	this.http_doris_write(buffer)
	this.Sender.Send(pipeId, bucket)
}

///创建采集表
func (this *Sender) create_table() (err error) {
	var (
		sql lgsql.ISys
	)
	if sql, err = lgsql.NewSys(lgsql.SetSqlType(lgsql.MySql), lgsql.SetSqlUrl(fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", this.options.GetDoris_user(), this.options.GetDoris_password(), this.options.GetDoris_ip(), this.options.GetDoris_sql_port(), this.options.GetDoris_dbname()))); err != nil {
		return
	}
	sqlstr := fmt.Sprintf("CREATE TABLE IF NOT EXISTS doris.%s \n"+
		"(\n"+
		"	`idss_collect_time` DATETIME COMMENT \"idss_collect_time\","+
		"	`idss_collect_ip` VARCHAR(20) COMMENT \"idss_collect_ip\","+
		"	`eqpt_ip` VARCHAR(20) COMMENT \"eqpt_ip\","+
		"	`idss_collect_expmatch` BOOLEAN COMMENT \"idss_collect_expmatch\","+
		"	`message` VARCHAR(1000) COMMENT \"message\""+
		")\n"+
		"ENGINE=olap\n"+
		"DUPLICATE KEY(`idss_collect_time`)\n"+
		"PARTITION BY RANGE(`idss_collect_time`) ()\n"+
		"DISTRIBUTED BY HASH(`idss_collect_time`) BUCKETS 16\n"+
		"PROPERTIES\n"+
		"(\n"+
		"    \"replication_num\" = \"3\",\n"+
		"    \"storage_medium\" = \"SSD\",\n"+
		"    \"dynamic_partition.enable\" = \"true\",\n"+
		"    \"dynamic_partition.time_unit\" = \"DAY\",\n"+
		"    \"dynamic_partition.end\" = \"3\",\n"+
		"    \"dynamic_partition.prefix\" = \"p\",\n"+
		"    \"dynamic_partition.buckets\" = \"16\"\n"+
		");", this.options.GetDoris_tableName())
	if _, err = sql.Exec(sqlstr); err != nil {
		this.Runner.Errorf("Sender Doris create table err:%v", err)
	}
	return
}

///获取表列名
func (this *Sender) gettablecolumn() (err error) {
	var db lgsql.ISys
	var result *sql.Rows
	db, err = lgsql.NewSys(
		lgsql.SetSqlType(lgsql.MySql),
		lgsql.SetSqlUrl(fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", this.options.GetDoris_user(), this.options.GetDoris_password(), this.options.GetDoris_ip(), this.options.GetDoris_sql_port(), this.options.GetDoris_dbname())),
	)
	if err != nil {
		this.Runner.Errorf("Sender Doris gettablecolumn err:%v", err)
		return
	} else {
		defer db.Close()
		this.column = make([]string, 0)
		sqlstr := fmt.Sprintf("select COLUMN_NAME from information_schema.COLUMNS where table_name = '%s';", this.options.GetDoris_tableName())
		if result, err = db.Query(sqlstr); err == nil {
			for result.Next() {
				column := ""
				if err = result.Scan(&column); err != nil {
					this.Runner.Errorf("Sender Doris gettablecolumn err:%v", err)
					return
				} else {
					this.column = append(this.column, column)
				}
			}
		} else {
			this.Runner.Errorf("Sender Doris gettablecolumn err:%v", err)
			return
		}
	}
	return
}

//发送数据
func (this *Sender) http_doris_write(body io.Reader) {
	client := &http.Client{}
	reqest, err := http.NewRequest(http.MethodPut, fmt.Sprintf("http://%s:%d/api/%s/%s/_stream_load", this.options.GetDoris_ip(), this.options.GetDoris_stream_port(), this.options.GetDoris_dbname(), this.options.GetDoris_tableName()), body)
	//增加header选项
	reqest.Header.Add("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(this.options.GetDoris_user()+":"+this.options.GetDoris_password())))
	reqest.Header.Add("EXPECT", "100-continue")
	reqest.Header.Add("label", id.NewXId())
	reqest.Header.Add("column_separator", this.options.GetDoris_column_separator())
	//处理返回结果
	response, err := client.Do(reqest)
	if err == nil {
		defer response.Body.Close()
		body, _ := ioutil.ReadAll(response.Body)
		responseBody := ResponseBody{}
		err := json.Unmarshal(body, &responseBody)
		if err != nil {
			this.Runner.Errorf("Sender Doris body:%s err:%v", body, err)
		}
		if responseBody.Status != "Success" {
			this.Runner.Errorf("Sender Doris err:%v", err)
		}
	} else {
		this.Runner.Errorf("Sender Doris err:%v", err)
	}
}
