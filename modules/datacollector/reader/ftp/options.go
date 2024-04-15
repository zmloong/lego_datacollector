package ftp

import (
	"fmt"
	"lego_datacollector/modules/datacollector/core"
	"regexp"

	"github.com/liwei1dao/lego/utils/mapstructure"
)

type (
	IOptions interface {
		core.IReaderOptions //继承基础配置
		GetFtp_server() string
		GetFtp_port() int32
		GetFtp_user() string
		GetFtp_password() string
		GetFtp_directory() string
		GetFtp_interval() string
		GetFtp_regularrules() string
		GetFtp_read_buffer_size() int32
		GetFtp_aax_collection_num() int
		GetFtp_exec_onstart() bool
	}
	Options struct {
		core.ReaderOptions            //继承基础配置
		Ftp_server             string //ftp 地址
		Ftp_port               int32  //ftp 访问端口
		Ftp_user               string //ftp 登录用户名
		Ftp_password           string //ftp 登录密码
		Ftp_directory          string //ftp 采集目录
		Ftp_interval           string //Ftp 采集间隔时间
		Ftp_regularrules       string //Ftp 文件过滤正则规则
		Ftp_read_buffer_size   int32  //Ftp 读取远程文件一次最多读取字节数
		Ftp_aax_collection_num int    //Ftp最大采集文件数 默认 1000
		Ftp_exec_onstart       bool   //Ftp 是否启动时查询一次
	}
)

func newOptions(config map[string]interface{}) (opt IOptions, err error) {
	options := &Options{
		Ftp_port:               21,
		Ftp_aax_collection_num: 1000,
		Ftp_read_buffer_size:   4096,
		Ftp_interval:           "0 */1 * * * ?", //采集定时执行 默认一分钟执行一次
		Ftp_regularrules:       "\\.(log|txt)$",
	}
	if config != nil {
		if err = mapstructure.Decode(config, options); err == nil {
			err = mapstructure.Decode(config, &options.ReaderOptions)
		}
	}
	if _, err = regexp.MatchString(options.Ftp_regularrules, "test.log"); err != nil {
		err = fmt.Errorf("newOptions Ftp_regularrules is err:%v", err)
		return
	}
	opt = options
	return
}

func (this *Options) GetFtp_server() string {
	return this.Ftp_server
}
func (this *Options) GetFtp_port() int32 {
	return this.Ftp_port
}
func (this *Options) GetFtp_user() string {
	return this.Ftp_user
}
func (this *Options) GetFtp_password() string {
	return this.Ftp_password
}
func (this *Options) GetFtp_directory() string {
	return this.Ftp_directory
}
func (this *Options) GetFtp_interval() string {
	return this.Ftp_interval
}
func (this *Options) GetFtp_regularrules() string {
	return this.Ftp_regularrules
}
func (this *Options) GetFtp_read_buffer_size() int32 {
	return this.Ftp_read_buffer_size
}

func (this *Options) GetFtp_aax_collection_num() int {
	return this.Ftp_aax_collection_num
}
func (this *Options) GetFtp_exec_onstart() bool {
	return this.Ftp_exec_onstart
}
