package db

import (
	"fmt"
	"lego_datacollector/comm"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/liwei1dao/lego/sys/log"
	lgredis "github.com/liwei1dao/lego/sys/redis"
)

func newSys(options Options) (sys *DB, err error) {
	sys = &DB{options: options}
	if sys.redis, err = lgredis.NewSys(
		lgredis.SetRedisType(options.RedisType),
		lgredis.SetRedis_Single_Addr(options.RedisUrl),
		lgredis.SetRedis_Single_DB(options.RedisDB),
		lgredis.SetRedis_Single_Password(options.RedisPassword),
		lgredis.Redis_Cluster_Addr([]string{options.RedisUrl}),
		lgredis.SetRedis_Cluster_Password(options.RedisPassword),
	); err != nil {
		return
	}
	return
}

type DB struct {
	options Options
	redis   lgredis.IRedisSys
	// mgo     mgo.IMongodb
}

func (this *DB) QueryServiceNodeInfo() (result map[string]*comm.ServiceNode, err error) {
	result = make(map[string]*comm.ServiceNode)
	err = this.redis.Get(Cache_ServiceNodeInfo, &result)
	return
}

//同步运行状态
func (this *DB) UpDateServiceNodeInfo(info *comm.ServiceNode) (original map[string]*comm.ServiceNode, err error) {
	var (
		rlock *lgredis.RedisMutex
	)

	if rlock, err = this.redis.NewRedisMutex(Cach_UpDateServiceNodeInfoLock); err != nil {
		return
	} else {
		if err = rlock.Lock(); err != nil {
			return
		}
		defer rlock.Unlock()
		original = make(map[string]*comm.ServiceNode)
		if err = this.redis.Get(Cache_ServiceNodeInfo, &original); err != nil && err != redis.Nil {
			return
		} else {
			original[info.Id] = info
			err = this.redis.Set(Cache_ServiceNodeInfo, original, -1)
		}
	}
	return
}

///删除采集任务
func (this *DB) DelRunner(rId string) (err error) {
	rediskey := fmt.Sprintf(Cache_RunnerRunInfo+"%s", rId)
	if err = this.redis.Delete(rediskey); err != nil {
		log.Errorf("DelRunner %s err:%v", rId, err)
	}
	rediskey = fmt.Sprintf(Cache_RunnerMeta, rId, "reader")
	if err = this.redis.Delete(rediskey); err != nil {
		log.Errorf("DelRunner %s err:%v", rId, err)
	}
	rediskey = fmt.Sprintf(Cache_RunnerConfig+"%s", rId)
	if err = this.redis.Delete(rediskey); err != nil {
		log.Errorf("DelRunner %s err:%v", rId, err)
	}
	return
}

///添加新的采集任务
func (this *DB) AddNewRunnerConfig(conf *comm.RunnerConfig) (err error) {
	rediskey := fmt.Sprintf(Cache_RunnerConfig+"%s", conf.Name)
	if err = this.redis.Set(rediskey, conf, -1); err != nil {
		log.Errorf("AddNewRunnerConfig %s err:%v", conf.Name, err)
	}
	return
}

///添加新的采集任务
func (this *DB) UpdateRunnerConfig_IsStop(rId string, isstopped bool) (err error) {
	result := &comm.RunnerConfig{}
	rediskey := fmt.Sprintf(Cache_RunnerConfig+"%s", rId)
	if err = this.redis.Get(rediskey, result); err != nil {
		log.Errorf("UpdateRunnerConfig_IsStop %s err:%v", rId, err)
		return
	}
	result.IsStopped = isstopped
	if err = this.redis.Set(rediskey, result, -1); err != nil {
		log.Errorf("UpdateRunnerConfig_IsStop %s err:%v", rId, err)
		return
	}
	return
}

///添加新的采集任务
func (this *DB) QueryRunnerConfig(rId string) (result *comm.RunnerConfig, err error) {
	result = &comm.RunnerConfig{}
	rediskey := fmt.Sprintf(Cache_RunnerConfig+"%s", rId)
	err = this.redis.Get(rediskey, result)
	return
}

//查询采集去运行信息数据
func (this *DB) ReadMetaData(rId, meta string, value interface{}) (err error) {
	rediskey := fmt.Sprintf(Cache_RunnerMeta, rId, meta)
	err = this.redis.Get(rediskey, value)
	return
}

//查询采集去运行信息数据
func (this *DB) WriteMetaData(rId, meta string, value interface{}) (err error) {
	rediskey := fmt.Sprintf(Cache_RunnerMeta, rId, meta)
	err = this.redis.Set(rediskey, value, 0)
	return
}

//查询采集去运行信息数据
func (this *DB) QueryRunnerRuninfo(rId string) (result *comm.RunnerRuninfo, err error) {
	result = &comm.RunnerRuninfo{}
	key := fmt.Sprintf(string(Cache_RunnerRunInfo)+"%s", rId)
	err = this.redis.Get(key, result)
	return
}

//同步运行状态
func (this *DB) WriteRunnerRuninfo(info *comm.RunnerRuninfo) (err error) {
	var (
		rlock    *lgredis.RedisMutex
		original *comm.RunnerRuninfo
		key      = fmt.Sprintf(Cache_RunnerRunInfo+"%s", info.RName)
	)
	if rlock, err = this.redis.NewRedisMutex(Cach_UpdateRunRunnerInfoLock); err != nil {
		return
	} else {
		if err = rlock.Lock(); err != nil {
			return
		}
		defer rlock.Unlock()
		original = &comm.RunnerRuninfo{
			RName:      info.RName,
			InstanceId: info.InstanceId,
			RunService: map[string]struct{}{},
			RecordTime: 0,
			ReaderCnt:  0,
			ParserCnt:  0,
			SenderCnt:  make(map[string]int64),
		}
		if err = this.redis.Get(key, &original); err != nil && err != redis.Nil {
			return
		} else {
			original.RName = info.RName
			for k, _ := range info.RunService {
				original.RunService[k] = struct{}{}
			}
			original.RecordTime = info.RecordTime
			original.ReaderCnt += info.ReaderCnt
			original.ParserCnt += info.ParserCnt
			for k, v := range info.SenderCnt {
				if _, ok := original.SenderCnt[k]; !ok {
					original.SenderCnt[k] = 0
				}
				original.SenderCnt[k] += v
			}
			err = this.redis.Set(key, original, time.Second*time.Duration(this.options.RInfoExpiration))
		}
	}
	return
}

//查询采集去运行信息数据
func (this *DB) CleanRunnerRuninfo(rId string) (err error) {
	key := fmt.Sprintf(string(Cache_RunnerRunInfo)+"%s", rId)
	err = this.redis.Delete(key)
	return
}

//获取查询启动采集任务锁
func (this *DB) GetQueryAndStartNoRunRunnerLock() (result *lgredis.RedisMutex, err error) {
	result, err = this.redis.NewRedisMutex(Cache_QueryAndStartNoRunRunnerLock, lgredis.Setdelay(time.Millisecond*100), lgredis.SetExpiry(10))
	return
}

//查询为运行服务器 使用分布式锁 确保任务执行安全
func (this *DB) QueryAndRunNoRunRunnerConfig(ip string) (result *comm.RunnerConfig, err error) {
	var (
		runkeys  []string
		runmap   map[string]struct{} = make(map[string]struct{})
		confkeys []string
	)
	if runkeys, err = this.redis.Keys(Cache_RunnerRunInfo + "*"); err != nil {
		return
	}
	for _, v := range runkeys {
		key := strings.Replace(v, string(Cache_RunnerRunInfo), "", 1)
		runmap[key] = struct{}{}
	}
	if confkeys, err = this.redis.Keys(Cache_RunnerConfig + "*"); err != nil {
		return
	}
	isContains := func(rip []string, ip string) bool {
		for _, v := range rip {
			if ip == v {
				return true
			}
		}
		return false
	}
	result = &comm.RunnerConfig{}
	for _, k := range confkeys {
		key := strings.Replace(k, string(Cache_RunnerConfig), "", 1)
		if _, ok := runmap[key]; !ok {
			if err = this.redis.Get(k, result); err == nil && !result.IsStopped && isContains(result.RunIp, ip) {
				log.Debugf("扫描到尾开启 任务:%s", k)
				return
			}
		}
	}
	err = redis.Nil
	return
}

//实时输入输入采集
func (this *DB) RealTimeStatisticsIn(rid string, in int64) (err error) {
	var result int64
	inkey := fmt.Sprintf(string(Cache_RunnerStatistics_Reader), rid, time.Now().Format("2006010215"))
	if result, err = this.redis.INCRBY(inkey, in); err == nil && result == in { //第一次写入
		this.redis.ExpireKey(inkey, this.options.RStatisticsExpiration)
	}
	return
}

//实时输入输出采集
func (this *DB) RealTimeStatisticsOut(rid, stype string, out int64) (err error) {
	var result int64
	outkey := fmt.Sprintf(string(Cache_RunnerStatistics_Sender), rid, stype, time.Now().Format("2006010215"))
	if result, err = this.redis.INCRBY(outkey, out); err == nil && result == out { //第一次写入
		this.redis.ExpireKey(outkey, this.options.RStatisticsExpiration)
	}
	return
}
