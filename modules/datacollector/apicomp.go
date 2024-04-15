package datacollector

import (
	"errors"
	"fmt"
	"lego_datacollector/comm"
	"lego_datacollector/sys/db"
	"time"

	"github.com/liwei1dao/lego"
	"github.com/liwei1dao/lego/base"
	"github.com/liwei1dao/lego/core"
	"github.com/liwei1dao/lego/core/cbase"
	"github.com/liwei1dao/lego/lib/modules/http"
	"github.com/liwei1dao/lego/sys/event"
	"github.com/liwei1dao/lego/sys/log"
	"github.com/liwei1dao/lego/sys/redis"
)

//获取默认配置信息
func DefRunnerConfig(maxCollDataSzie uint64) *_AddNeRunnerReq {
	return &_AddNeRunnerReq{
		RunnerConfig: *comm.DefRunnerConfig(maxCollDataSzie),
		IsClean:      true,
	}
}

type _AddNeRunnerReq struct {
	comm.RunnerConfig
	IsClean bool   `json:"isclean"` //是否清理记录信息
	Sign    string `json:"sign"`
}

type RunnerIds struct {
	RunerId    string `json:"rid"` //任务id
	InstanceId string `json:"iid"` //实例Id
}

type _StartRunnerReq struct {
	Ids     []RunnerIds `json:"ids"`
	ReqTime int64       `json:"reqtime"`
	Sign    string      `json:"sign"`
}
type _DriveRunnerReq struct {
	Ids     []RunnerIds `json:"ids"`
	ReqTime int64       `json:"reqtime"`
	Sign    string      `json:"sign"`
}

type _StopRunnerReq struct {
	Ids     []RunnerIds `json:"ids"`
	ReqTime int64       `json:"reqtime"`
	Sign    string      `json:"sign"`
}
type _DeleteRunnerReq struct {
	Ids     []RunnerIds `json:"ids"`
	ReqTime int64       `json:"reqtime"`
	Sign    string      `json:"sign"`
}

//批处理结果返回
type ResultManyData struct {
	SuccIds []string `json:"succ"`
	FailIds []string `json:"fail"`
	ErrMsg  []string `json:"errmsg"`
}

type ApiComp struct {
	cbase.ModuleCompBase
	module IDatCollector
}

func (this *ApiComp) Init(service core.IService, module core.IModule, comp core.IModuleComp, options core.IModuleOptions) (err error) {
	err = this.ModuleCompBase.Init(service, module, comp, options)
	this.module = module.(IDatCollector)
	logkit := this.module.Group("/datacollector")
	logkit.POST("/addrunner", this.AddNeRunnerReq)
	logkit.POST("/startrunners", this.StartRunnerReq)
	logkit.POST("/driverunners", this.DriveRunnerReq)
	logkit.POST("/stoprunners", this.StopRunnerReq)
	logkit.POST("/delrunners", this.DeleteRunnerReq)
	return
}
func (this *ApiComp) AddNeRunnerReq(c *http.Context) {
	defer lego.Recover("AddNeRunnerReq")
	req := DefRunnerConfig(this.module.GetOptions().GetMaxCollDataSzie())
	c.ShouldBindJSON(req)
	event.TriggerEvent(comm.Event_WriteLog, req.Name, req.InstanceId, fmt.Sprintf("收到任务创建请求:%+v", req), 0, time.Now().Unix())
	log.Debugf("AddNeRunnerReq:%+v", req)
	var (
		code    core.ErrorCode
		err     error
		results map[string]*base.Result
		result  interface{}
	)
	defer func() {
		this.module.HttpStatusOK(c, code, result)
	}()

	if sgin := this.module.ParamSign(map[string]interface{}{"name": req.Name}); sgin != req.Sign {
		log.Errorf("AddNeRunnerReq SignError sgin:%s", sgin)
		code = core.ErrorCode_SignError
		return
	}
	//初步娇艳配置文件是否正确
	if len(req.Name) == 0 || len(req.RunIp) == 0 || req.MaxProcs == 0 ||
		req.MaxBatchInterval == 0 || req.MaxBatchLen == 0 || req.MaxBatchSize == 0 || req.MaxCollDataSzie == 0 ||
		req.ReaderConfig == nil || len(req.ReaderConfig) == 0 ||
		req.ParserConf == nil || len(req.ParserConf) == 0 ||
		req.SendersConfig == nil || len(req.SendersConfig) == 0 {
		err = errors.New("AddNeRunnerReq Configuration error")
		code = core.ErrorCode_ReqParameterError
		result = err.Error()
		return
	}
	if _, err = db.QueryRunnerConfig(req.Name); err == nil { //已经有相同任务了
		if err = this.stoprunner(req.Name); err != nil && err != redis.RedisNil {
			err = fmt.Errorf("AddNeRunnerReq stoprunner %v", err)
			log.Errorf("%v", err)
			code = comm.ErrorCode_StopRunnerError
			result = err.Error()
			return
		}
		if req.IsClean {
			db.DelRunner(req.Name)
		}
	}
	//判断服务任务类型 确定哪些服务可以多实例运行
	switch req.RunnerConfig.ReaderConfig["type"] {
	case "kafka", "socket":
		break
	default:
		req.RunnerConfig.RunIp = req.RunnerConfig.RunIp[0:1]
		break
	}
	if results, err = this.module.GetService().RpcInvokeByIps(req.RunnerConfig.RunIp, "datacollector", comm.RPC_CreateRunner, true, &req.RunnerConfig); err != nil {
		log.Errorf("AddNeRunnerReq err:%v", err)
		code = core.ErrorCode_RpcFuncExecutionError
		result = err.Error()
	} else {
		log.Debugf("AddNeRunnerReq results:%+v", results)
		for _, v := range results {
			log.Debugf("AddNeRunnerReq v:%+v", v)
			if v.Err != nil {
				code = core.ErrorCode_RpcFuncExecutionError
				err = fmt.Errorf("AddNeRunnerReq id:%s ip:%s err:%v ", req.Name, v.Index, v.Err)
				result = err.Error()
				log.Errorf("%v", err)
				go this.stoprunner(req.Name)
				break
			}
		}
	}
}
func (this *ApiComp) StartRunnerReq(c *http.Context) {
	defer lego.Recover("StartRunnerReq")
	req := &_StartRunnerReq{}
	c.ShouldBindJSON(req)
	// for _, v := range req.Ids {
	// 	event.TriggerEvent(comm.Event_WriteLog, v.RunerId, v.InstanceId, fmt.Sprintf("收到任务启动请求:%+v", req), 0, time.Now().Unix())
	// }
	defer log.Debugf("StartRunnerReq:%+v", req)
	var (
		code    core.ErrorCode
		err     error
		errstr  string
		rconf   *comm.RunnerConfig
		result  *ResultManyData
		results map[string]*base.Result
	)
	defer func() {
		this.module.HttpStatusOK(c, code, result)
	}()
	if sgin := this.module.ParamSign(map[string]interface{}{"reqtime": req.ReqTime}); sgin != req.Sign {
		log.Errorf("StartRunnerReq SignError sgin:%s", sgin)
		code = core.ErrorCode_SignError
		return
	}
	if len(req.Ids) == 0 {
		err = errors.New("StartRunnerReq ReqParameter error")
		code = core.ErrorCode_ReqParameterError
		return
	}
	result = &ResultManyData{
		SuccIds: make([]string, 0),
		FailIds: make([]string, 0),
		ErrMsg:  make([]string, 0),
	}
	for _, v := range req.Ids {
		if _, err = db.QueryRunnerRuninfo(v.RunerId); err == nil {
			err = fmt.Errorf("StartRunnerReq Runner:%s is already start", v)
			log.Errorf("%s", err)
			result.SuccIds = append(result.SuccIds, v.RunerId)
			continue
		}
		if rconf, err = db.QueryRunnerConfig(v.RunerId); err != nil {
			err = fmt.Errorf("StartRunnerReq:%s err:%v", v, err)
			log.Errorf("%s", err)
			result.FailIds = append(result.FailIds, v.RunerId)
			result.ErrMsg = append(result.ErrMsg, err.Error())
			continue
		}
		if results, err = this.module.GetService().RpcInvokeByIps(rconf.RunIp, "datacollector", comm.RPC_StartRunner, true, v.RunerId, v.InstanceId); err != nil {
			err = fmt.Errorf("StartRunnerReq:%s err:%v", v, err)
			log.Errorf("%s", err)
			result.FailIds = append(result.FailIds, v.RunerId)
			result.ErrMsg = append(result.ErrMsg, err.Error())
			continue
		} else {
			for _, v1 := range results {
				if v1.Err != nil {
					errstr = fmt.Sprintf("StartRunnerReq id:%s ip:%s err:%v ", v, v1.Index, v1.Err)
					result.FailIds = append(result.FailIds, v.RunerId)
					result.ErrMsg = append(result.ErrMsg, errstr)
					log.Errorf("%s", errstr)
					go this.stoprunner(v.RunerId)
					continue
				}
			}
		}
		result.SuccIds = append(result.SuccIds, v.RunerId)
	}
}
func (this *ApiComp) DriveRunnerReq(c *http.Context) {
	defer lego.Recover("DriveRunnerReq")
	req := &_DriveRunnerReq{}
	c.ShouldBindJSON(req)
	// for _, v := range req.Ids {
	// 	event.TriggerEvent(comm.Event_WriteLog, v, fmt.Sprintf("收到任务驱动请求:%+v", req), 0, time.Now().Unix())
	// }
	defer log.Debugf("DriveRunnerReq:%+v", req)
	var (
		code   core.ErrorCode
		err    error
		result *ResultManyData
	)
	defer func() {
		this.module.HttpStatusOK(c, code, result)
	}()
	if sgin := this.module.ParamSign(map[string]interface{}{"reqtime": req.ReqTime}); sgin != req.Sign {
		log.Errorf("DriveRunnerReq SignError sgin:%s", sgin)
		code = core.ErrorCode_SignError
		return
	}
	if len(req.Ids) == 0 {
		err = errors.New("DriveRunnerReq ReqParameter err")
		code = core.ErrorCode_ReqParameterError
		return
	}
	result = &ResultManyData{
		SuccIds: make([]string, 0),
		FailIds: make([]string, 0),
		ErrMsg:  make([]string, 0),
	}
	for _, v := range req.Ids {
		if err = this.driverunner(v.RunerId); err != nil {
			log.Errorf("DriveRunnerReq:%s err:%v", v, err)
			result.FailIds = append(result.FailIds, v.RunerId)
			result.ErrMsg = append(result.ErrMsg, err.Error())
			continue
		}
		result.SuccIds = append(result.SuccIds, v.RunerId)
	}
}
func (this *ApiComp) StopRunnerReq(c *http.Context) {
	defer lego.Recover("StopRunnerReq")
	req := &_StopRunnerReq{}
	c.ShouldBindJSON(req)
	// for _, v := range req.Ids {
	// 	event.TriggerEvent(comm.Event_WriteLog, v.RunerId, v.InstanceId, fmt.Sprintf("收到任务停止请求:%+v", req), 0, time.Now().Unix())
	// }
	defer log.Debugf("StopRunnerReq:%+v", req)
	var (
		code   core.ErrorCode
		err    error
		result *ResultManyData
	)
	defer func() {
		this.module.HttpStatusOK(c, code, result)
	}()
	if sgin := this.module.ParamSign(map[string]interface{}{"reqtime": req.ReqTime}); sgin != req.Sign {
		log.Errorf("StopRunnerReq SignError sgin:%s", sgin)
		code = core.ErrorCode_SignError
		return
	}
	if len(req.Ids) == 0 {
		err = errors.New("StopRunnerReq ReqParameter err")
		code = core.ErrorCode_ReqParameterError
		return
	}
	result = &ResultManyData{
		SuccIds: make([]string, 0),
		FailIds: make([]string, 0),
		ErrMsg:  make([]string, 0),
	}
	for _, v := range req.Ids {
		if err = this.stoprunner(v.RunerId); err != nil {
			if err != redis.RedisNil {
				log.Errorf("StopRunnerReq:%s err:%v", v, err)
				result.FailIds = append(result.FailIds, v.RunerId)
				result.ErrMsg = append(result.ErrMsg, err.Error())
			} else {
				result.SuccIds = append(result.SuccIds, v.RunerId)
			}
			continue
		}
		result.SuccIds = append(result.SuccIds, v.RunerId)
	}
}
func (this *ApiComp) DeleteRunnerReq(c *http.Context) {
	defer lego.Recover("DeleteRunnerReq")
	req := &_DeleteRunnerReq{}
	c.ShouldBindJSON(req)
	// for _, v := range req.Ids {
	// 	event.TriggerEvent(comm.Event_WriteLog, v.RunerId, v.InstanceId, fmt.Sprintf("收到任务删除请求:%+v", req), 0, time.Now().Unix())
	// }
	defer log.Debugf("DeleteRunnerReq:%+v", req)
	var (
		code   core.ErrorCode
		err    error
		result *ResultManyData
	)
	defer func() {
		this.module.HttpStatusOK(c, code, result)
	}()

	if sgin := this.module.ParamSign(map[string]interface{}{"reqtime": req.ReqTime}); sgin != req.Sign {
		log.Errorf("DeleteRunnerReq SignError sgin:%s", sgin)
		code = core.ErrorCode_SignError
		return
	}

	if len(req.Ids) == 0 {
		err = errors.New("DeleteRunnerReq ReqParameter err")
		code = core.ErrorCode_ReqParameterError
		return
	}
	result = &ResultManyData{
		SuccIds: make([]string, 0),
		FailIds: make([]string, 0),
		ErrMsg:  make([]string, 0),
	}
	for _, v := range req.Ids {
		if err = this.stoprunner(v.RunerId); err != nil {
			if err != redis.RedisNil {
				log.Errorf("DeleteRunnerReq:%s err:%v", v, err)
				result.FailIds = append(result.FailIds, v.RunerId)
				result.ErrMsg = append(result.ErrMsg, err.Error())
			} else {
				result.SuccIds = append(result.SuccIds, v.RunerId)
			}
			continue
		}
		if err = db.DelRunner(v.RunerId); err != nil {
			log.Errorf("DeleteRunnerReq:%s err:%v", v, err)
			result.FailIds = append(result.FailIds, v.RunerId)
			result.ErrMsg = append(result.ErrMsg, err.Error())
			continue
		}
		result.SuccIds = append(result.SuccIds, v.RunerId)
	}
}
func (this *ApiComp) stoprunner(rid string) (err error) {
	var (
		runner_runinfo *comm.RunnerRuninfo
		ids            []string
		results        map[string]*base.Result
		n              = 0
	)
	if runner_runinfo, err = db.QueryRunnerRuninfo(rid); err == nil {
		ids = make([]string, len(runner_runinfo.RunService))
		for k, _ := range runner_runinfo.RunService {
			ids[n] = k
			n++
		}
		if results, err = this.module.GetService().RpcInvokeByIds(ids, comm.RPC_StopRunner, true, rid, runner_runinfo.InstanceId); err != nil && err != Error_NoRunner {
			err = fmt.Errorf("stoprunner:%s err:%v", rid, err)
			return
		} else {
			for _, v := range results {
				if v.Err != nil {
					log.Errorf("stoprunner:%s ip:%s err:%v", rid, v.Index, v.Err)
					err = v.Err
					return
				}
			}
		}
	} else {
		// err = fmt.Errorf("no found %s", rid)
		log.Errorf("stoprunner no found %s", rid)
		err = nil
	}
	return
}
func (this *ApiComp) driverunner(rid string) (err error) {
	var (
		runner_runinfo *comm.RunnerRuninfo
		ids            []string
		results        map[string]*base.Result
		n              = 0
	)
	if runner_runinfo, err = db.QueryRunnerRuninfo(rid); err == nil {
		ids = make([]string, len(runner_runinfo.RunService))
		for k, _ := range runner_runinfo.RunService {
			ids[n] = k
			n++
		}
		if results, err = this.module.GetService().RpcInvokeByIds(ids, comm.RPC_DriveRunner, true, rid, runner_runinfo.InstanceId); err != nil && err.Error() != Error_NoRunner.Error() {
			err = fmt.Errorf("driverunner:%s err:%v", rid, err)
			return
		} else {
			for _, v1 := range results {
				if v1.Err != nil {
					log.Errorf("driverunner:%s ip:%s err:%v", rid, v1.Index, v1.Err)
					err = v1.Err
					return
				}
			}
		}
	} else {
		log.Errorf("driverunner:%s on run err:%v", rid, err)
		err = nil
	}
	return
}
