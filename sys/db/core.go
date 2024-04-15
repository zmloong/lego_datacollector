package db

import (
	"lego_datacollector/comm"

	"github.com/liwei1dao/lego/sys/redis"
	lgredis "github.com/liwei1dao/lego/sys/redis"
)

type (
	ISys interface {
		QueryServiceNodeInfo() (result map[string]*comm.ServiceNode, err error)                          //查询服务节点
		UpDateServiceNodeInfo(info *comm.ServiceNode) (original map[string]*comm.ServiceNode, err error) //更新服务节点列表
		DelRunner(rId string) (err error)                                                                //删除采集任务
		AddNewRunnerConfig(conf *comm.RunnerConfig) (err error)
		UpdateRunnerConfig_IsStop(rId string, isstopped bool) (err error)
		QueryRunnerConfig(rId string) (esult *comm.RunnerConfig, err error)
		ReadMetaData(rId, meta string, value interface{}) (err error)
		WriteMetaData(rId, meta string, value interface{}) (err error)
		WriteRunnerRuninfo(info *comm.RunnerRuninfo) (err error)
		QueryRunnerRuninfo(rId string) (result *comm.RunnerRuninfo, err error)
		CleanRunnerRuninfo(rId string) (err error)
		GetQueryAndStartNoRunRunnerLock() (result *lgredis.RedisMutex, err error)
		QueryAndRunNoRunRunnerConfig(ip string) (result *comm.RunnerConfig, err error)
		RealTimeStatisticsIn(rid string, in int64) (err error)
		RealTimeStatisticsOut(rid, stype string, out int64) (err error)
	}
	RInfoTable struct {
		Id   string                `bson:"_id"`
		Dats []*comm.RunnerRuninfo `bson:"data"`
	}
)

var (
	defsys ISys
)

const ( //Cache
	Cache_ServiceNodeInfo              string = "ServiceNodeInfo"                //数据采集服务节点信息
	Cache_RunnerConfig                 string = "RunnerConfig:"                  //采集器配置信息
	Cache_RunnerRunInfo                string = "RunnerInfo:"                    //采集器缓存信息
	Cache_RunnerMeta                   string = "RunnerMeta:%s-%s"               //采集器元数据
	Cache_RunnerStatistics_Reader      string = "%s_collect-in_%s"               //采集器统计 采集条数
	Cache_RunnerStatistics_Sender      string = "%s_collect-%s-out_%s"           //采集器统计 发送条数
	Cache_QueryAndStartNoRunRunnerLock string = "QueryAndStartNoRunRunnerLock"   //查询并启动为运行的采集任务
	Cach_UpdateRunRunnerInfoLock       string = "Cach_UpdateRunRunnerInfoLock"   //更新采集起运行信息分布式锁
	Cach_UpDateServiceNodeInfoLock     string = "Cach_UpDateServiceNodeInfoLock" //更新采集器服务节点信息分布式锁
)
const ( //DB
// Sql_RunnerConfigsDataTable core.SqlTable = "configs"    //配置文件数据表
// Sql_RunnerRunInfoTable     core.SqlTable = "runnerinfo" //采集器运行信息表
// Sql_RunnerMetaTable        core.SqlTable = "metas"      //元数据表
)

func OnInit(config map[string]interface{}, option ...Option) (err error) {
	defsys, err = newSys(newOptions(config, option...))
	return
}

func NewSys(option ...Option) (sys ISys, err error) {
	sys, err = newSys(newOptionsByOption(option...))
	return
}

func QueryServiceNodeInfo() (result map[string]*comm.ServiceNode, err error) {
	return defsys.QueryServiceNodeInfo()
}

//更新服务节点列表
func UpDateServiceNodeInfo(info *comm.ServiceNode) (original map[string]*comm.ServiceNode, err error) {
	return defsys.UpDateServiceNodeInfo(info)
}

///删除采集任务
func DelRunner(rId string) (err error) {
	return defsys.DelRunner(rId)
}

//添加新的采集任务
func AddNewRunnerConfig(conf *comm.RunnerConfig) (err error) {
	return defsys.AddNewRunnerConfig(conf)
}

func UpdateRunnerConfig_IsStop(rId string, isstopped bool) (err error) {
	return defsys.UpdateRunnerConfig_IsStop(rId, isstopped)
}

func QueryRunnerConfig(rId string) (esult *comm.RunnerConfig, err error) {
	return defsys.QueryRunnerConfig(rId)
}
func ReadMetaData(rId, meta string, value interface{}) (err error) {
	return defsys.ReadMetaData(rId, meta, value)
}
func WriteMetaData(rId, meta string, value interface{}) (err error) {
	return defsys.WriteMetaData(rId, meta, value)
}
func WriteRunnerRuninfo(info *comm.RunnerRuninfo) (err error) {
	return defsys.WriteRunnerRuninfo(info)
}

//查询采集器运行信息
func QueryRunnerRuninfo(rId string) (result *comm.RunnerRuninfo, err error) {
	return defsys.QueryRunnerRuninfo(rId)
}

func CleanRunnerRuninfo(rId string) (err error) {
	return defsys.CleanRunnerRuninfo(rId)
}

//返回分布式锁-查询启动未运行采集器
func GetQueryAndStartNoRunRunnerLock() (result *redis.RedisMutex, err error) {
	return defsys.GetQueryAndStartNoRunRunnerLock()
}

//查询为启动的采集任务
func QueryAndRunNoRunRunnerConfig(ip string) (result *comm.RunnerConfig, err error) {
	return defsys.QueryAndRunNoRunRunnerConfig(ip)
}

func RealTimeStatisticsIn(rid string, in int64) (err error) {
	return defsys.RealTimeStatisticsIn(rid, in)
}
func RealTimeStatisticsOut(rid, stype string, out int64) (err error) {
	return defsys.RealTimeStatisticsOut(rid, stype, out)
}
