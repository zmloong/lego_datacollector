package xml

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
		GetXml_collectiontype() CollectionType
		GetXml_max_collection_size() int64
		GetXml_max_collection_num() int
		GetXml_server_addr() string
		GetXml_server_port() int
		GetXml_server_user() string
		GetXml_server_password() string
		GetXml_node_root() string
		GetXml_node_data() string
		GetXml_directory() string
		GetXml_interval() string
		GetXml_regularrules() string
		GetXml_exec_onstart() bool
	}
	Options struct {
		core.ReaderOptions                     //继承基础配置
		Xml_collectiontype      CollectionType //采集类型
		Xml_max_collection_size int64          //Xml 采集最大文件尺寸
		Xml_max_collection_num  int            //Xml 采集最大文件数
		Xml_server_addr         string         //远程服务地址
		Xml_server_port         int            //远程服务 端口
		Xml_server_user         string         //远程服务账号
		Xml_server_password     string         //远程服务密码
		Xml_node_root           string         //root节点
		Xml_node_data           string         //data节点
		Xml_directory           string         //Xml 采集目录
		Xml_interval            string         //Xml 采集间隔时间
		Xml_regularrules        string         //Sftp 文件过滤正则规则
		Xml_exec_onstart        bool
	}
)

func newOptions(config map[string]interface{}) (opt IOptions, err error) {
	options := &Options{
		Xml_regularrules: "\\.(xml)$",
	}
	if config != nil {
		if err = mapstructure.Decode(config, options); err == nil {
			err = mapstructure.Decode(config, &options.ReaderOptions)
		}
	}
	opt = options
	return
}

func (this *Options) GetXml_collectiontype() CollectionType {
	return this.Xml_collectiontype
}
func (this *Options) GetXml_max_collection_size() int64 {
	return this.Xml_max_collection_size
}
func (this *Options) GetXml_max_collection_num() int {
	return this.Xml_max_collection_num
}
func (this *Options) GetXml_server_addr() string {
	return this.Xml_server_addr
}
func (this *Options) GetXml_server_port() int {
	return this.Xml_server_port
}
func (this *Options) GetXml_server_user() string {
	return this.Xml_server_user
}
func (this *Options) GetXml_server_password() string {
	return this.Xml_server_password
}
func (this *Options) GetXml_node_root() string {
	return this.Xml_node_root
}
func (this *Options) GetXml_node_data() string {
	return this.Xml_node_data
}
func (this *Options) GetXml_directory() string {
	return this.Xml_directory
}
func (this *Options) GetXml_interval() string {
	return this.Xml_interval
}
func (this *Options) GetXml_regularrules() string {
	return this.Xml_regularrules
}
func (this *Options) GetXml_exec_onstart() bool {
	return this.Xml_exec_onstart
}
