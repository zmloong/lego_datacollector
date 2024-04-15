package metaer

import (
	"lego_datacollector/modules/datacollector/core"
	"lego_datacollector/sys/db"
	"sync"

	"github.com/liwei1dao/lego/sys/mgo"
	"github.com/liwei1dao/lego/sys/redis"
)

func NewMetaer(runner core.IRunner) (metaer core.IMetaer, err error) {
	metaer = &RedisMetaer{}
	if err = metaer.Init(runner); err != nil {
		return
	}
	return
}

type RedisMetaer struct {
	runner core.IRunner
	lock   sync.RWMutex
	mates  map[string]core.IMetaerData
}

func (this *RedisMetaer) Init(runner core.IRunner) (err error) {
	this.runner = runner
	this.mates = make(map[string]core.IMetaerData)
	return
}

//注册加载元数据
func (this *RedisMetaer) Read(meta core.IMetaerData) (err error) {
	this.lock.Lock()
	defer this.lock.Unlock()
	if err = db.ReadMetaData(this.runner.Name(), meta.GetName(), meta.GetMetae()); err != nil {
		if err == redis.RedisNil || err == mgo.MongodbNil {
			err = nil
		} else {
			this.runner.Errorf("RedisMetaer Read:%s err:%v", meta.GetName(), err)
		}
	}
	return
}

//注册加载元数据
func (this *RedisMetaer) Write(meta core.IMetaerData) (err error) {
	this.lock.Lock()
	defer this.lock.Unlock()
	if err = db.WriteMetaData(this.runner.Name(), meta.GetName(), meta.GetMetae()); err != nil {
		this.runner.Errorf("RedisMetaer Write:%s err:%v", meta.GetName(), err)
	}
	return
}

//采集器需要关闭时元数据这边需要保存处理
func (this *RedisMetaer) Close() (err error) {
	this.lock.Lock()
	defer this.lock.Unlock()
	for k, v := range this.mates {
		if err = this.Write(v); err != nil {
			this.runner.Errorf("Metaer MetaData:%s Write Fatal err:%v", k, err)
			return
		}
	}
	return
}
