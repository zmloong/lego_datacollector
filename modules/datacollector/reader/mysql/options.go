package mysql

import (
	"lego_datacollector/modules/datacollector/core"
	"strings"

	"github.com/liwei1dao/lego/utils/mapstructure"
)

const (
	StaticTableCollection  MyCollectionType = iota //静态表采集
	DynamicTableCollection                         //动态表采集
)

type (
	MyCollectionType uint8
	IOptions         interface {
		core.IReaderOptions //继承基础配置
		GetMysql_collection_type() MyCollectionType
		GetMysql_datasource() string
		GetMysql_ip() string
		GetMysql_database() string
		GetMysql_tables() []string
		GetMysql_sql() string
		GetMysql_batch_intervel() string
		GetMysql_exec_onstart() bool
		GetMysql_limit_batch() int
		GetMysql_timestamp_key() string
		GetMysql_startDate() string
		GetMysql_dateUnit() string
		GetMysql_dateGap() int
		GetMysql_datetimeFormat() string
		GetMysql_schema() []string
	}
	Options struct {
		core.ReaderOptions                     //继承基础配置
		Mysql_collection_type MyCollectionType //采集模式
		Mysql_datasource      string           //mysql 连接字段
		Mysql_ip              string           //mysql 连接字段
		Mysql_database        string           //数据库
		Mysql_tables          []string         //查询表集合
		Mysql_sql             string           //自定义执行语句
		Mysql_limit_batch     int              //每次查询条数
		Mysql_timestamp_key   string           //递增的时间列相关
		Mysql_startDate       string           //动态表的开始采集时间
		Mysql_datetimeFormat  string           //动态表时间格式
		Mysql_dateUnit        string           //年表 月表 日表
		Mysql_dateGap         int              //动态表 的 是假间隔 默认 1 (几年一表 几月一表 几天一表)
		Mysql_batch_intervel  string           //mysql 扫描间隔采用 cron 定时任务查询 默认 1m 扫描一次
		Mysql_exec_onstart    bool             //mysql 是否启动时查询一次
		Mysql_schema          []string         //类型转换
	}
)

func newOptions(config map[string]interface{}) (opt IOptions, err error) {
	options := &Options{
		Mysql_sql:            "Select * From `$TABLE$`",
		Mysql_limit_batch:    100,
		Mysql_batch_intervel: "0 */1 * * * ?",
		Mysql_exec_onstart:   false,
	}
	if config != nil {
		if err = mapstructure.Decode(config, options); err == nil {
			err = mapstructure.Decode(config, &options.ReaderOptions)
		}
	}
	conf := options.GetMysql_datasource()
	addrip := conf[strings.Index(conf, "(")+1 : strings.LastIndex(conf, ":")]
	options.Mysql_ip = addrip
	opt = options
	return
}

func (this *Options) GetMysql_collection_type() MyCollectionType {
	return this.Mysql_collection_type
}

func (this *Options) GetMysql_datasource() string {
	return this.Mysql_datasource
}
func (this *Options) GetMysql_ip() string {
	return this.Mysql_ip
}
func (this *Options) GetMysql_database() string {
	return this.Mysql_database
}
func (this *Options) GetMysql_startDate() string {
	return this.Mysql_startDate
}
func (this *Options) GetMysql_dateUnit() string {
	return this.Mysql_dateUnit
}
func (this *Options) GetMysql_dateGap() int {
	return this.Mysql_dateGap
}
func (this *Options) GetMysql_limit_batch() int {
	return this.Mysql_limit_batch
}

func (this *Options) GetMysql_datetimeFormat() string {
	return this.Mysql_datetimeFormat
}

func (this *Options) GetMysql_tables() []string {
	return this.Mysql_tables
}

func (this *Options) GetMysql_sql() string {
	return this.Mysql_sql
}

// func (this *Options) GetMysql_offset_key() string {
// 	return this.Mysql_offset_key
// }
func (this *Options) GetMysql_timestamp_key() string {
	return this.Mysql_timestamp_key
}
func (this *Options) GetMysql_batch_intervel() string {
	return this.Mysql_batch_intervel
}
func (this *Options) GetMysql_exec_onstart() bool {
	return this.Mysql_exec_onstart
}
func (this *Options) GetMysql_schema() []string {
	return this.Mysql_schema
}
