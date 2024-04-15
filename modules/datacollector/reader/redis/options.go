package redis

import (
	"lego_datacollector/modules/datacollector/core"

	"github.com/liwei1dao/lego/sys/redis"
	"github.com/liwei1dao/lego/utils/mapstructure"
)

type (
	MyCollectionType uint8
	IOptions         interface {
		core.IReaderOptions //继承基础配置
		GetRedis_type() redis.RedisType
		GetRedis_url() []string
		GetRedis_db() int
		GetRedis_password() string
		GetRedis_pattern() string
		GetRedis_intervel() string
		GetRedis_exec_onstart() bool
		GetRedis_autoClose() bool
	}
	Options struct {
		core.ReaderOptions //继承基础配置
		Redis_type         redis.RedisType
		Redis_url          []string
		Redis_db           int
		Redis_password     string
		Redis_pattern      string
		Redis_intervel     string //定期扫描
		Redis_exec_onstart bool   //mysql 是否启动时查询一次
		Redis_autoClose    bool   //是否自动关闭
	}
)

func newOptions(config map[string]interface{}) (opt IOptions, err error) {
	options := &Options{
		Redis_pattern: "*",
	}
	if config != nil {
		if err = mapstructure.Decode(config, options); err == nil {
			err = mapstructure.Decode(config, &options.ReaderOptions)
		}
	}
	opt = options
	return
}

func (this *Options) GetRedis_type() redis.RedisType {
	return this.Redis_type
}

func (this *Options) GetRedis_url() []string {
	return this.Redis_url
}
func (this *Options) GetRedis_db() int {
	return this.Redis_db
}
func (this *Options) GetRedis_password() string {
	return this.Redis_password
}
func (this *Options) GetRedis_pattern() string {
	return this.Redis_pattern
}

func (this *Options) GetRedis_intervel() string {
	return this.Redis_intervel
}
func (this *Options) GetRedis_exec_onstart() bool {
	return this.Redis_exec_onstart
}

func (this *Options) GetRedis_autoClose() bool {
	return this.Redis_autoClose
}
