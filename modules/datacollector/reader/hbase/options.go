package hbase

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
		GetHbase_collection_type() MyCollectionType
		GetHbase_host() string
		GetHbase_tables() []string
		GetHbase_startDate() string
		GetHbase_dateUnit() string
		GetHbase_dateGap() int
		GetHbase_datetimeFormat() string
		GetHbase_schema() []string
		GetHbase_exec_onstart() bool
		GetHbase_batch_intervel() string
	}
	Options struct {
		core.ReaderOptions                     //继承基础配置
		Hbase_collection_type MyCollectionType //采集模式
		Hbase_host            string           //postgre 连接地址
		Hbase_tables          []string         //查询表集合
		Hbase_startDate       string           //动态表的开始采集时间
		Hbase_datetimeFormat  string           //动态表时间格式
		Hbase_dateUnit        string           //年表 月表 日表
		Hbase_dateGap         int              //动态表 的 是假间隔 默认 1 (几年一表 几月一表 几天一表)
		Hbase_batch_intervel  string           //postgre 扫描间隔采用 cron 定时任务查询 默认 1m 扫描一次
		Hbase_exec_onstart    bool             //postgre 是否启动时查询一次
		Hbase_schema          []string         //类型转换
	}
)

func newOptions(config map[string]interface{}) (opt IOptions, err error) {
	options := &Options{
		Hbase_batch_intervel: "0 */1 * * * ?",
		Hbase_exec_onstart:   false,
	}
	if config != nil {
		if err = mapstructure.Decode(config, options); err == nil {
			err = mapstructure.Decode(config, &options.ReaderOptions)
		}
	}
	opt = options
	return
}

func (this *Options) GetHbase_collection_type() MyCollectionType {
	return this.Hbase_collection_type
}
func (this *Options) GetHbase_host() string {
	return this.Hbase_host
}
func (this *Options) GetHbase_tables() []string {
	return this.Hbase_tables
}
func (this *Options) GetHbase_startDate() string {
	return this.Hbase_startDate
}
func (this *Options) GetHbase_dateUnit() string {
	return this.Hbase_dateUnit
}
func (this *Options) GetHbase_dateGap() int {
	return this.Hbase_dateGap
}
func (this *Options) GetHbase_datetimeFormat() string {
	return this.Hbase_datetimeFormat
}
func (this *Options) GetHbase_batch_intervel() string {
	return this.Hbase_batch_intervel
}
func (this *Options) GetHbase_exec_onstart() bool {
	return this.Hbase_exec_onstart
}
func (this *Options) GetHbase_schema() []string {
	return this.Hbase_schema
}
