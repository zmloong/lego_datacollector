package datacollector

import (
	"errors"
	"fmt"
	"lego_datacollector/comm"
	"lego_datacollector/modules/datacollector/core"
	"lego_datacollector/sys/db"
	"sync"

	"github.com/liwei1dao/lego"
	"github.com/liwei1dao/lego/base"
	lgcore "github.com/liwei1dao/lego/core"
	"github.com/liwei1dao/lego/core/cbase"
	lgcron "github.com/liwei1dao/lego/sys/cron"
	"github.com/liwei1dao/lego/sys/event"
	"github.com/liwei1dao/lego/sys/log"
	"github.com/liwei1dao/lego/sys/redis"
	"github.com/robfig/cron/v3"
)

type RunnerMgrComp struct {
	cbase.ModuleCompBase
	options           IOptions
	service           base.IClusterService
	lock              sync.RWMutex
	runners           map[string]core.IRunner
	scanrunner        cron.EntryID
	syncrunnerId      cron.EntryID
	statisticrunnerId cron.EntryID
}

func (this *RunnerMgrComp) GetRunners() (runners []string) {
	runners = make([]string, len(this.runners))
	n := 0
	for k, v := range this.runners {
		if v.GetRunnerState() == core.Runner_Runing {
			runners[n] = k
			n++
		}
	}
	return runners[0:n]
}

func (this *RunnerMgrComp) Init(service lgcore.IService, module lgcore.IModule, comp lgcore.IModuleComp, options lgcore.IModuleOptions) (err error) {
	err = this.ModuleCompBase.Init(service, module, comp, options)
	this.options = options.(IOptions)
	this.service = service.(base.IClusterService)
	this.runners = make(map[string]core.IRunner)
	if this.syncrunnerId, err = lgcron.AddFunc(this.options.GetRinfoSyncInterval(), this.syncRunnerInfo); err != nil {
		log.Errorf("RunnerMgrComp Init start syncRunnerInfo err:%v", err)
		return
	}
	//修改成实时统计 此处展示屏蔽
	// if this.statisticrunnerId, err = lgcron.AddFunc(this.options.GetRStatisticInterval(), this.statisticRunner); err != nil {
	// 	log.Errorf("RunnerMgrComp Init start statisticRunner err:%v", err)
	// 	return
	// }
	return
}

func (this *RunnerMgrComp) Start() (err error) {
	if err = this.ModuleCompBase.Start(); err != nil {
		return
	}
	//加入集群之后 再开始具体业务
	event.RegisterGO(lgcore.Event_RegistryStart, func() {
		if this.options.GetIsAutoStartRunner() {
			log.Debugf("RunnerMgrComp 开启定时扫描")
			go this.ScanRunnerTask() //启动服务默认启动一次
			if this.scanrunner, err = lgcron.AddFunc(this.options.GetScanNoStartRunnerInterval(), this.ScanRunnerTask); err != nil {
				log.Errorf("RunnerMgrComp lgcron.AddFunc err:%v", err)
			}
		}
	})
	event.RegisterGO(lgcore.Event_LoseService, this.Event_LoseService)
	return
}

func (this *RunnerMgrComp) Destroy() (err error) {
	err = this.ModuleCompBase.Destroy()
	lgcron.Remove(this.scanrunner)
	lgcron.Remove(this.syncrunnerId)
	lgcron.Remove(this.statisticrunnerId)
	this.lock.Lock()
	defer this.lock.Unlock()
	for _, v := range this.runners {
		if err = v.Close(core.Runner_Runing, "DataCollector send del signal"); err != nil {
			log.Errorf("RunnerMgrComp Destroy")
		}
	}
	return
}

func (this *RunnerMgrComp) Event_LoseService(sID string) {
	if sID != this.service.GetId() {
		this.ScanRunnerTask()
	}
}

//查询并启动为执行采集器
func (this *RunnerMgrComp) ScanRunnerTask() {
	var (
		err   error
		lock  *redis.RedisMutex
		rconf *comm.RunnerConfig
	)
	if lock, err = db.GetQueryAndStartNoRunRunnerLock(); err != nil {
		log.Errorf("ScanRunnerTask GetQueryAndStartNoRunRunnerLock err:%v", err)
		return
	}
	lock.Lock()
	defer lock.Unlock()
	if rconf, err = db.QueryAndRunNoRunRunnerConfig(this.service.GetIp()); err != nil {
		if err == redis.RedisNil { //当前没有为启动采集任务
			return
		} else {
			log.Errorf("ScanRunnerTask QueryAndRunNoRunRunnerConfig err:%v", err)
			return
		}
	}
	if err = this.StartRunnerByConfig(rconf); err != nil {
		log.Errorf("ScanRunnerTask StartRunnerByConfig err:%v", err)
	}
	return
}

//启动采集器
func (this *RunnerMgrComp) StartRunnerByConfig(conf *comm.RunnerConfig) (err error) {
	var (
		ok     bool
		runner core.IRunner
	)
	defer lego.Recover("StartRunnerByConfig")
	this.lock.RLock()
	runner, ok = this.runners[conf.Name]
	this.lock.RUnlock()
	if !ok || runner.GetRunnerState() == core.Runner_Stoped {
		if runner, err = newRunner(conf,
			this.service.GetId(),
			this.service.GetIp(),
		); err == nil {
			if err = runner.Start(); err == nil {
				this.lock.Lock()
				this.runners[conf.Name] = runner
				this.lock.Unlock()
				event.TriggerEvent(comm.Event_UpdatRunner) //通知更新采集服务列表
				log.Infof("StartRunnerByConfig Runner:%s Succ!", conf.Name)
			} else {
				runner.Close(core.Runner_Starting, fmt.Sprintf("Start err:%v", err))
				err = fmt.Errorf("StartRunnerByConfig Start err:%v", err)
			}
		} else {
			err = fmt.Errorf("StartRunnerByConfig newRunner err:%v", err)
		}
	} else {
		err = errors.New("runner is already")
	}
	return
}

//停止采集器
func (this *RunnerMgrComp) DriveRunner(rId string) (err error) {
	var (
		ok     bool
		runner core.IRunner
	)
	this.lock.RLock()
	runner, ok = this.runners[rId]
	this.lock.RUnlock()
	if ok {
		if err = runner.Drive(); err != nil {
			log.Errorf("DriveRunner err:%v", err)
		}
	} else {
		err = Error_NoRunner
	}
	return
}

//停止采集器
func (this *RunnerMgrComp) StopRunner(rId string) (err error) {
	var (
		ok     bool
		runner core.IRunner
	)
	defer lego.Recover("StopRunner")
	this.lock.RLock()
	runner, ok = this.runners[rId]
	this.lock.RUnlock()
	if ok {
		if err = runner.Close(core.Runner_Runing, "RunnerMgrComp send stop signal"); err == nil {
			this.lock.RLock()
			delete(this.runners, rId)
			this.lock.RUnlock()
		} else {
			log.Errorf("StopRunner err:%v", err)
		}
	} else {
		err = Error_NoRunner
	}
	return
}

//删除采集器
func (this *RunnerMgrComp) DelRunner(conf *comm.RunnerConfig) (err error) {
	var (
		ok     bool
		runner core.IRunner
	)
	defer lego.Recover("DelRunner")
	this.lock.RLock()
	runner, ok = this.runners[conf.Name]
	this.lock.RUnlock()
	if ok {
		if err = runner.Close(core.Runner_Runing, "RunnerMgrComp send del signal"); err == nil {
			this.lock.Lock()
			delete(this.runners, conf.Name)
			this.lock.Unlock()
			event.TriggerEvent(comm.Event_UpdatRunner)
		}
	} else {
		err = Error_NoRunner
	}
	return
}

func (this *RunnerMgrComp) syncRunnerInfo() {
	this.lock.RLock()
	for _, v := range this.runners {
		if v.GetRunnerState() == core.Runner_Runing {
			go v.SyncRunnerInfo()
		}
	}
	this.lock.RUnlock()
}

func (this *RunnerMgrComp) statisticRunner() {
	this.lock.RLock()
	for _, v := range this.runners {
		if v.GetRunnerState() == core.Runner_Runing {
			go v.StatisticRunner()
		}
	}
	this.lock.RUnlock()
}
