package datacollector

import (
	"github.com/liwei1dao/lego/lib/modules/http"
	"github.com/liwei1dao/lego/utils/mapstructure"
)

type (
	IOptions interface {
		http.IOptions
		GetSignKey() string
		GetIsAutoStartRunner() bool
		GetScanNoStartRunnerInterval() string
		GetRinfoSyncInterval() string
		GetRStatisticInterval() string
		GetMaxCollDataSzie() uint64
	}
	Options struct {
		http.Options
		SignKey                   string //签名密钥
		IsAutoStartRunner         bool   //是否自动拉起Runner
		ScanNoStartRunnerInterval string //扫描为启动采集任务间隔
		RinfoSyncInterval         string //RunerInfo 同步间隔
		RStatisticInterval        string //RunerInfo 统计间隔
		MaxCollDataSzie           uint64
	}
)

func (this *Options) GetSignKey() string {
	return this.SignKey
}

func (this *Options) GetScanNoStartRunnerInterval() string {
	return this.ScanNoStartRunnerInterval
}

func (this *Options) GetIsAutoStartRunner() bool {
	return this.IsAutoStartRunner
}

func (this *Options) GetRinfoSyncInterval() string {
	return this.RinfoSyncInterval
}

func (this *Options) GetRStatisticInterval() string {
	return this.RStatisticInterval
}

func (this *Options) GetMaxCollDataSzie() uint64 {
	return this.MaxCollDataSzie
}

func (this *Options) LoadConfig(settings map[string]interface{}) (err error) {
	this.ScanNoStartRunnerInterval = "0 */10 * * * ?" //默认每十分钟执行一次
	this.RinfoSyncInterval = "0 */1 * * * ?"          //默认每分钟执行一次
	this.RStatisticInterval = "59 59 * * * ?"         //每小时的59分59秒执行
	this.MaxCollDataSzie = 2 * 1024 * 1024
	this.IsAutoStartRunner = true
	if err = this.Options.LoadConfig(settings); err == nil {
		if settings != nil {
			err = mapstructure.Decode(settings, this)
		}
	}
	return
}
