package doris

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
		GetDoris_collection_type() MyCollectionType
		GetDoris_addr() string
		GetDoris_db_name() string
		GetDoris_tables() []string
		GetDoris_sql() string
		GetDoris_limit_batch() int
		GetDoris_timestamp_key() string
		GetDoris_startDate() string
		GetDoris_datetimeFormat() string
		GetDoris_dateUnit() string
		GetDoris_dateGap() int
		GetDoris_batch_intervel() string
		GetDoris_exec_onstart() bool
		GetDoris_schema() []string
	}
	Options struct {
		core.ReaderOptions                     //继承基础配置
		Doris_collection_type MyCollectionType //采集模式
		Doris_addr            string
		Doris_db_name         string
		Doris_tables          []string //查询表集合
		Doris_sql             string   //自定义执行语句
		Doris_limit_batch     int      //每次查询条数
		Doris_timestamp_key   string   //递增的时间列相关
		Doris_startDate       string   //动态表的开始采集时间
		Doris_datetimeFormat  string   //动态表时间格式
		Doris_dateUnit        string   //年表 月表 日表
		Doris_dateGap         int      //动态表 的 是假间隔 默认 1 (几年一表 几月一表 几天一表)
		Doris_batch_intervel  string   //Doris 扫描间隔采用 cron 定时任务查询 默认 1m 扫描一次
		Doris_exec_onstart    bool     //mysql 是否启动时查询一次
		Doris_schema          []string //类型转换
	}
)

func newOptions(config map[string]interface{}) (opt IOptions, err error) {
	options := &Options{
		Doris_sql:            "Select * From `$TABLE$`",
		Doris_limit_batch:    100,
		Doris_batch_intervel: "0 */1 * * * ?",
		Doris_exec_onstart:   false,
	}
	if config != nil {
		if err = mapstructure.Decode(config, options); err == nil {
			err = mapstructure.Decode(config, &options.ReaderOptions)
		}
	}
	opt = options
	return
}
func (this *Options) GetDoris_collection_type() MyCollectionType {
	return this.Doris_collection_type
}
func (this *Options) GetDoris_addr() string {
	return this.Doris_addr
}
func (this *Options) GetDoris_db_name() string {
	return this.Doris_db_name
}
func (this *Options) GetDoris_tables() []string {
	return this.Doris_tables
}
func (this *Options) GetDoris_sql() string {
	return this.Doris_sql
}
func (this *Options) GetDoris_limit_batch() int {
	return this.Doris_limit_batch
}
func (this *Options) GetDoris_timestamp_key() string {
	return this.Doris_timestamp_key
}
func (this *Options) GetDoris_startDate() string {
	return this.Doris_startDate
}
func (this *Options) GetDoris_datetimeFormat() string {
	return this.Doris_datetimeFormat
}
func (this *Options) GetDoris_dateUnit() string {
	return this.Doris_dateUnit
}
func (this *Options) GetDoris_dateGap() int {
	return this.Doris_dateGap
}
func (this *Options) GetDoris_batch_intervel() string {
	return this.Doris_batch_intervel
}
func (this *Options) GetDoris_exec_onstart() bool {
	return this.Doris_exec_onstart
}
func (this *Options) GetDoris_schema() []string {
	return this.Doris_schema
}
