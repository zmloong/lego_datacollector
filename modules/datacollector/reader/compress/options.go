package compress

import (
	"lego_datacollector/modules/datacollector/core"

	"github.com/liwei1dao/lego/utils/mapstructure"
)

const (
	FTPCollection CollectionType = iota //ftp压缩文件采集
)

type (
	CollectionType uint8
	IOptions       interface {
		core.IReaderOptions //继承基础配置
		GetCompress_collectiontype() CollectionType
		GetCompress_max_collection_size() int64
		GetCompress_max_collection_num() int
		GetCompress_server_addr() string
		GetCompress_server_port() int
		GetCompress_server_user() string
		GetCompress_server_password() string
		GetCompress_directory() string
		GetCompress_interval() string
		GetCompress_regularrules() string
		GetCompress_inter_regularrules() string
		GetCompress_xml_regularrules() string
		GetCompress_xml_root_node() string
		GetCompress_exec_onstart() bool
	}
	Options struct {
		core.ReaderOptions                          //继承基础配置
		Compress_collectiontype      CollectionType //采集类型
		Compress_max_collection_size int64          //Compress 采集最大文件尺寸
		Compress_max_collection_num  int            //Compress 采集最大文件数
		Compress_server_addr         string         //远程服务地址
		Compress_server_port         int            //远程服务 端口
		Compress_server_user         string         //远程服务账号
		Compress_server_password     string         //远程服务密码
		Compress_directory           string         //Compress 采集目录
		Compress_interval            string         //Compress 采集间隔时间
		Compress_regularrules        string         //Compress 文件过滤正则规则
		Compress_inter_regularrules  string         //Compress 里面文件过滤正则规则
		Compress_xml_regularrules    string         //Compress 匹配xml文件后缀名
		Compress_xml_root_node       string         //Compress xml文件解析指定数据提取根节点
		Compress_exec_onstart        bool
	}
)

func newOptions(config map[string]interface{}) (opt IOptions, err error) {
	options := &Options{
		Compress_regularrules: "\\.(tar.gz)$",
	}
	if config != nil {
		if err = mapstructure.Decode(config, options); err == nil {
			err = mapstructure.Decode(config, &options.ReaderOptions)
		}
	}
	opt = options
	return
}

func (this *Options) GetCompress_collectiontype() CollectionType {
	return this.Compress_collectiontype
}
func (this *Options) GetCompress_max_collection_size() int64 {
	return this.Compress_max_collection_size
}
func (this *Options) GetCompress_max_collection_num() int {
	return this.Compress_max_collection_num
}
func (this *Options) GetCompress_server_addr() string {
	return this.Compress_server_addr
}
func (this *Options) GetCompress_server_port() int {
	return this.Compress_server_port
}
func (this *Options) GetCompress_server_user() string {
	return this.Compress_server_user
}
func (this *Options) GetCompress_server_password() string {
	return this.Compress_server_password
}
func (this *Options) GetCompress_directory() string {
	return this.Compress_directory
}
func (this *Options) GetCompress_interval() string {
	return this.Compress_interval
}
func (this *Options) GetCompress_regularrules() string {
	return this.Compress_regularrules
}
func (this *Options) GetCompress_inter_regularrules() string {
	return this.Compress_inter_regularrules
}
func (this *Options) GetCompress_xml_regularrules() string {
	return this.Compress_xml_regularrules
}
func (this *Options) GetCompress_xml_root_node() string {
	return this.Compress_xml_root_node
}
func (this *Options) GetCompress_exec_onstart() bool {
	return this.Compress_exec_onstart
}
