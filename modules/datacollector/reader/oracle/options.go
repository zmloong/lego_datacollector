package oracle

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
		GetOracle_collection_type() MyCollectionType
		GetOracle_datasource() string
		GetOracle_database() string
		GetOracle_ip() string
		GetOracle_tables() []string
		GetOracle_sql() string
		GetOracle_batch_intervel() string
		GetOracle_exec_onstart() bool
		GetOracle_limit_batch() int
		GetOracle_offset_key() string
		GetOracle_timestamp_key() string
		GetOracle_startDate() string
		GetOracle_dateUnit() string
		GetOracle_dateGap() int
		GetOracle_datetimeFormat() string
		GetOracle_schema() []string
	}
	Options struct {
		core.ReaderOptions                      //继承基础配置
		Oracle_collection_type MyCollectionType //采集模式
		Oracle_datasource      string           //mysql 连接字段
		Oracle_ip              string           //源ip
		Oracle_database        string           //数据库
		Oracle_tables          []string         //查询表集合
		Oracle_sql             string           //自定义执行语句
		Oracle_limit_batch     int              //每次查询条数
		Oracle_offset_key      string           //递增列(非时间)相关
		Oracle_timestamp_key   string           //递增的时间列相关
		Oracle_startDate       string           //动态表的开始采集时间
		Oracle_datetimeFormat  string           //动态表时间格式
		Oracle_dateUnit        string           //年表 月表 日表
		Oracle_dateGap         int              //动态表 的 是假间隔 默认 1 (几年一表 几月一表 几天一表)
		Oracle_batch_intervel  string           //mysql 扫描间隔采用 cron 定时任务查询 默认 1m 扫描一次
		Oracle_exec_onstart    bool             //mysql 是否启动时查询一次
		Oracle_schema          []string         //类型转换
	}
)

func newOptions(config map[string]interface{}) (opt IOptions, err error) {
	options := &Options{
		Oracle_sql:            `select "$TABLE$".*,ROWNUM as INCREMENTAL From "$TABLE$"`,
		Oracle_offset_key:     "INCREMENTAL",
		Oracle_limit_batch:    100,
		Oracle_batch_intervel: "0 */1 * * * ?",
		Oracle_exec_onstart:   false,
	}
	if config != nil {
		if err = mapstructure.Decode(config, options); err == nil {
			err = mapstructure.Decode(config, &options.ReaderOptions)
		}
	}
	conf := options.GetOracle_datasource()
	addrip := conf[strings.Index(conf, "@")+1 : strings.LastIndex(conf, ":")]
	options.Oracle_ip = addrip
	opt = options
	return
}

func (this *Options) GetOracle_collection_type() MyCollectionType {
	return this.Oracle_collection_type
}

func (this *Options) GetOracle_datasource() string {
	return this.Oracle_datasource
}
func (this *Options) GetOracle_ip() string {
	return this.Oracle_ip
}
func (this *Options) GetOracle_database() string {
	return this.Oracle_database
}
func (this *Options) GetOracle_startDate() string {
	return this.Oracle_startDate
}
func (this *Options) GetOracle_dateUnit() string {
	return this.Oracle_dateUnit
}
func (this *Options) GetOracle_dateGap() int {
	return this.Oracle_dateGap
}
func (this *Options) GetOracle_limit_batch() int {
	return this.Oracle_limit_batch
}

func (this *Options) GetOracle_datetimeFormat() string {
	return this.Oracle_datetimeFormat
}

func (this *Options) GetOracle_tables() []string {
	return this.Oracle_tables
}

func (this *Options) GetOracle_sql() string {
	return this.Oracle_sql
}

func (this *Options) GetOracle_offset_key() string {
	return this.Oracle_offset_key
}
func (this *Options) GetOracle_timestamp_key() string {
	return this.Oracle_timestamp_key
}
func (this *Options) GetOracle_batch_intervel() string {
	return this.Oracle_batch_intervel
}
func (this *Options) GetOracle_exec_onstart() bool {
	return this.Oracle_exec_onstart
}
func (this *Options) GetOracle_schema() []string {
	return this.Oracle_schema
}
