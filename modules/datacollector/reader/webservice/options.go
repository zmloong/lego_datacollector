package webservice

import (
	"lego_datacollector/modules/datacollector/core"
	"strings"
	"time"

	"github.com/mitchellh/mapstructure"
)

type (
	IOptions interface {
		core.IReaderOptions //继承基础配置
		GetWebservice_address() string
		GetWebservice_reqbody() string
		GetWebservice_xml_body_node() string
		GetWebservice_dialTimeout() int
		GetWebservice_respTimeout() int
		GetWebservice_interval() string
		GetWebservice_exec_onstart() bool
		GetDialTimeout() time.Duration
		GetRespTimeout() time.Duration
	}
	Options struct {
		core.ReaderOptions       //继承基础配置
		Webservice_address       string
		Webservice_reqbody       string
		Webservice_xml_body_node string
		Webservice_dialTimeout   int
		Webservice_respTimeout   int
		Webservice_interval      string
		Webservice_exec_onstart  bool
		DialTimeout              time.Duration
		RespTimeout              time.Duration
	}
)

func newOptions(config map[string]interface{}) (opt IOptions, err error) {
	time_out, err := time.ParseDuration("30s")
	if err != nil {
		return nil, err
	}
	options := &Options{
		RespTimeout:         time_out,
		DialTimeout:         time_out,
		Webservice_interval: "0 */1 * * * ?",
	}
	if config != nil {
		if err = mapstructure.Decode(config, options); err == nil {
			err = mapstructure.Decode(config, &options.ReaderOptions)
		}
	}
	if options.Webservice_respTimeout >= 0 {
		options.RespTimeout = time.Duration(options.GetWebservice_respTimeout()) * time.Second
	}
	if options.Webservice_dialTimeout >= 0 {
		options.RespTimeout = time.Duration(options.GetWebservice_dialTimeout()) * time.Second
	}
	if !strings.HasPrefix(options.Webservice_address, "http://") {
		options.Webservice_address = "http://" + options.Webservice_address
	}
	opt = options
	return
}
func (this *Options) GetWebservice_address() string {
	return this.Webservice_address
}
func (this *Options) GetWebservice_reqbody() string {
	return this.Webservice_reqbody
}
func (this *Options) GetWebservice_xml_body_node() string {
	return this.Webservice_xml_body_node
}
func (this *Options) GetWebservice_dialTimeout() int {
	return this.Webservice_dialTimeout
}
func (this *Options) GetWebservice_respTimeout() int {
	return this.Webservice_respTimeout
}
func (this *Options) GetWebservice_interval() string {
	return this.Webservice_interval
}
func (this *Options) GetWebservice_exec_onstart() bool {
	return this.Webservice_exec_onstart
}
func (this *Options) GetDialTimeout() time.Duration {
	return this.DialTimeout
}
func (this *Options) GetRespTimeout() time.Duration {
	return this.RespTimeout
}
