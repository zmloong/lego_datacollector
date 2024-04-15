package binary

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
		GetBinary_collectiontype() CollectionType
		GetBinary_max_collection_num() int
		GetBinary_readerbuf_size() int64
		GetBinary_server_addr() string
		GetBinary_server_port() int
		GetBinary_server_user() string
		GetBinary_server_password() string
		GetBinary_directory() string
		GetBinary_interval() string
		GetBinary_regularrules() string
		GetBinary_exec_onstart() bool
	}
	Options struct {
		core.ReaderOptions                       //继承基础配置
		Binary_collectiontype     CollectionType //采集类型
		Binary_max_collection_num int            //Binary 采集最大文件数
		Binary_readerbuf_size     int64          //Binary 读取文件缓存大小
		Binary_server_addr        string         //远程服务地址
		Binary_server_port        int            //远程服务 端口
		Binary_server_user        string         //远程服务账号
		Binary_server_password    string         //远程服务密码
		Binary_directory          string         //Binary 采集目录
		Binary_interval           string         //Binary 采集间隔时间
		Binary_regularrules       string         //Sftp 文件过滤正则规则
		Binary_exec_onstart       bool
	}
)

func newOptions(config map[string]interface{}) (opt IOptions, err error) {
	options := &Options{
		Binary_readerbuf_size:     4096,
		Binary_max_collection_num: 1000,
		Binary_regularrules:       "\\.(log|txt)$",
		Binary_interval:           "0 */1 * * * ?",
	}
	if config != nil {
		if err = mapstructure.Decode(config, options); err == nil {
			err = mapstructure.Decode(config, &options.ReaderOptions)
		}
	}
	opt = options
	return
}

func (this *Options) GetBinary_collectiontype() CollectionType {
	return this.Binary_collectiontype
}
func (this *Options) GetBinary_readerbuf_size() int64 {
	return this.Binary_readerbuf_size
}
func (this *Options) GetBinary_max_collection_num() int {
	return this.Binary_max_collection_num
}

func (this *Options) GetBinary_server_addr() string {
	return this.Binary_server_addr
}
func (this *Options) GetBinary_server_port() int {
	return this.Binary_server_port
}
func (this *Options) GetBinary_server_user() string {
	return this.Binary_server_user
}
func (this *Options) GetBinary_server_password() string {
	return this.Binary_server_password
}
func (this *Options) GetBinary_directory() string {
	return this.Binary_directory
}
func (this *Options) GetBinary_interval() string {
	return this.Binary_interval
}
func (this *Options) GetBinary_regularrules() string {
	return this.Binary_regularrules
}
func (this *Options) GetBinary_exec_onstart() bool {
	return this.Binary_exec_onstart
}
