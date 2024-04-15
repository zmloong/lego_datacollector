package file

import (
	"errors"
	"lego_datacollector/modules/datacollector/core"

	"github.com/liwei1dao/lego/utils/mapstructure"
)

type (
	IOptions interface {
		core.ISenderOptions //继承基础配置
		GetOpenMaxfiles() int32
		GetSendfilePath() string
		GetKeyFileSenderTimestampKey() string
		GetKeyFilePartition() int
	}
	Options struct {
		core.SenderOptions              //继承基础配置
		File_send_path           string //ftp 地址
		File_send_max_open_files int32  //ftp 访问端口
		File_send_timestamp_key  string //ftp 访问端口
		File_partition           int    //ftp 访问端口
	}
)

func newOptions(config map[string]interface{}) (opt IOptions, err error) {
	options := &Options{
		File_send_max_open_files: 10,
	}
	if config != nil {
		if err = mapstructure.Decode(config, options); err == nil {
			err = mapstructure.Decode(config, &options.SenderOptions)
		}
		if err != nil {
			return
		}
	}
	if len(options.File_send_path) == 0 {
		err = errors.New("sender file Missing necessary configuration:File_send_path")
	}
	opt = options
	return
}

func (this *Options) GetOpenMaxfiles() int32 {
	return this.File_send_max_open_files
}

func (this *Options) GetSendfilePath() string {
	return this.File_send_path
}
func (this *Options) GetKeyFileSenderTimestampKey() string {
	return this.File_send_timestamp_key
}

func (this *Options) GetKeyFilePartition() int {
	return this.File_partition
}
