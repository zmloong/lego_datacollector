package datacollector

import (
	"fmt"
	"lego_datacollector/comm"
	"lego_datacollector/sys/db"
	nethttp "net/http"
	"reflect"
	"sort"
	"strings"
	"time"

	"github.com/liwei1dao/lego"
	"github.com/liwei1dao/lego/base"
	"github.com/liwei1dao/lego/core"
	"github.com/liwei1dao/lego/lib/modules/http"
	"github.com/liwei1dao/lego/sys/event"
	"github.com/liwei1dao/lego/sys/log"
	"github.com/liwei1dao/lego/utils/crypto/md5"
)

func NewModule() core.IModule {
	m := new(DataCollector)
	return m
}

type DataCollector struct {
	http.Http
	service  base.IClusterService
	options  IOptions
	apicomp  *ApiComp
	rmgrcomp *RunnerMgrComp
}

func (this *DataCollector) GetType() core.M_Modules {
	return comm.SM_DataCollectorModule
}

func (this *DataCollector) NewOptions() (options core.IModuleOptions) {
	return new(Options)
}

func (this *DataCollector) GetOptions() (options IOptions) {
	return this.options
}

func (this *DataCollector) GetService() base.IClusterService {
	return this.service
}
func (this *DataCollector) GetListenPort() int {
	return this.options.GetListenPort()
}
func (this *DataCollector) GetRunner() []string {
	return this.rmgrcomp.GetRunners()
}

func (this *DataCollector) Init(service core.IService, module core.IModule, options core.IModuleOptions) (err error) {
	this.service = service.(base.IClusterService)
	this.options = options.(IOptions)
	if err = this.Http.Init(service, module, options); err != nil {
		return
	}
	return
}

func (this *DataCollector) Start() (err error) {
	if err = this.Http.Start(); err != nil {
		return
	}
	this.service.RegisterGO(comm.RPC_CreateRunner, this.RPC_CreateRunner)
	this.service.RegisterGO(comm.RPC_StartRunner, this.RPC_StartRunner)
	this.service.RegisterGO(comm.RPC_StopRunner, this.RPC_StopRunner)
	// this.service.RegisterGO(comm.RPC_DelRunner, this.RPC_DelRunner)
	return
}

func (this *DataCollector) OnInstallComp() {
	this.Http.OnInstallComp()
	this.apicomp = this.RegisterComp(new(ApiComp)).(*ApiComp)
	this.rmgrcomp = this.RegisterComp(new(RunnerMgrComp)).(*RunnerMgrComp)
}

//签名接口
func (this *DataCollector) ParamSign(param map[string]interface{}) (sign string) {
	var keys []string
	for k, _ := range param {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	builder := strings.Builder{}
	for _, v := range keys {
		builder.WriteString(v)
		builder.WriteString("=")
		switch reflect.TypeOf(param[v]).Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr, reflect.Float32, reflect.Float64:
			builder.WriteString(fmt.Sprintf("%d", param[v]))
			break
		default:
			builder.WriteString(fmt.Sprintf("%s", param[v]))
			break
		}
		builder.WriteString("&")
	}
	builder.WriteString("key=" + this.options.GetSignKey())
	log.Infof("orsign:%s", builder.String())
	sign = md5.MD5EncToLower(builder.String())
	return
}

//统一输出接口
func (this *DataCollector) HttpStatusOK(c *http.Context, code core.ErrorCode, data interface{}) {
	defer log.Debugf("%s resp: code:%d data:%+v", c.Request.RequestURI, code, data)
	c.JSON(nethttp.StatusOK, http.OutJson{ErrorCode: code, Message: comm.GetErrorCodeMsg(code), Data: data})
}

//RPC-----------------------------------------------------------------------------------------------------------
//启动采集任务
func (this *DataCollector) RPC_CreateRunner(conf *comm.RunnerConfig) (result string, errstr string) {
	defer lego.Recover("RPC_CreateRunner")
	var (
		err error
	)
	log.Debugf("RPC_CreateRunner:%+v", conf)
	event.TriggerEvent(comm.Event_WriteLog, conf.Name, conf.InstanceId, "启动任务", 0, time.Now().Unix())
	defer func() {
		if err != nil {
			event.TriggerEvent(comm.Event_WriteLog, conf.Name, conf.InstanceId, fmt.Sprintf("启动任务失败 错误:%v", err), 1, time.Now().Unix())
		} else {
			event.TriggerEvent(comm.Event_WriteLog, conf.Name, conf.InstanceId, "启动任务成功", 0, time.Now().Unix())
		}
	}()
	if err = this.rmgrcomp.StartRunnerByConfig(conf); err != nil {
		errstr = err.Error()
		log.Errorf("RPC_CreateRunner rmgrcomp.StartRunnerByConfig err:%v", err)
	} else {
		if err = db.AddNewRunnerConfig(conf); err != nil {
			errstr = err.Error()
			log.Errorf("RPC_CreateRunner db.AddNewRunnerConfig err:%v", err)
		}
	}

	return
}

func (this *DataCollector) RPC_StartRunner(rId, iid string) (result string, errstr string) {
	defer lego.Recover("RPC_StartRunner")
	var (
		rconf *comm.RunnerConfig
		rinfo *comm.RunnerRuninfo
		err   error
	)
	result = rId
	event.TriggerEvent(comm.Event_WriteLog, rId, iid, "启动任务", 0, time.Now().Unix())
	defer func() {
		if err != nil {
			event.TriggerEvent(comm.Event_WriteLog, rId, iid, fmt.Sprintf("启动任务失败 错误:%v", err), 1, time.Now().Unix())
		} else {
			event.TriggerEvent(comm.Event_WriteLog, rId, iid, "启动任务成功", 0, time.Now().Unix())
		}
	}()
	if rinfo, err = db.QueryRunnerRuninfo(rId); err == nil {
		errstr = fmt.Sprintf("Runner:%s is already start:%+v", rId, rinfo)
		log.Errorf("RPC_StartRunner db.QueryRunnerRuninfo err:%s", errstr)
	} else {
		if rconf, err = db.QueryRunnerConfig(rId); err != nil {
			errstr = fmt.Sprintf("Runner:%s is no exist", rId)
			log.Errorf("RPC_StartRunner db.QueryRunnerConfig err:%s", errstr)
		} else {
			rconf.InstanceId = iid
			if err = this.rmgrcomp.StartRunnerByConfig(rconf); err != nil {
				errstr = err.Error()
				log.Errorf("RPC_StartRunner rmgrcomp.StartRunnerByConfig err:%v", err)
			} else {
				db.AddNewRunnerConfig(rconf)
			}
		}
	}
	return
}

//驱动采集任务
func (this *DataCollector) RPC_DriveRunner(rId, iid string) (result string, errstr string) {
	defer lego.Recover("RPC_DriveRunner")
	var (
		rinfo *comm.RunnerRuninfo
		err   error
	)
	result = rId
	event.TriggerEvent(comm.Event_WriteLog, rId, iid, "驱动任务", 0, time.Now().Unix())
	defer func() {
		if err != nil {
			event.TriggerEvent(comm.Event_WriteLog, rId, iid, fmt.Sprintf("驱动任务失败 错误:%v", err), 1, time.Now().Unix())
		} else {
			event.TriggerEvent(comm.Event_WriteLog, rId, iid, "驱动任务成功", 0, time.Now().Unix())
		}
	}()
	if rinfo, err = db.QueryRunnerRuninfo(rId); err != nil {
		errstr = fmt.Sprintf("Runner:%s is no start:%+v", rId, rinfo)
		log.Errorf("RPC_DriveRunner err %s", errstr)
	} else {
		if err = this.rmgrcomp.DriveRunner(rId); err != nil {
			errstr = err.Error()
			log.Errorf("RPC_DriveRunner DriveRunner err %v", err)
		}
	}
	return
}

//停止采集任务
func (this *DataCollector) RPC_StopRunner(rId, iid string) (result string, errstr string) {
	defer lego.Recover("RPC_StopRunner")
	var (
		err error
	)
	result = rId
	event.TriggerEvent(comm.Event_WriteLog, rId, iid, "停止任务", 0, time.Now().Unix())
	defer func() {
		if iid != "" {
			if err != nil && err != Error_NoRunner && err != Error_RunnerStoped && err != Error_RunnerStoping {
				event.TriggerEvent(comm.Event_WriteLog, rId, iid, fmt.Sprintf("停止任务失败 错误:%v", err), 1, time.Now().Unix())
			} else if err != nil && err == Error_RunnerStoping {
				event.TriggerEvent(comm.Event_WriteLog, rId, iid, "任务停止中...", 0, time.Now().Unix())
			} else {
				event.TriggerEvent(comm.Event_WriteLog, rId, iid, "停止任务成功", 0, time.Now().Unix())
			}
		}
	}()
	if err = this.rmgrcomp.StopRunner(rId); err != nil && err != Error_NoRunner && err != Error_RunnerStoped {
		errstr = fmt.Sprintf("RPC_StopRunner:%s is stop err:%v", rId, err)
		log.Errorf("%s", errstr)
		return
	}
	db.UpdateRunnerConfig_IsStop(rId, true)
	return
}

// //启动采集任务
// func (this *DataCollector) RPC_DelRunner(rId string) (result string, errstr string) {
// 	var (
// 		rconf *comm.RunnerConfig
// 		err   error
// 	)
// 	result = rId

// 	db.DelRunner(rId)
// 	return
// }
