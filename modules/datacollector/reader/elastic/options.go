package elastic

import (
	"fmt"
	"lego_datacollector/modules/datacollector/core"
	"strings"

	"github.com/liwei1dao/lego/utils/mapstructure"
)

const (
	StaticTableCollection  MyCollectionType = iota //静态表采集（全量）
	DynamicTableCollection                         //动态表采集
)
const (
	TimestampTypeEsDate = 1 //ES存储时间格式
	TimestampTypeUnix   = 2 //时间戳
)

// const (
// 	Taskselfcontinue = 0 //增量 不关闭
// 	Taskselfclose    = 1 //全量自动关闭
// )

type (
	MyCollectionType uint8
	IOptions         interface {
		core.IReaderOptions //继承基础配置
		GetElastic_collection_type() MyCollectionType
		//GetElastic_taskselfclose() int
		GetElastic_addr() string
		GetElastic_ip() string
		// GetElastic_database() string
		GetElastic_tables() []string
		// GetElastic_sql() string
		GetElastic_batch_intervel() string
		GetElastic_exec_onstart() bool
		GetElastic_limit_batch() int
		GetElastic_query_starttime() string
		GetElastic_query_endtime() string
		GetQuery_starttime() int64
		GetQuery_endtime() int64
		GetElastic_offset_key() string
		GetElastic_timestamp_type() int
		GetElastic_timestamp_key() string
		GetElastic_startDate() string
		GetElastic_dateUnit() string
		GetElastic_dateGap() int
		GetElastic_datetimeFormat() string
		GetElastic_schema() []string
	}
	Options struct {
		core.ReaderOptions                       //继承基础配置
		Elastic_collection_type MyCollectionType //采集模式
		//Elastic_taskselfclose   int              //是否采集完自动关闭任务： 0 不关闭-增量 1关闭-全量
		Elastic_addr string //elastic 连接字段
		Elastic_ip   string //elastic ip
		//Elastic_database        string           //数据库
		Elastic_tables []string //查询索引 index集合
		//Elastic_sql             string           //自定义执行语句
		Elastic_limit_batch     int      //每次查询条数
		Elastic_query_starttime string   //搜索开始时间
		Elastic_query_endtime   string   //搜索结束时间
		Query_starttime         int64    //搜索开始时间戳
		Query_endtime           int64    //搜索结束时间戳
		Elastic_offset_key      string   //递增列(非时间)相关
		Elastic_timestamp_type  int      //递增的时间列相关 1 es日期格式 2 时间戳
		Elastic_timestamp_key   string   //递增的时间列相关
		Elastic_startDate       string   //动态表的开始采集时间
		Elastic_datetimeFormat  string   //动态表时间格式
		Elastic_dateUnit        string   //年表 月表 日表
		Elastic_dateGap         int      //动态表 的 是假间隔 默认 1 (几年一表 几月一表 几天一表)
		Elastic_batch_intervel  string   //elastic 扫描间隔采用 cron 定时任务查询 默认 1m 扫描一次
		Elastic_exec_onstart    bool     //elastic 是否启动时查询一次
		Elastic_schema          []string //类型转换
	}
)

func newOptions(config map[string]interface{}) (opt IOptions, err error) {
	options := &Options{
		//Elastic_sql: "Select * From `$TABLE$`",
		// Elastic_offset_key:     "INCREMENTAL",
		Elastic_collection_type: StaticTableCollection,
		Elastic_timestamp_type:  1, //标准日期格式
		Elastic_limit_batch:     100,
		Elastic_batch_intervel:  "0 */1 * * * ?",
		Elastic_exec_onstart:    false,
	}
	if config != nil {
		if err = mapstructure.Decode(config, options); err == nil {
			err = mapstructure.Decode(config, &options.ReaderOptions)
		}
	}
	eshost := options.GetElastic_addr()
	options.Elastic_ip = strings.Split(eshost, ":")[0]
	if !strings.HasPrefix(eshost, "http://") && !strings.HasPrefix(eshost, "https://") {
		options.Elastic_addr = "http://" + eshost
	}
	// if options.Elastic_timestamp_type == TimestampTypeUnix && options.GetElastic_query_starttime() != "" {
	// 	var times time.Time
	// 	times, err = time.Parse("2006-01-02 15:04:05", options.GetElastic_query_starttime())
	// 	if err != nil {
	// 		err = fmt.Errorf("GetElastic_query_starttime err:%v", err)
	// 		return
	// 	}
	// 	options.Query_starttime = times.Unix()
	// 	// options.Query_starttime = 1637203890000
	// }
	// if options.Elastic_timestamp_type == TimestampTypeUnix && options.GetElastic_query_endtime() != "" {
	// 	var times time.Time
	// 	times, err = time.Parse("2006-01-02 15:04:05", options.GetElastic_query_endtime())
	// 	if err != nil {
	// 		err = fmt.Errorf("GetElastic_query_starttime err:%v", err)
	// 		return
	// 	}
	// 	options.Query_endtime = times.Unix()
	// 	// options.Query_endtime = 1637204220001
	// }
	fmt.Printf("Query_starttime:%d,Query_endtime:%d", options.Query_starttime, options.Query_endtime)
	opt = options
	return
}

func (this *Options) GetElastic_collection_type() MyCollectionType {
	return this.Elastic_collection_type
}

// func (this *Options) GetElastic_taskselfclose() int {
// 	return this.Elastic_taskselfclose
// }
func (this *Options) GetElastic_addr() string {
	return this.Elastic_addr
}
func (this *Options) GetElastic_ip() string {
	return this.Elastic_ip
}

// func (this *Options) GetElastic_database() string {
// 	return this.Elastic_database
// }
func (this *Options) GetElastic_startDate() string {
	return this.Elastic_startDate
}
func (this *Options) GetElastic_dateUnit() string {
	return this.Elastic_dateUnit
}
func (this *Options) GetElastic_dateGap() int {
	return this.Elastic_dateGap
}
func (this *Options) GetElastic_limit_batch() int {
	return this.Elastic_limit_batch
}
func (this *Options) GetElastic_query_starttime() string {
	return this.Elastic_query_starttime
}
func (this *Options) GetElastic_query_endtime() string {
	return this.Elastic_query_endtime
}
func (this *Options) GetQuery_starttime() int64 {
	return this.Query_starttime
}
func (this *Options) GetQuery_endtime() int64 {
	return this.Query_endtime
}
func (this *Options) GetElastic_datetimeFormat() string {
	return this.Elastic_datetimeFormat
}

func (this *Options) GetElastic_tables() []string {
	return this.Elastic_tables
}

// func (this *Options) GetElastic_sql() string {
// 	return this.Elastic_sql
// }

func (this *Options) GetElastic_offset_key() string {
	return this.Elastic_offset_key
}
func (this *Options) GetElastic_timestamp_type() int {
	return this.Elastic_timestamp_type
}
func (this *Options) GetElastic_timestamp_key() string {
	return this.Elastic_timestamp_key
}
func (this *Options) GetElastic_batch_intervel() string {
	return this.Elastic_batch_intervel
}
func (this *Options) GetElastic_exec_onstart() bool {
	return this.Elastic_exec_onstart
}
func (this *Options) GetElastic_schema() []string {
	return this.Elastic_schema
}
