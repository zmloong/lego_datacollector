package comm

import "github.com/liwei1dao/lego/core"

const (
	SM_DataCollectorModule core.M_Modules = "SM_DataCollectorModule" //TieR服务提供模块
)

const (
	RPC_CreateRunner core.Rpc_Key = "RPC_CreateRunner" //创建采集任务
	RPC_StartRunner  core.Rpc_Key = "RPC_StartRunner"  //启动采集任务
	RPC_DriveRunner  core.Rpc_Key = "RPC_DriveRunner"  //驱动采集任务
	RPC_StopRunner   core.Rpc_Key = "RPC_StopRunner"   //停止采集任务
	// RPC_DelRunner    core.Rpc_Key = "RPC_DelRunner"    //删除采集任务
)

const (
	Event_UpdatRunner     core.Event_Key = "Event_UpdatRunner"     //更新采集器服务
	Event_WriteLog        core.Event_Key = "Event_WriteLog"        //写入日志事件
	Event_RunnerAutoClose core.Event_Key = "Event_RunnerAutoClose" //任务自动关闭
)

type (
	IDateCollector interface {
		GetListenPort() int
		GetRunner() []string
	}
	StatisticItem struct {
		SuccCnt int64
		FailCnt int64
	}

	RunnerRuninfo struct {
		RName      string              //tags:{bson:"_id"} 采集器名称
		InstanceId string              //实例Id
		RunService map[string]struct{} //采集器运行服务
		RecordTime int64               //记录时间
		ReaderCnt  int64               //读取总条数
		ParserCnt  int64               //[解析统计]
		SenderCnt  map[string]int64    //[发送成功, 发送失败]
	}
	ServiceNode struct {
		Id        string   `json:"Id"`        //服务id
		IP        string   `json:"IP"`        //服务运行地址
		Port      int      `json:"Port"`      //端口
		Wight     float32  `json:"Wight"`     //权重
		RunnerNum int      `json:"RunnerNum"` //运行任务数
		Runner    []string `json:"Runner"`    //运行采集任务列表
		State     int      `json:"State"`     // 0 停止 1 运行
	}
)
