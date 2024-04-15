package httpfetch

import (
	"encoding/json"
	"fmt"
	"lego_datacollector/modules/datacollector/core"
	"strings"
	"time"

	"github.com/mitchellh/mapstructure"
)

type (
	IOptions interface {
		core.IReaderOptions //继承基础配置
		GetHttpfetch_method() string
		GetHttpfetch_headers() string
		GetHttpfetch_headersMap() map[string]string
		GetHttpfetch_address() string
		GetHttpfetch_body() string
		GetHttpfetch_dialTimeout() int
		GetHttpfetch_respTimeout() int
		GetHttpfetch_interval() string
		GetHttpfetch_exec_onstart() bool
		GetDialTimeout() time.Duration
		GetRespTimeout() time.Duration
	}
	Options struct {
		core.ReaderOptions     //继承基础配置
		Httpfetch_method       string
		Httpfetch_headers      string
		Httpfetch_headersMap   map[string]string
		Httpfetch_address      string
		Httpfetch_body         string
		Httpfetch_dialTimeout  int
		Httpfetch_respTimeout  int
		Httpfetch_interval     string
		Httpfetch_exec_onstart bool
		DialTimeout            time.Duration
		RespTimeout            time.Duration
	}
)

func newOptions(config map[string]interface{}) (opt IOptions, err error) {
	time_out, err := time.ParseDuration("30s")
	if err != nil {
		return nil, err
	}
	options := &Options{
		Httpfetch_method:   "GET",
		DialTimeout:        time_out,
		RespTimeout:        time_out,
		Httpfetch_interval: "0 */1 * * * ?",
	}
	if config != nil {
		if err = mapstructure.Decode(config, options); err == nil {
			err = mapstructure.Decode(config, &options.ReaderOptions)
		}
	}
	if options.Httpfetch_dialTimeout >= 0 {
		options.DialTimeout = time.Duration(options.GetHttpfetch_dialTimeout()) * time.Second
	}
	if options.Httpfetch_respTimeout >= 0 {
		options.RespTimeout = time.Duration(options.GetHttpfetch_respTimeout()) * time.Second
	}
	if options.Httpfetch_headers != "" {
		if err = json.Unmarshal([]byte(options.Httpfetch_headers), &options.Httpfetch_headersMap); err != nil {
			return nil, fmt.Errorf("Reader httpfetch Httpfetch_headersMap json.Unmarshal:%v", err)
		}
	}
	if !strings.HasPrefix(options.Httpfetch_address, "http://") && !strings.HasPrefix(options.Httpfetch_address, "https://") {
		options.Httpfetch_address = "http://" + options.Httpfetch_address
	}
	opt = options
	return
}
func (this *Options) GetHttpfetch_method() string {
	return this.Httpfetch_method
}
func (this *Options) GetHttpfetch_headersMap() map[string]string {
	return this.Httpfetch_headersMap
}
func (this *Options) GetHttpfetch_headers() string {
	return this.Httpfetch_headers
}
func (this *Options) GetHttpfetch_address() string {
	return this.Httpfetch_address
}
func (this *Options) GetHttpfetch_body() string {
	return this.Httpfetch_body
}
func (this *Options) GetHttpfetch_dialTimeout() int {
	return this.Httpfetch_dialTimeout
}
func (this *Options) GetHttpfetch_respTimeout() int {
	return this.Httpfetch_respTimeout
}
func (this *Options) GetHttpfetch_interval() string {
	return this.Httpfetch_interval
}
func (this *Options) GetHttpfetch_exec_onstart() bool {
	return this.Httpfetch_exec_onstart
}

func (this *Options) GetDialTimeout() time.Duration {
	return this.DialTimeout
}
func (this *Options) GetRespTimeout() time.Duration {
	return this.RespTimeout
}
