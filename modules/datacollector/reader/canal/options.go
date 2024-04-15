package canal

import (
	"lego_datacollector/modules/datacollector/core"

	"github.com/liwei1dao/lego/utils/mapstructure"
)

type (
	IOptions interface {
		core.IReaderOptions //继承基础配置
		GetCanal_Addr() string
		GetCanal_Port() int
		GetCanal_User() string
		GetCanal_Password() string
		GetCanal_Destination() string
		GetCanal_Filter() string
	}
	Options struct {
		core.ReaderOptions //继承基础配置
		Canal_addr         string
		Canal_port         int
		Canal_user         string
		Canal_password     string
		Canal_destination  string
		Canal_filter       string
	}
)

func newOptions(config map[string]interface{}) (opt IOptions, err error) {
	options := &Options{
		Canal_filter: ".*",
	}
	if config != nil {
		if err = mapstructure.Decode(config, options); err == nil {
			err = mapstructure.Decode(config, &options.ReaderOptions)
		}
	}
	opt = options
	return
}

func (this *Options) GetCanal_Addr() string {
	return this.Canal_addr
}

func (this *Options) GetCanal_Port() int {
	return this.Canal_port
}
func (this *Options) GetCanal_User() string {
	return this.Canal_user
}
func (this *Options) GetCanal_Password() string {
	return this.Canal_password
}

func (this *Options) GetCanal_Destination() string {
	return this.Canal_destination
}

// https://github.com/alibaba/canal/wiki/AdminGuide
//mysql 数据解析关注的表，Perl正则表达式.
//
//多个正则之间以逗号(,)分隔，转义符需要双斜杠(\\)
//
//常见例子：
//
//  1.  所有表：.*   or  .*\\..*
//	2.  canal schema下所有表： canal\\..*
//	3.  canal下的以canal打头的表：canal\\.canal.*
//	4.  canal schema下的一张表：canal\\.test1
//  5.  多个规则组合使用：canal\\..*,mysql.test1,mysql.test2 (逗号分隔)
func (this *Options) GetCanal_Filter() string {
	return this.Canal_filter
}
