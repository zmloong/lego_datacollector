package sftp

import (
	"fmt"
	"lego_datacollector/modules/datacollector/core"
	"regexp"

	"github.com/liwei1dao/lego/utils/mapstructure"
)

type (
	IOptions interface {
		core.IReaderOptions
		GetSftp_server() string
		GetSftp_port() int
		GetSftp_user() string
		GetSftp_password() string
		GetSftp_directory() string
		GetSftp_interval() string
		GetSftp_regularrules() string
		GetSftp_read_buffer_size() int32
		GetSftp_aax_collection_num() int
		GetSftp_exec_onstart() bool
	}
	Options struct {
		core.ReaderOptions
		Sftp_server             string //Sftp 地址
		Sftp_port               int    //Sftp 访问端口
		Sftp_user               string //Sftp 登录用户名
		Sftp_password           string //Sftp 登录密码
		Sftp_directory          string //Sftp 采集目录
		Sftp_interval           string //Sftp 采集间隔时间
		Sftp_regularrules       string //Sftp 文件过滤正则规则
		Sftp_read_buffer_size   int32  //Sftp 读取远程文件一次最多读取字节数
		Sftp_aax_collection_num int    //Ftp最大采集文件数 默认 1000
		Sftp_exec_onstart       bool
	}
)

func newOptions(config map[string]interface{}) (opt IOptions, err error) {
	options := &Options{
		Sftp_port:               22,
		Sftp_interval:           "0 */1 * * * ?", //默认采集间隔1m
		Sftp_read_buffer_size:   4096,
		Sftp_aax_collection_num: 1000,
		Sftp_regularrules:       "\\.(log|txt)$",
	}
	if config != nil {
		if err = mapstructure.Decode(config, options); err == nil {
			err = mapstructure.Decode(config, &options.ReaderOptions)
		}
	}
	if _, err = regexp.MatchString(options.Sftp_regularrules, "test.log"); err != nil {
		err = fmt.Errorf("newOptions Ftp_regularrules is err:%v", err)
		return
	}
	opt = options
	return
}

func (this *Options) GetSftp_server() string {
	return this.Sftp_server
}
func (this *Options) GetSftp_port() int {
	return this.Sftp_port
}
func (this *Options) GetSftp_user() string {
	return this.Sftp_user
}
func (this *Options) GetSftp_password() string {
	return this.Sftp_password
}
func (this *Options) GetSftp_directory() string {
	return this.Sftp_directory
}
func (this *Options) GetSftp_interval() string {
	return this.Sftp_interval
}
func (this *Options) GetSftp_regularrules() string {
	return this.Sftp_regularrules
}
func (this *Options) GetSftp_read_buffer_size() int32 {
	return this.Sftp_read_buffer_size
}

func (this *Options) GetSftp_aax_collection_num() int {
	return this.Sftp_aax_collection_num
}
func (this *Options) GetSftp_exec_onstart() bool {
	return this.Sftp_exec_onstart
}
