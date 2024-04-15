package dm

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
		GetDM_collection_type() MyCollectionType
		GetDM_datasource() string
		GetDM_database() string
		GetDM_tables() []string
		GetDM_sql() string
		GetDM_batch_intervel() string
		GetDM_exec_onstart() bool
		GetDM_limit_batch() int
		GetDM_offset_key() string
		GetDM_timestamp_key() string
		GetDM_startDate() string
		GetDM_dateUnit() string
		GetDM_dateGap() int
		GetDM_datetimeFormat() string
		GetDM_schema() []string
	}
	Options struct {
		core.ReaderOptions                  //继承基础配置
		DM_collection_type MyCollectionType //采集模式
		DM_datasource      string           //mysql 连接字段
		DM_database        string           //数据库
		DM_tables          []string         //查询表集合
		DM_sql             string           //自定义执行语句
		DM_limit_batch     int              //每次查询条数
		DM_offset_key      string           //递增列(非时间)相关
		DM_timestamp_key   string           //递增的时间列相关
		DM_startDate       string           //动态表的开始采集时间
		DM_datetimeFormat  string           //动态表时间格式
		DM_dateUnit        string           //年表 月表 日表
		DM_dateGap         int              //动态表 的 是假间隔 默认 1 (几年一表 几月一表 几天一表)
		DM_batch_intervel  string           //mysql 扫描间隔采用 cron 定时任务查询 默认 1m 扫描一次
		DM_exec_onstart    bool             //mysql 是否启动时查询一次
		DM_schema          []string         //类型转换
	}
)

func newOptions(config map[string]interface{}) (opt IOptions, err error) {
	options := &Options{
		DM_sql:            "select * From $TABLE$",
		DM_limit_batch:    100,
		DM_batch_intervel: "0 */1 * * * ?",
		DM_exec_onstart:   false,
	}
	if config != nil {
		if err = mapstructure.Decode(config, options); err == nil {
			err = mapstructure.Decode(config, &options.ReaderOptions)
		}
	}
	opt = options
	return
}

func (this *Options) GetDM_collection_type() MyCollectionType {
	return this.DM_collection_type
}

func (this *Options) GetDM_datasource() string {
	return this.DM_datasource
}
func (this *Options) GetDM_database() string {
	return this.DM_database
}
func (this *Options) GetDM_startDate() string {
	return this.DM_startDate
}
func (this *Options) GetDM_dateUnit() string {
	return this.DM_dateUnit
}
func (this *Options) GetDM_dateGap() int {
	return this.DM_dateGap
}
func (this *Options) GetDM_limit_batch() int {
	return this.DM_limit_batch
}

func (this *Options) GetDM_datetimeFormat() string {
	return this.DM_datetimeFormat
}

func (this *Options) GetDM_tables() []string {
	return this.DM_tables
}

func (this *Options) GetDM_sql() string {
	return this.DM_sql
}

func (this *Options) GetDM_offset_key() string {
	return this.DM_offset_key
}
func (this *Options) GetDM_timestamp_key() string {
	return this.DM_timestamp_key
}
func (this *Options) GetDM_batch_intervel() string {
	return this.DM_batch_intervel
}
func (this *Options) GetDM_exec_onstart() bool {
	return this.DM_exec_onstart
}
func (this *Options) GetDM_schema() []string {
	return this.DM_schema
}
