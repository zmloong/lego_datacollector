package xls

import (
	"lego_datacollector/modules/datacollector/core"

	"github.com/liwei1dao/lego/utils/mapstructure"
)

const (
	LoaclCollection CollectionType = iota //本地文件采集
	FTPCollection                         //ftp文件采集
	SFTPCollection                        //sftp文件采集
)

type (
	CollectionType uint8
	IOptions       interface {
		core.IReaderOptions //继承基础配置
		GetXls_collectiontype() CollectionType
		GetXls_max_collection_size() int64
		GetXls_max_collection_num() int
		GetXls_server_addr() string
		GetXls_server_port() int
		GetXls_server_user() string
		GetXls_server_password() string
		GetXls_key_line() int
		GetXls_data_line() int
		GetXls_directory() string
		GetXls_interval() string
		GetXls_regularrules() string
		GetXls_exec_onstart() bool
	}
	Options struct {
		core.ReaderOptions                     //继承基础配置
		Xls_collectiontype      CollectionType //采集类型
		Xls_max_collection_size int64          //Xls 采集最大文件尺寸
		Xls_max_collection_num  int            //Xls 采集最大文件数
		Xls_server_addr         string         //远程服务地址
		Xls_server_port         int            //远程服务 端口
		Xls_server_user         string         //远程服务账号
		Xls_server_password     string         //远程服务密码
		Xls_key_line            int            //键值行
		Xls_data_line           int            //数据查询起始行号
		Xls_directory           string         //Xls 采集目录
		Xls_interval            string         //Xls 采集间隔时间
		Xls_regularrules        string         //Sftp 文件过滤正则规则
		Xls_exec_onstart        bool
	}
)

func newOptions(config map[string]interface{}) (opt IOptions, err error) {
	options := &Options{
		Xls_key_line:     0,
		Xls_data_line:    1,
		Xls_regularrules: "\\.(xls)$",
	}
	if config != nil {
		if err = mapstructure.Decode(config, options); err == nil {
			err = mapstructure.Decode(config, &options.ReaderOptions)
		}
	}
	opt = options
	return
}

func (this *Options) GetXls_collectiontype() CollectionType {
	return this.Xls_collectiontype
}
func (this *Options) GetXls_max_collection_size() int64 {
	return this.Xls_max_collection_size
}
func (this *Options) GetXls_max_collection_num() int {
	return this.Xls_max_collection_num
}
func (this *Options) GetXls_server_addr() string {
	return this.Xls_server_addr
}
func (this *Options) GetXls_server_port() int {
	return this.Xls_server_port
}
func (this *Options) GetXls_server_user() string {
	return this.Xls_server_user
}
func (this *Options) GetXls_server_password() string {
	return this.Xls_server_password
}
func (this *Options) GetXls_key_line() int {
	return this.Xls_key_line
}

func (this *Options) GetXls_data_line() int {
	return this.Xls_data_line
}
func (this *Options) GetXls_directory() string {
	return this.Xls_directory
}
func (this *Options) GetXls_interval() string {
	return this.Xls_interval
}
func (this *Options) GetXls_regularrules() string {
	return this.Xls_regularrules
}
func (this *Options) GetXls_exec_onstart() bool {
	return this.Xls_exec_onstart
}
