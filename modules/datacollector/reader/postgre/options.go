package postgre

import (
	"lego_datacollector/modules/datacollector/core"
	"strconv"
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
		GetPostgre_collection_type() MyCollectionType
		GetPostgre_addr() string
		GetPostgre_ip() string
		GetPostgre_port() int
		GetPostgre_user() string
		GetPostgre_password() string
		GetPostgre_database() string
		GetPostgre_tables() []string
		GetPostgre_sql() string
		GetPostgre_batch_intervel() string
		GetPostgre_exec_onstart() bool
		GetPostgre_limit_batch() int
		GetPostgre_offset_key() string
		GetPostgre_timestamp_key() string
		GetPostgre_startDate() string
		GetPostgre_dateUnit() string
		GetPostgre_dateGap() int
		GetPostgre_datetimeFormat() string
		GetPostgre_schema() []string
	}
	Options struct {
		core.ReaderOptions                       //继承基础配置
		Postgre_collection_type MyCollectionType //采集模式
		Postgre_addr            string           //postgre 连接地址
		Postgre_ip              string           //ip
		Postgre_port            int              //端口
		Postgre_user            string           //用户名
		Postgre_password        string           //密码
		Postgre_database        string           //数据库
		Postgre_tables          []string         //查询表集合
		Postgre_sql             string           //自定义执行语句
		Postgre_limit_batch     int              //每次查询条数
		Postgre_offset_key      string           //递增列(非时间)相关
		Postgre_timestamp_key   string           //递增的时间列相关
		Postgre_startDate       string           //动态表的开始采集时间
		Postgre_datetimeFormat  string           //动态表时间格式
		Postgre_dateUnit        string           //年表 月表 日表
		Postgre_dateGap         int              //动态表 的 是假间隔 默认 1 (几年一表 几月一表 几天一表)
		Postgre_batch_intervel  string           //postgre 扫描间隔采用 cron 定时任务查询 默认 1m 扫描一次
		Postgre_exec_onstart    bool             //postgre 是否启动时查询一次
		Postgre_schema          []string         //类型转换
	}
)

func newOptions(config map[string]interface{}) (opt IOptions, err error) {
	options := &Options{
		Postgre_sql:            "select row_number() over () as INCREMENTAL,* from $TABLE$",
		Postgre_offset_key:     "INCREMENTAL",
		Postgre_limit_batch:    100,
		Postgre_batch_intervel: "0 */1 * * * ?",
		Postgre_exec_onstart:   false,
	}
	if config != nil {
		if err = mapstructure.Decode(config, options); err == nil {
			err = mapstructure.Decode(config, &options.ReaderOptions)
		}
	}
	conf := options.GetPostgre_addr()
	options.Postgre_ip = strings.Split(conf, ":")[0]
	options.Postgre_port, _ = strconv.Atoi(strings.Split(conf, ":")[1])
	opt = options
	return
}

func (this *Options) GetPostgre_collection_type() MyCollectionType {
	return this.Postgre_collection_type
}

func (this *Options) GetPostgre_addr() string {
	return this.Postgre_addr
}
func (this *Options) GetPostgre_ip() string {
	return this.Postgre_ip
}
func (this *Options) GetPostgre_port() int {
	return this.Postgre_port
}
func (this *Options) GetPostgre_user() string {
	return this.Postgre_user
}
func (this *Options) GetPostgre_password() string {
	return this.Postgre_password
}
func (this *Options) GetPostgre_database() string {
	return this.Postgre_database
}
func (this *Options) GetPostgre_startDate() string {
	return this.Postgre_startDate
}
func (this *Options) GetPostgre_dateUnit() string {
	return this.Postgre_dateUnit
}
func (this *Options) GetPostgre_dateGap() int {
	return this.Postgre_dateGap
}
func (this *Options) GetPostgre_limit_batch() int {
	return this.Postgre_limit_batch
}
func (this *Options) GetPostgre_datetimeFormat() string {
	return this.Postgre_datetimeFormat
}
func (this *Options) GetPostgre_tables() []string {
	return this.Postgre_tables
}
func (this *Options) GetPostgre_sql() string {
	return this.Postgre_sql
}
func (this *Options) GetPostgre_offset_key() string {
	return this.Postgre_offset_key
}
func (this *Options) GetPostgre_timestamp_key() string {
	return this.Postgre_timestamp_key
}
func (this *Options) GetPostgre_batch_intervel() string {
	return this.Postgre_batch_intervel
}
func (this *Options) GetPostgre_exec_onstart() bool {
	return this.Postgre_exec_onstart
}
func (this *Options) GetPostgre_schema() []string {
	return this.Postgre_schema
}
