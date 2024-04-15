package services

import (
	"fmt"
	"io/ioutil"
	"lego_datacollector/comm"
	"net/http"
	"time"

	"github.com/liwei1dao/lego/base"
	"github.com/liwei1dao/lego/core"
	"github.com/liwei1dao/lego/core/cbase"
	"github.com/liwei1dao/lego/sys/event"
	"github.com/liwei1dao/lego/sys/log"
	"github.com/liwei1dao/lego/sys/sql"
	"github.com/liwei1dao/lego/utils/mapstructure"
)

func NewServiceDataSyncInfoPushComp() core.IServiceComp {
	comp := new(ServiceDataSyncInfoPushComp)
	return comp
}

/*
维护采集任务集群配置组件
*/
type (
	ServiceDataSyncInfoPushCompOptions struct {
		cbase.ServiceCompOptions
		IsOpen                bool
		Mysql_addr            string
		Mysql_table           string
		DataSyncInfoPush_addr string
	}
	ServiceDataSyncInfoPushComp struct {
		cbase.ServiceCompBase
		service base.IClusterService
		options *ServiceDataSyncInfoPushCompOptions
		sql     sql.ISys
	}
)

///参数序列化
func (this *ServiceDataSyncInfoPushCompOptions) LoadConfig(settings map[string]interface{}) (err error) {
	if err = this.ServiceCompOptions.LoadConfig(settings); err == nil {
		if settings != nil {
			err = mapstructure.Decode(settings, this)
		}
	}
	return
}
func (this *ServiceDataSyncInfoPushComp) GetName() core.S_Comps {
	return "DataSyncInfoPushComp"
}

func (this *ServiceDataSyncInfoPushComp) NewOptions() (options core.ICompOptions) {
	return new(ServiceDataSyncInfoPushCompOptions)
}

func (this *ServiceDataSyncInfoPushComp) Init(service core.IService, comp core.IServiceComp, options core.ICompOptions) (err error) {
	err = this.ServiceCompBase.Init(service, comp, options)
	this.service = service.(base.IClusterService)
	this.options = options.(*ServiceDataSyncInfoPushCompOptions)
	if this.options.IsOpen {
		this.sql, err = sql.NewSys(sql.SetSqlType(sql.MySql), sql.SetSqlUrl(this.options.Mysql_addr))
	}
	return
}

func (this *ServiceDataSyncInfoPushComp) Start() (err error) {
	err = this.ServiceCompBase.Start()
	if this.options.IsOpen {
		event.RegisterGO(comm.Event_WriteLog, this.WriteLog)
		event.RegisterGO(comm.Event_RunnerAutoClose, this.RunnerAutoCloseNotice)
	}
	return
}

func (this *ServiceDataSyncInfoPushComp) Destroy() (err error) {
	err = this.ServiceCompBase.Destroy()
	this.sql.Close()
	return
}

//Event---------------------------------------------------------------------------------------------------------
func (this *ServiceDataSyncInfoPushComp) WriteLog(rid, iid, logstr string, iswarn int, t int64) {
	// log.Debugf("rid:%s iid:%s iswarn:%d logstr:%s", rid, iid, iswarn, logstr)
	sqlstr := fmt.Sprintf(`insert into %s(id, syn_number, content, warn_type, log_type, client_id,flow_num, create_staff, create_date, update_staff, update_date, status_cd, remark)
	VALUES(dmp.NEXTVAL('DMP_SEQ'), '%s', '%s',%d, 0 ,'%s','%s', null,'%s', null, null,'1000', null);`, this.options.Mysql_table, rid, logstr, iswarn, this.service.GetId(), iid, time.Unix(t, 0).Format("2006-01-02 15:04:05"))
	if _, err := this.sql.ExecContext(sqlstr); err != nil {
		log.Errorf("err:%v", err)
	}
}

///采集任务自动结束通知 数据同步任务完成通知
func (this *ServiceDataSyncInfoPushComp) RunnerAutoCloseNotice(rid, iid string, issucc bool) {
	state := 5
	if issucc {
		state = 4
	}
	if resp, err := http.Post(fmt.Sprintf(this.options.DataSyncInfoPush_addr+"?clientId=%s&code=%s&flowNum=%s&status=%d", this.service.GetId(), rid, iid, state),
		"application/json",
		nil,
	); err == nil {
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err == nil {
			log.Debugf("RunnerAutoCloseNotice result:%s", string(body))
		} else {
			log.Errorf("RunnerAutoCloseNotice err:%v", err)
		}
	} else {
		log.Errorf("RunnerAutoCloseNotice err:%v", err)
	}
}
