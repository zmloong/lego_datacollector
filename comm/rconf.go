package comm

import "github.com/liwei1dao/lego/core"

type BusinessType string

const (
	Collect BusinessType = "collect"
)

type (
	RunnerConfig struct {
		InstanceId       string                   `json:"instanceid"`        //任务的实例Id
		Name             string                   `json:"name" bson:"_id"`   //采集器名称 唯一
		BusinessType     BusinessType             `json:"businessType"`      //任务分类
		MaxProcs         int                      `json:"maxprocs"`          //采集器任务并发数
		RunIp            []string                 `json:"ip" `               //采集器运行ip 默认 0.0.0.0 自动分配
		IsStopped        bool                     `json:"isstopped"`         //是否停止采集
		MaxBatchLen      int                      `json:"batch_len"`         //采集器批处理最大消息条数
		MaxBatchSize     int                      `json:"batch_size"`        //采集器批处理最大字节数
		MaxBatchInterval int                      `json:"batch_interval"`    //最大发送时间间隔
		DataChanSize     int                      `json:"data_chan_size"`    //采集数据管道缓存大小
		MaxCollDataSzie  uint64                   `json:"max_colldata_size"` //最大采集数据大小
		ReaderConfig     map[string]interface{}   `json:"reader"`            //读取器配置
		ParserConf       map[string]interface{}   `json:"parser"`            //解析器配置
		TranfromssConfig []map[string]interface{} `json:"transforms"`        //转换器配置
		SendersConfig    []map[string]interface{} `json:"senders"`           //发送器配置
	}
)

//获取默认配置信息
func DefRunnerConfig(maxCollDataSzie uint64) *RunnerConfig {
	return &RunnerConfig{
		RunIp:            []string{core.AutoIp},
		MaxProcs:         8,
		MaxBatchLen:      1000,
		MaxBatchSize:     100 * 1024 * 1024,
		MaxBatchInterval: 5,   //最大发送间隔最大五秒
		DataChanSize:     200, //默认数据管道大小
		MaxCollDataSzie:  maxCollDataSzie,
	}
}
