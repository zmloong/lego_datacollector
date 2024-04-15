package hive

import (
	"lego_datacollector/modules/datacollector/core"

	"github.com/mitchellh/mapstructure"
)

const (
	CollectionSender SenderHiveType = iota
	SynchronizeSender
)

type (
	SenderHiveType int8
	IOptions       interface {
		core.ISenderOptions //继承基础配置
		GetHdfs_addr() string
		GetHdfs_user() string
		GetHdfs_path() string
		GetHdfs_size() int
		GetHdfs_dir_intervel() string
		GetHdfs_timeout() int

		GetHdfs_Kerberos_Enable() bool
		GetHdfs_Kerberos_Realm() string
		GetHdfs_Kerberos_ServiceName() string
		GetHdfs_Kerberos_Username() string
		GetHdfs_Kerberos_KeyTabPath() string
		GetHdfs_Kerberos_KerberosConfigPath() string
		GetHdfs_Kerberos_HadoopConfPath() string

		GetHive_addr() string
		GetHive_username() string
		GetHive_password() string
		GetHive_auth() string
		GetHive_sendertype() SenderHiveType
		GetHive_dbname() string
		GetHive_tableName() string
		GetHive_mapping() map[string]string
		GetHive_column_separator() string
		GetHive_json() bool
	}
	Options struct {
		core.SenderOptions                      //继承基础配置
		Hdfs_addr                        string `json:"hdfs_addr"`         //
		Hdfs_user                        string `json:"hdfs_user"`         //
		Hdfs_path                        string `json:"hdfs_path"`         //
		Hdfs_size                        int    `json:"hdfs_size"`         //
		Hdfs_dir_intervel                string `json:"hdfs_dir_intervel"` //
		Hdfs_timeout                     int    `json:"hdfs_timeout"`      //
		Hdfs_Kerberos_Enable             bool   //是否开启 Kerberos 认证
		Hdfs_Kerberos_Realm              string //Kerberos 认证 Realm 字段
		Hdfs_Kerberos_ServiceName        string //Kerberos 认证 ServiceName 字段
		Hdfs_Kerberos_Username           string //Kerberos 认证 Username 字段
		Hdfs_Kerberos_KeyTabPath         string //Kerberos 认证 KeyTabPath 文件路径
		Hdfs_Kerberos_KerberosConfigPath string //Kerberos 认证 KerberosConfigPath 文件路径
		Hdfs_Kerberos_HadoopConfPath     string //Kerberos 认证 core-site.xml和 hdfs-site.xml文件路径

		Hive_addr             string
		Hive_username         string
		Hive_password         string
		Hive_auth             string
		Hive_sendertype       SenderHiveType
		Hive_dbname           string
		Hive_tableName        string
		Hive_mapping          map[string]string
		Hive_column_separator string
		Hive_json             bool
	}
)

func newOptions(config map[string]interface{}) (opt IOptions, err error) {
	options := &Options{
		Hdfs_timeout:          15,
		Hdfs_size:             134217728,
		Hive_sendertype:       CollectionSender,
		Hive_column_separator: "|",
		Hive_dbname:           "default",
	}
	if config != nil {
		if err = mapstructure.Decode(config, options); err == nil {
			err = mapstructure.Decode(config, &options.SenderOptions)
		}
	}
	opt = options
	return
}
func (this *Options) GetHdfs_addr() string {
	return this.Hdfs_addr
}
func (this *Options) GetHdfs_user() string {
	return this.Hdfs_user
}
func (this *Options) GetHdfs_path() string {
	return this.Hdfs_path
}
func (this *Options) GetHdfs_size() int {
	return this.Hdfs_size
}
func (this *Options) GetHdfs_timeout() int {
	return this.Hdfs_timeout
}
func (this *Options) GetHdfs_dir_intervel() string {
	return this.Hdfs_dir_intervel
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

func (this *Options) GetHive_sendertype() SenderHiveType {
	return this.Hive_sendertype
}
func (this *Options) GetHive_dbname() string {
	return this.Hive_dbname
}
func (this *Options) GetHive_tableName() string {
	return this.Hive_tableName
}
func (this *Options) GetHive_mapping() map[string]string {
	return this.Hive_mapping
}
func (this *Options) GetHive_addr() string {
	return this.Hive_addr
}
func (this *Options) GetHive_username() string {
	return this.Hive_username
}
func (this *Options) GetHive_password() string {
	return this.Hive_password
}
func (this *Options) GetHive_auth() string {
	return this.Hive_auth
}
func (this *Options) GetHive_column_separator() string {
	return this.Hive_column_separator
}
func (this *Options) GetHive_json() bool {
	return this.Hive_json
}
