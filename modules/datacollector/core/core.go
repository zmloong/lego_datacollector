package core

import (
	"lego_datacollector/comm"

	"github.com/liwei1dao/lego/sys/log"
)

const (
	KeyRaw               = "message"               //采集数据
	KeyTimestamp         = "idss_collect_time"     //采集时间戳
	KeyIdsCollecip       = "idss_collect_ip"       //采集服务器ip
	KeyEqptip            = "eqpt_ip"               //数据源ip
	KeyIdsCollecexpmatch = "idss_collect_expmatch" //正则过滤结果标签
	KeyVendor            = "vendor"                //厂商名称
	KeyDevtype           = "dev_type"              //设备父级名称
	KeyProduct           = "product"               //设备名称
	KeyEqptname          = "eqpt_name"             //设备+厂商名称
	KeyEqptDeviceType    = "eqpt_device_type"      //日志类型名称
)

const (
	Runner_Stoped   RunnerState = iota //已停止
	Runner_Initing                     //初始化中
	Runner_Starting                    //启动中
	Runner_Runing                      //运行中
	Runner_Stoping                     //关闭中
)

var (
	RunnerSuccAutoClose string = "Runner Succ Auto Close"
	RunnerFailAutoClose string = "Runner Fail Auto Close"
)

type (
	RunnerState     int32
	ICollDataBucket interface {
		AddCollData(data ICollData) (full bool)
		IsEmplty() bool
		IsFull() bool
		CurrCap() int
		Items() []ICollData
		SuccItems() []ICollData
		SuccItemsCount() (reslut int)
		ErrItems() []ICollData
		ErrItemsCount() (reslut int)
		Reset()
		SetError(err error)
		Error() (err error)
		ToString(onlysucc bool) (value string, err error) //序列化字符串接口
		MarshalJSON() ([]byte, error)
	}
	//采集数据结构
	ICollData interface {
		SetError(err error)                  //错误信息
		GetError() error                     //错误信息
		GetSource() string                   //数据来源
		GetTime() int64                      //采集时间
		GetData() map[string]interface{}     //整理数据
		GetValue() interface{}               //采集原数据
		GetSize() uint64                     //数据大小
		ToString() (value string, err error) //序列化
		MarshalJSON() ([]byte, error)        //重构json序列化接口
	}
	Statistics struct {
		Name       string
		Type       string
		Statistics int64
	}

	//采集器结构
	IRunner interface {
		log.ILog
		Name() (name string)
		InstanceId() string
		BusinessType() comm.BusinessType
		ServiceId() string
		ServiceIP() string
		MaxProcs() (maxprocs int)
		MaxCollDataSzie() (msgmaxsize uint64)
		GetRunnerState() RunnerState
		Init() (err error)
		Start() (err error)
		Drive() (err error) //驱动工作 提供外部调度器使用 个采集模块根据自己的需求才实现此接口
		Close(state RunnerState, closemsg string) (err error)
		Metaer() IMetaer
		Reader() IReader
		StatisticsIn() chan<- *Statistics
		StatisticsOut() chan<- *Statistics
		ReaderPipe() chan<- ICollData
		Push_ParserPipe(bucket ICollDataBucket)
		ParserPipe() <-chan ICollDataBucket
		Push_TransformsPipe(bucket ICollDataBucket)
		TransformsPipe(index int) (pipe <-chan ICollDataBucket, ok bool)
		Push_NextTransformsPipe(index int, bucket ICollDataBucket)
		Push_SenderPipe(bucket ICollDataBucket)
		SenderPipe(stype string) (pipe <-chan ICollDataBucket, ok bool)
		SyncRunnerInfo()
		StatisticRunner()
	}
	IMetaerData interface {
		GetName() string
		GetMetae() interface{} //注意这处返回 指针对象 map对象许返回&map
	}
	//元数据
	IMetaer interface {
		Init(runner IRunner) (err error)
		Close() (err error)
		Read(meta IMetaerData) (err error)
		Write(meta IMetaerData) (err error)
	}
	//读取器
	IReader interface {
		GetRunner() IRunner
		GetType() string
		GetEncoding() Encoding //编码格式
		Start() (err error)
		Drive() (err error) //驱动工作 外部程序驱动采集器工作
		Close() (err error)
		Input() chan<- ICollData
	}
	//读取器
	IParser interface {
		GetRunner() IRunner
		GetType() string
		Start() (err error)
		Close() (err error)
		Parse(bucket ICollDataBucket)
	}
	//变换器
	ITransforms interface {
		GetRunner() IRunner
		Start() (err error)
		Close() (err error)
		Trans(bucket ICollDataBucket)
	}
	//读取器
	ISender interface {
		GetRunner() IRunner
		GetType() string
		Start() (err error)
		Run(pipeId int, pipe <-chan ICollDataBucket, params ...interface{})
		Close() (err error)
		Send(pipeId int, bucket ICollDataBucket, params ...interface{})
		ReadCnt() int64
		ReadAnResetCnt() int64
	}
)
