package mysql

import (
	"fmt"
	"lego_datacollector/modules/datacollector/core"

	"github.com/mitchellh/mapstructure"
)

type (
	IOptions interface {
		core.ISenderOptions //继承基础配置
		GetMysql_datasource() string
		GetMysql_database() string
		GetMysql_table() string
		GetMysql_mapping() map[string]string
	}
	Options struct {
		core.SenderOptions                   //继承基础配置
		Mysql_datasource   string            //mysql 连接字段
		Mysql_database     string            //数据库
		Mysql_table        string            //数据存储表
		Mysql_mapping      map[string]string //数据库同步映射表
	}
)

func newOptions(config map[string]interface{}) (opt IOptions, err error) {
	options := &Options{}
	if config != nil {
		if err = mapstructure.Decode(config, options); err == nil {
			err = mapstructure.Decode(config, &options.SenderOptions)
		}
	}
	if len(options.Mysql_mapping) == 0 {
		err = fmt.Errorf("options 配置错误 映射表不符合规范")
		return
	}
	opt = options
	return
}

func (this *Options) GetMysql_datasource() string {
	return this.Mysql_datasource
}
func (this *Options) GetMysql_database() string {
	return this.Mysql_database
}
func (this *Options) GetMysql_table() string {
	return this.Mysql_table
}
func (this *Options) GetMysql_mapping() map[string]string {
	return this.Mysql_mapping
}
