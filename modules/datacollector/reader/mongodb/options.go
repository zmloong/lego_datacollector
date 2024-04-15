package mongodb

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
		GetMongodb_collection_type() MyCollectionType
		GetMongodb_datasource() string
		GetMongodb_database() string
		GetMongodb_tables() []string
		GetMongodb_batch_intervel() string
		GetMongodb_exec_onstart() bool
		GetMongodb_limit_batch() int
		GetMongodb_timestamp_key() string
		GetMongodb_startDate() string
		GetMongodb_dateUnit() string
		GetMongodb_dateGap() int
		GetMongodb_datetimeFormat() string
	}
	Options struct {
		core.ReaderOptions                       //继承基础配置
		Mongodb_collection_type MyCollectionType //采集模式
		Mongodb_datasource      string           //Mongodb 连接字段
		Mongodb_database        string           //数据库
		Mongodb_tables          []string         //查询表集合
		Mongodb_limit_batch     int              //每次查询条数
		Mongodb_timestamp_key   string           //递增的时间列相关
		Mongodb_startDate       string           //动态表的开始采集时间
		Mongodb_datetimeFormat  string           //动态表时间格式
		Mongodb_dateUnit        string           //年表 月表 日表
		Mongodb_dateGap         int              //动态表 的 是假间隔 默认 1 (几年一表 几月一表 几天一表)
		Mongodb_batch_intervel  string           //Mongodb 扫描间隔采用 cron 定时任务查询 默认 1m 扫描一次
		Mongodb_exec_onstart    bool             //Mongodb 是否启动时查询一次
	}
)

func newOptions(config map[string]interface{}) (opt IOptions, err error) {
	options := &Options{
		Mongodb_limit_batch:    100,
		Mongodb_batch_intervel: "0 */1 * * * ?",
		Mongodb_exec_onstart:   false,
	}
	if config != nil {
		if err = mapstructure.Decode(config, options); err == nil {
			err = mapstructure.Decode(config, &options.ReaderOptions)
		}
	}
	opt = options
	return
}

func (this *Options) GetMongodb_collection_type() MyCollectionType {
	return this.Mongodb_collection_type
}

func (this *Options) GetMongodb_datasource() string {
	return this.Mongodb_datasource
}
func (this *Options) GetMongodb_database() string {
	return this.Mongodb_database
}
func (this *Options) GetMongodb_startDate() string {
	return this.Mongodb_startDate
}
func (this *Options) GetMongodb_dateUnit() string {
	return this.Mongodb_dateUnit
}
func (this *Options) GetMongodb_dateGap() int {
	return this.Mongodb_dateGap
}
func (this *Options) GetMongodb_limit_batch() int {
	return this.Mongodb_limit_batch
}

func (this *Options) GetMongodb_datetimeFormat() string {
	return this.Mongodb_datetimeFormat
}

func (this *Options) GetMongodb_tables() []string {
	return this.Mongodb_tables
}
func (this *Options) GetMongodb_timestamp_key() string {
	return this.Mongodb_timestamp_key
}
func (this *Options) GetMongodb_batch_intervel() string {
	return this.Mongodb_batch_intervel
}
func (this *Options) GetMongodb_exec_onstart() bool {
	return this.Mongodb_exec_onstart
}
