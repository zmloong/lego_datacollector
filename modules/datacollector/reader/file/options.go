package file

import (
	"lego_datacollector/modules/datacollector/core"

	"github.com/liwei1dao/lego/utils/mapstructure"
)

type (
	IOptions interface {
		core.IReaderOptions //继承基础配置
		GetFile_path() string
		GetFile_readbuff_size() int
		GetFile_scan_initial() string
		GetFile_exec_onstart() bool
	}
	Options struct {
		core.ReaderOptions        //继承基础配置
		File_path          string //文件路径
		File_readbuff_size int    //读取缓存大小
		File_scan_initial  string //扫描间隔 单位秒
		File_exec_onstart  bool   //Ftp 是否启动时查询一次
	}
)

func newOptions(config map[string]interface{}) (opt IOptions, err error) {
	options := &Options{
		File_readbuff_size: 4096,
		File_scan_initial:  "0 */1 * * * ?", //采集定时执行 默认一分钟执行一次
	}
	if config != nil {
		if err = mapstructure.Decode(config, options); err == nil {
			err = mapstructure.Decode(config, &options.ReaderOptions)
		}
	}
	opt = options
	return
}

func (this *Options) GetFile_path() string {
	return this.File_path
}

func (this *Options) GetFile_readbuff_size() int {
	return this.File_readbuff_size
}
func (this *Options) GetFile_scan_initial() string {
	return this.File_scan_initial
}

func (this *Options) GetFile_exec_onstart() bool {
	return this.File_exec_onstart
}
