package db

import (
	"time"

	"github.com/liwei1dao/lego/sys/redis"
	"github.com/liwei1dao/lego/utils/mapstructure"
)

type Option func(*Options)
type Options struct {
	RedisType             redis.RedisType
	RedisUrl              string
	RedisDB               int
	RedisPassword         string
	MongodbUrl            string
	MongodbDatabase       string
	OnceInsertIdNum       int
	RInfoExpiration       int //RInfo 过期时间 单位 秒 默认 120
	RStatisticsExpiration int //RStatistics 过期时间 单位 秒 60*60*24*7 = 604800
	TimeOut               time.Duration
}

func SetRedisUrl(v string) Option {
	return func(o *Options) {
		o.RedisUrl = v
	}
}
func SetRedisDB(v int) Option {
	return func(o *Options) {
		o.RedisDB = v
	}
}
func SetRedisPassword(v string) Option {
	return func(o *Options) {
		o.RedisPassword = v
	}
}
func SetMongodbUrl(v string) Option {
	return func(o *Options) {
		o.MongodbUrl = v
	}
}

func SetMongodbDatabase(v string) Option {
	return func(o *Options) {
		o.MongodbDatabase = v
	}
}

func SetOnceInsertIdNum(v int) Option {
	return func(o *Options) {
		o.OnceInsertIdNum = v
	}
}

func SetTimeOut(v time.Duration) Option {
	return func(o *Options) {
		o.TimeOut = v
	}
}

func newOptions(config map[string]interface{}, opts ...Option) Options {
	options := Options{
		TimeOut: time.Second * 3,
	}
	if config != nil {
		mapstructure.Decode(config, &options)
	}
	for _, o := range opts {
		o(&options)
	}
	return options
}

func newOptionsByOption(opts ...Option) Options {
	options := Options{
		RInfoExpiration:       120,
		RStatisticsExpiration: 604800,
		TimeOut:               time.Second * 3,
	}
	for _, o := range opts {
		o(&options)
	}
	return options
}
