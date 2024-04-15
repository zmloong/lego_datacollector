package hdfs

import (
	"fmt"
	"lego_datacollector/modules/datacollector/core"
	"regexp"
	"strings"

	"github.com/liwei1dao/lego/utils/mapstructure"
)

type (
	IOptions interface {
		core.IReaderOptions
		GetHdfs_addr() string
		GetHdfs_ip() string
		GetHdfs_user() string
		GetHdfs_password() string
		GetHdfs_directory() string
		GetHdfs_interval() string
		GetHdfs_regularrules() string
		GetHdfs_read_buffer_size() int32
		GetHdfs_max_collection_num() int
		GetHdfs_exec_onstart() bool
		GetHdfs_Kerberos_Enable() bool
		GetHdfs_Kerberos_Realm() string
		GetHdfs_Kerberos_ServiceName() string
		GetHdfs_Kerberos_Username() string
		GetHdfs_Kerberos_KeyTabPath() string
		GetHdfs_Kerberos_KerberosConfigPath() string
		GetHdfs_Kerberos_HadoopConfPath() string
	}
	Options struct {
		core.ReaderOptions
		Hdfs_addr               string //Hdfs 地址
		Hdfs_ip                 string //Hdfs 源ip地址
		Hdfs_user               string //Hdfs 登录用户名
		Hdfs_password           string //Hdfs 登录密码
		Hdfs_directory          string //Hdfs 采集目录
		Hdfs_interval           string //Hdfs 采集间隔时间
		Hdfs_regularrules       string //Hdfs 文件过滤正则规则
		Hdfs_read_buffer_size   int32  //Hdfs 读取远程文件一次最多读取字节数
		Hdfs_max_collection_num int    //Hdfs 最大采集文件数 默认 1000
		Hdfs_exec_onstart       bool

		Hdfs_Kerberos_Enable             bool   //是否开启 Kerberos 认证
		Hdfs_Kerberos_Realm              string //Kerberos 认证 Realm 字段
		Hdfs_Kerberos_ServiceName        string //Kerberos 认证 ServiceName 字段
		Hdfs_Kerberos_Username           string //Kerberos 认证 Username 字段
		Hdfs_Kerberos_KeyTabPath         string //Kerberos 认证 KeyTabPath 文件路径
		Hdfs_Kerberos_KerberosConfigPath string //Kerberos 认证 KerberosConfigPath 文件路径
		Hdfs_Kerberos_HadoopConfPath     string //Kerberos 认证 core-site.xml和 hdfs-site.xml文件路径
	}
)

func newOptions(config map[string]interface{}) (opt IOptions, err error) {
	options := &Options{
		Hdfs_interval:           "0 */1 * * * ?", //默认采集间隔1m
		Hdfs_read_buffer_size:   4096,
		Hdfs_max_collection_num: 1000,
		Hdfs_regularrules:       "\\.(log|txt)$",
	}
	if config != nil {
		if err = mapstructure.Decode(config, options); err == nil {
			err = mapstructure.Decode(config, &options.ReaderOptions)
		}
	}
	if _, err = regexp.MatchString(options.Hdfs_regularrules, "test.log"); err != nil {
		err = fmt.Errorf("newOptions Ftp_regularrules is err:%v", err)
		return
	}
	options.Hdfs_ip = strings.Split(options.GetHdfs_addr(), ":")[0]
	opt = options
	return
}

func (this *Options) GetHdfs_addr() string {
	return this.Hdfs_addr
}
func (this *Options) GetHdfs_ip() string {
	return this.Hdfs_ip
}
func (this *Options) GetHdfs_user() string {
	return this.Hdfs_user
}
func (this *Options) GetHdfs_password() string {
	return this.Hdfs_password
}
func (this *Options) GetHdfs_directory() string {
	return this.Hdfs_directory
}
func (this *Options) GetHdfs_interval() string {
	return this.Hdfs_interval
}
func (this *Options) GetHdfs_regularrules() string {
	return this.Hdfs_regularrules
}
func (this *Options) GetHdfs_read_buffer_size() int32 {
	return this.Hdfs_read_buffer_size
}

func (this *Options) GetHdfs_max_collection_num() int {
	return this.Hdfs_max_collection_num
}
func (this *Options) GetHdfs_exec_onstart() bool {
	return this.Hdfs_exec_onstart
}

func (this *Options) GetHdfs_Kerberos_Enable() bool {
	return this.Hdfs_Kerberos_Enable
}
func (this *Options) GetHdfs_Kerberos_Realm() string {
	return this.Hdfs_Kerberos_Realm
}
func (this *Options) GetHdfs_Kerberos_ServiceName() string {
	return this.Hdfs_Kerberos_ServiceName
}
func (this *Options) GetHdfs_Kerberos_Username() string {
	return this.Hdfs_Kerberos_Username
}
func (this *Options) GetHdfs_Kerberos_KeyTabPath() string {
	return this.Hdfs_Kerberos_KeyTabPath
}
func (this *Options) GetHdfs_Kerberos_KerberosConfigPath() string {
	return this.Hdfs_Kerberos_KerberosConfigPath
}
func (this *Options) GetHdfs_Kerberos_HadoopConfPath() string {
	return this.Hdfs_Kerberos_HadoopConfPath
}
