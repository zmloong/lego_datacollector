package xlsx

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
		core.ReaderOptions                      //继承基础配置
		Xlsx_collectiontype      CollectionType //采集类型
		Xlsx_max_collection_size int64          //Xls 采集最大文件尺寸
		Xlsx_max_collection_num  int            //Xls 采集最大文件数
		Xlsx_server_addr         string         //远程服务地址
		Xlsx_server_port         int            //远程服务 端口
		Xlsx_server_user         string         //远程服务账号
		Xlsx_server_password     string         //远程服务密码
		Xlsx_key_line            int            //键值行
		Xlsx_data_line           int            //数据查询起始行号
		Xlsx_directory           string         //Xls 采集目录
		Xlsx_interval            string         //Xls 采集间隔时间
		Xlsx_regularrules        string         //Sftp 文件过滤正则规则
		Xlsx_exec_onstart        bool
	}
)

func newOptions(config map[string]interface{}) (opt IOptions, err error) {
	options := &Options{
		Xlsx_key_line:     0,
		Xlsx_data_line:    1,
		Xlsx_regularrules: "\\.(xlsx)$",
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
	return this.Xlsx_collectiontype
}
func (this *Options) GetXls_max_collection_size() int64 {
	return this.Xlsx_max_collection_size
}
func (this *Options) GetXls_max_collection_num() int {
	return this.Xlsx_max_collection_num
}
func (this *Options) GetXls_server_addr() string {
	return this.Xlsx_server_addr
}
func (this *Options) GetXls_server_port() int {
	return this.Xlsx_server_port
}
func (this *Options) GetXls_server_user() string {
	return this.Xlsx_server_user
}
func (this *Options) GetXls_server_password() string {
	return this.Xlsx_server_password
}
func (this *Options) GetXls_key_line() int {
	return this.Xlsx_key_line
}

func (this *Options) GetXls_data_line() int {
	return this.Xlsx_data_line
}
func (this *Options) GetXls_directory() string {
	return this.Xlsx_directory
}
func (this *Options) GetXls_interval() string {
	return this.Xlsx_interval
}
func (this *Options) GetXls_regularrules() string {
	return this.Xlsx_regularrules
}
func (this *Options) GetXls_exec_onstart() bool {
	return this.Xlsx_exec_onstart
}
