package gmsm2

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
		GetGmsm2_collectiontype() CollectionType
		GetGmsm2_max_collection_size() int64
		GetGmsm2_max_collection_num() int
		GetGmsm2_prikey() string
		GetGmsm2_prikey_pwd() string
		GetGmsm2_server_addr() string
		GetGmsm2_server_port() int
		GetGmsm2_server_user() string
		GetGmsm2_server_password() string
		GetGmsm2_node_root() string
		GetGmsm2_node_data() string
		GetGmsm2_directory() string
		GetGmsm2_interval() string
		GetGmsm2_regularrules() string
		GetGmsm2_exec_onstart() bool
	}
	Options struct {
		core.ReaderOptions                       //继承基础配置
		Gmsm2_collectiontype      CollectionType //采集类型
		Gmsm2_max_collection_size int64          //Gmsm2 采集最大文件尺寸
		Gmsm2_max_collection_num  int            //Gmsm2 采集最大文件数
		Gmsm2_prikey              string         //国密私钥地址
		Gmsm2_prikey_pwd          string         //私钥密码
		Gmsm2_server_addr         string         //远程服务地址
		Gmsm2_server_port         int            //远程服务 端口
		Gmsm2_server_user         string         //远程服务账号
		Gmsm2_server_password     string         //远程服务密码
		Gmsm2_node_root           string         //root节点
		Gmsm2_node_data           string         //data节点
		Gmsm2_directory           string         //Gmsm2 采集目录
		Gmsm2_interval            string         //Gmsm2 采集间隔时间
		Gmsm2_regularrules        string         //Sftp 文件过滤正则规则
		Gmsm2_exec_onstart        bool
	}
)

func newOptions(config map[string]interface{}) (opt IOptions, err error) {
	options := &Options{
		Gmsm2_regularrules: "\\.(log|txt)$",
	}
	if config != nil {
		if err = mapstructure.Decode(config, options); err == nil {
			err = mapstructure.Decode(config, &options.ReaderOptions)
		}
	}
	opt = options
	return
}

func (this *Options) GetGmsm2_collectiontype() CollectionType {
	return this.Gmsm2_collectiontype
}
func (this *Options) GetGmsm2_max_collection_size() int64 {
	return this.Gmsm2_max_collection_size
}
func (this *Options) GetGmsm2_max_collection_num() int {
	return this.Gmsm2_max_collection_num
}

func (this *Options) GetGmsm2_prikey() string {
	return this.Gmsm2_prikey
}

func (this *Options) GetGmsm2_prikey_pwd() string {
	return this.Gmsm2_prikey_pwd
}

func (this *Options) GetGmsm2_server_addr() string {
	return this.Gmsm2_server_addr
}
func (this *Options) GetGmsm2_server_port() int {
	return this.Gmsm2_server_port
}
func (this *Options) GetGmsm2_server_user() string {
	return this.Gmsm2_server_user
}
func (this *Options) GetGmsm2_server_password() string {
	return this.Gmsm2_server_password
}
func (this *Options) GetGmsm2_node_root() string {
	return this.Gmsm2_node_root
}
func (this *Options) GetGmsm2_node_data() string {
	return this.Gmsm2_node_data
}
func (this *Options) GetGmsm2_directory() string {
	return this.Gmsm2_directory
}
func (this *Options) GetGmsm2_interval() string {
	return this.Gmsm2_interval
}
func (this *Options) GetGmsm2_regularrules() string {
	return this.Gmsm2_regularrules
}
func (this *Options) GetGmsm2_exec_onstart() bool {
	return this.Gmsm2_exec_onstart
}
