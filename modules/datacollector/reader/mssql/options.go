package mssql

import (
	"lego_datacollector/modules/datacollector/core"

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
		GetMssql_collection_type() MyCollectionType
		GetMssql_datasource() string
		GetMssql_database() string
		GetMssql_tables() []string
		GetMssql_sql() string
		GetMssql_batch_intervel() string
		GetMssql_exec_onstart() bool
		GetMssql_limit_batch() int
		GetMssql_offset_key() string
		GetMssql_timestamp_key() string
		GetMssql_startDate() string
		GetMssql_dateUnit() string
		GetMssql_dateGap() int
		GetMssql_datetimeFormat() string
		GetMssql_schema() []string
	}
	Options struct {
		core.ReaderOptions                     //继承基础配置
		Mssql_collection_type MyCollectionType //采集模式
		Mssql_datasource      string           //mssql 连接字段
		Mssql_database        string           //数据库
		Mssql_tables          []string         //查询表集合
		Mssql_sql             string           //自定义执行语句
		Mssql_limit_batch     int              //每次查询条数
		Mssql_offset_key      string           //递增列(非时间)相关
		Mssql_timestamp_key   string           //递增的时间列相关
		Mssql_startDate       string           //动态表的开始采集时间
		Mssql_datetimeFormat  string           //动态表时间格式
		Mssql_dateUnit        string           //年表 月表 日表
		Mssql_dateGap         int              //动态表 的 是假间隔 默认 1 (几年一表 几月一表 几天一表)
		Mssql_batch_intervel  string           //mssql 扫描间隔采用 cron 定时任务查询 默认 1m 扫描一次
		Mssql_exec_onstart    bool             //mssql 是否启动时查询一次
		Mssql_schema          []string         //类型转换
	}
)

func newOptions(config map[string]interface{}) (opt IOptions, err error) {
	options := &Options{
		Mssql_sql:            "select row_number() over (order by (select 0)) as INCREMENTAL,* from $TABLE$",
		Mssql_offset_key:     "INCREMENTAL",
		Mssql_limit_batch:    100,
		Mssql_batch_intervel: "0 */1 * * * ?",
		Mssql_exec_onstart:   false,
	}
	if config != nil {
		if err = mapstructure.Decode(config, options); err == nil {
			err = mapstructure.Decode(config, &options.ReaderOptions)
		}
	}
	opt = options
	return
}

func (this *Options) GetMssql_collection_type() MyCollectionType {
	return this.Mssql_collection_type
}

func (this *Options) GetMssql_datasource() string {
	return this.Mssql_datasource
}
func (this *Options) GetMssql_database() string {
	return this.Mssql_database
}
func (this *Options) GetMssql_startDate() string {
	return this.Mssql_startDate
}
func (this *Options) GetMssql_dateUnit() string {
	return this.Mssql_dateUnit
}
func (this *Options) GetMssql_dateGap() int {
	return this.Mssql_dateGap
}
func (this *Options) GetMssql_limit_batch() int {
	return this.Mssql_limit_batch
}

func (this *Options) GetMssql_datetimeFormat() string {
	return this.Mssql_datetimeFormat
}

func (this *Options) GetMssql_tables() []string {
	return this.Mssql_tables
}

func (this *Options) GetMssql_sql() string {
	return this.Mssql_sql
}
func (this *Options) GetMssql_offset_key() string {
	return this.Mssql_offset_key
}
func (this *Options) GetMssql_timestamp_key() string {
	return this.Mssql_timestamp_key
}
func (this *Options) GetMssql_batch_intervel() string {
	return this.Mssql_batch_intervel
}
func (this *Options) GetMssql_exec_onstart() bool {
	return this.Mssql_exec_onstart
}
func (this *Options) GetMssql_schema() []string {
	return this.Mssql_schema
}
