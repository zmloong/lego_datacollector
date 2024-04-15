package snmptrap

import (
	"fmt"
	"lego_datacollector/modules/datacollector/core"
	"time"

	"github.com/mitchellh/mapstructure"
)

type (
	IOptions interface {
		core.IReaderOptions //继承基础配置
		GetSnmptrap_address() string
		GetTrap_time_out() time.Duration
		GetTrap_interval() time.Duration
	}
	Options struct {
		core.ReaderOptions //继承基础配置
		Snmptrap_address   string
		Snmptrap_time_out  string
		Snmptrap_interval  string
		Trap_time_out      time.Duration
		Trap_interval      time.Duration
	}
)

func newOptions(config map[string]interface{}) (opt IOptions, err error) {
	time_out, err := time.ParseDuration("30s")
	if err != nil {
		return nil, err
	}
	interval, err := time.ParseDuration("30s")
	if err != nil {
		return nil, err
	}
	options := &Options{
		Trap_time_out: time_out,
		Trap_interval: interval,
	}
	if config != nil {
		if err = mapstructure.Decode(config, options); err == nil {
			err = mapstructure.Decode(config, &options.ReaderOptions)
		}
	}
	if len(options.Snmptrap_time_out) != 0 {
		if options.Trap_time_out, err = time.ParseDuration(options.Snmptrap_time_out); err != nil {
			err = fmt.Errorf("Reader snmptrap newOptions err:Snmptrap_time_out%s", options.Snmptrap_time_out)
			return
		}
	}
	if len(options.Snmptrap_interval) != 0 {
		if options.Trap_interval, err = time.ParseDuration(options.Snmptrap_interval); err != nil {
			err = fmt.Errorf("Reader snmptrap newOptions err:Snmptrap_interval%s", options.Snmptrap_interval)
			return
		}
	}
	opt = options
	return
}
func (this *Options) GetSnmptrap_address() string {
	return this.Snmptrap_address
}
func (this *Options) GetTrap_time_out() time.Duration {
	return this.Trap_time_out
}
func (this *Options) GetTrap_interval() time.Duration {
	return this.Trap_interval
}
