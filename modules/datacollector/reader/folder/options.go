package folder

import (
	"fmt"
	"lego_datacollector/modules/datacollector/core"
	"regexp"

	"github.com/liwei1dao/lego/utils/mapstructure"
)

type (
	IOptions interface {
		core.IReaderOptions                //继承基础配置
		GetFolder_directory() string       //获取采集目录
		GetFolder_regular_filter() string  //查询正则过滤
		GetFolder_readerbuf_size() int     //读取缓存大小
		GetFolder_scan_initial() string    //文件采集单次读取量
		GetFolder_max_collection_num() int //最大采集文件数 默认 1000
		GetFolder_exec_onstart() bool        //Ftp 是否启动时查询一次
	}
	Options struct {
		core.ReaderOptions               //继承基础配置
		Folder_directory          string //采集魔窟
		Folder_regular_filter     string //正则过滤 默认 *
		Folder_readerbuf_size     int    //读取缓存大小
		Folder_scan_initial       string //扫描间隔 corn 表达式
		Folder_max_collection_num int    //最大采集文件数 默认 1000
		Folder_exec_onstart         bool   //Ftp 是否启动时查询一次
	}
)

func newOptions(config map[string]interface{}) (opt IOptions, err error) {
	options := &Options{
		Folder_regular_filter:     "\\.(log|txt)$",
		Folder_readerbuf_size:     4096,
		Folder_scan_initial:       "0 */1 * * * ?",
		Folder_max_collection_num: 1000,
	}
	if config != nil {
		if err = mapstructure.Decode(config, options); err == nil {
			err = mapstructure.Decode(config, &options.ReaderOptions)
		}
	}
	if len(options.Folder_directory) == 0 {
		err = fmt.Errorf("newOptions Folder_directory is empty")
		return
	}
	if _, err = regexp.MatchString(options.Folder_regular_filter, "test.log"); err != nil {
		err = fmt.Errorf("newOptions Folder_regular_filter is err:%v", err)
		return
	}
	opt = options
	return
}
func (this *Options) GetFolder_directory() string {
	return this.Folder_directory
}

func (this *Options) GetFolder_regular_filter() string {
	return this.Folder_regular_filter
}

func (this *Options) GetFolder_readerbuf_size() int {
	return this.Folder_readerbuf_size
}

func (this *Options) GetFolder_scan_initial() string {
	return this.Folder_scan_initial
}

func (this *Options) GetFolder_max_collection_num() int {
	return this.Folder_max_collection_num
}

func (this *Options) GetFolder_exec_onstart() bool {
	return this.Folder_exec_onstart
}
