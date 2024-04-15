package redis

import (
	"context"
	"fmt"
	"lego_datacollector/comm"
	"lego_datacollector/modules/datacollector/core"
	msql "lego_datacollector/modules/datacollector/metaer/sql"
	"lego_datacollector/modules/datacollector/reader"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/liwei1dao/lego"
	lgcron "github.com/liwei1dao/lego/sys/cron"
	"github.com/liwei1dao/lego/sys/event"
	"github.com/liwei1dao/lego/sys/redis"
	"github.com/robfig/cron/v3"
)

type Reader struct {
	reader.Reader
	options      IOptions            //以接口对象传递参数 方便后期继承扩展
	meta         msql.ITableMetaData //愿数据
	addr         string
	redis        redis.IRedisSys //redis 驱动库
	rungoroutine int32           //运行采集携程数
	isend        int32
	runId        cron.EntryID
	wg           sync.WaitGroup       //用于等待采集器完全关闭
	colltask     chan *msql.TableMeta //文件采集任务
}

func (this *Reader) Init(runner core.IRunner, reader core.IReader, meta core.IMetaerData, options core.IReaderOptions) (err error) {
	if err = this.Reader.Init(runner, reader, meta, options); err != nil {
		return
	}
	this.options = options.(IOptions)
	this.meta = meta.(msql.ITableMetaData)
	this.addr = this.options.GetRedis_url()[0][0:strings.LastIndex(this.options.GetRedis_url()[0], ":")]
	this.redis, err = redis.NewSys(
		redis.SetRedisType(this.options.GetRedis_type()),
		redis.SetRedis_Single_Addr(this.options.GetRedis_url()[0]),
		redis.SetRedis_Single_DB(this.options.GetRedis_db()),
		redis.SetRedis_Single_Password(this.options.GetRedis_password()),
		redis.Redis_Cluster_Addr(this.options.GetRedis_url()),
		redis.SetRedis_Cluster_Password(this.options.GetRedis_password()),
	)
	return
}

func (this *Reader) Start() (err error) {
	err = this.Reader.Start()
	atomic.StoreInt32(&this.rungoroutine, 0)
	atomic.StoreInt32(&this.isend, 0)
	if !this.options.GetAutoClose() {
		if this.runId, err = lgcron.AddFunc(this.options.GetRedis_intervel(), this.scanSql); err != nil {
			err = fmt.Errorf("lgcron.AddFunc corn:%s err:%v", this.options.GetRedis_intervel(), err)
			return
		}
		if this.options.GetRedis_exec_onstart() {
			go this.scanSql()
		}
	} else {
		go this.scanSql()
	}
	return
}

///外部调度器 驱动执行  此接口 不可阻塞
func (this *Reader) Drive() (err error) {
	err = this.Reader.Drive()
	go this.scanSql()
	return
}

func (this *Reader) Close() (err error) {
	if !this.options.GetAutoClose() {
		lgcron.Remove(this.runId)
	}
	this.wg.Wait()
	this.redis.Close()
	err = this.Reader.Close()
	return
}

//定时扫描sql
func (this *Reader) scanSql() {
	var (
		tables           map[string]uint64
		collectiontables []*msql.TableMeta
		table            *msql.TableMeta
		ok               bool
		err              error
	)
	if !atomic.CompareAndSwapInt32(&this.isend, 0, 1) {
		this.Runner.Debugf("Reader is collectioning rungoroutine:%d", atomic.LoadInt32(&this.rungoroutine))
		this.SyncMeta()
		return
	}
	this.Runner.Debugf("Reader scan start!")
	if tables, err = this.scanddatabase(); err == nil && len(tables) > 0 {
		collectiontables = make([]*msql.TableMeta, 0)
		for k, _ := range tables {
			if table, ok = this.meta.GetTableMeta(k); !ok {
				table = &msql.TableMeta{
					TableName:              k,
					TableDataCount:         0,
					TableAlreadyReadOffset: 0,
				}
				this.meta.SetTableMeta(k, table)
			}
			if table.TableAlreadyReadOffset == 0 {
				collectiontables = append(collectiontables, table)
			}
		}
		if len(collectiontables) > 0 {
			procs := this.Runner.MaxProcs()
			if len(collectiontables) < this.Runner.MaxProcs() {
				procs = len(collectiontables)
			}
			this.colltask = make(chan *msql.TableMeta, len(collectiontables))
			for _, v := range collectiontables {
				this.colltask <- v
			}
			this.Runner.Debugf("Reader Rides start new collection:%d", atomic.AddInt32(&this.rungoroutine, int32(procs)))
			this.wg.Add(procs)
			for i := 0; i < procs; i++ {
				go this.asyncollection(this.redis, this.colltask)
			}
		} else {
			this.Runner.Debugf("Reader scan no found new data!")
			if this.options.GetAutoClose() { //自动关闭
				event.TriggerEvent(comm.Event_WriteLog, this.Runner.Name(), this.Runner.InstanceId(), "未扫描到新的可采集数据!", 1, time.Now().Unix())
				go this.AutoClose(core.RunnerFailAutoClose)
			} else {
				atomic.StoreInt32(&this.isend, 0)
			}
		}
	} else if err != nil { //链接对象异常断开 需要重联
		this.Runner.Errorf("Reader scanddatabase err:%v", err)
		if this.options.GetAutoClose() { //自动关闭
			event.TriggerEvent(comm.Event_WriteLog, this.Runner.Name(), this.Runner.InstanceId(), fmt.Sprintf("扫描目标数据源失败:%v", err), 1, time.Now().Unix())
			go this.AutoClose(core.RunnerFailAutoClose)
		} else {
			atomic.StoreInt32(&this.isend, 0)
			if strings.Contains(err.Error(), "client is closed") && this.Runner.GetRunnerState() == core.Runner_Runing {
				this.redis.Close()
				if this.redis, err = redis.NewSys(
					redis.SetRedisType(this.options.GetRedis_type()),
					redis.SetRedis_Single_Addr(this.options.GetRedis_url()[0]),
					redis.SetRedis_Single_DB(this.options.GetRedis_db()),
					redis.SetRedis_Single_Password(this.options.GetRedis_password()),
					redis.Redis_Cluster_Addr(this.options.GetRedis_url()),
					redis.SetRedis_Cluster_Password(this.options.GetRedis_password()),
				); err != nil {
					this.Runner.Errorf("Reader recomment sql err:%v", err)
				} else { //重启成功 重写扫描
					go this.scanSql()
				}
			}
		}
	} else {
		this.Runner.Debugf("Reader scan no found table!")
		if this.options.GetAutoClose() { //自动关闭
			event.TriggerEvent(comm.Event_WriteLog, this.Runner.Name(), this.Runner.InstanceId(), "未找到目标数据Key!", 1, time.Now().Unix())
			go this.AutoClose(core.RunnerFailAutoClose)
		} else {
			atomic.StoreInt32(&this.isend, 0)
		}
	}
}

func (this *Reader) asyncollection(db redis.IRedisSys, fmeta <-chan *msql.TableMeta) {
	defer lego.Recover(fmt.Sprintf("%s Reader", this.Runner.Name()))
	defer atomic.AddInt32(&this.rungoroutine, -1)
locp:
	for v := range fmeta {
		if err := this.collection_table(db, v); err != nil {
			this.Runner.Errorf("Reader collection_table err:%v", err)
		}
		if this.Runner.GetRunnerState() == core.Runner_Stoping || len(fmeta) == 0 {
			break locp
		} else {
			this.SyncMeta()
		}
	}
	this.SyncMeta()
	this.wg.Done()
	if atomic.CompareAndSwapInt32(&this.isend, 1, 2) { //最后一个任务已经完成
		this.collectionend()
	}
	this.Runner.Debugf("Reader asyncollection exit succ!")
}

///采集结束
func (this *Reader) collectionend() {
	close(this.colltask)
	this.wg.Wait()
	atomic.StoreInt32(&this.isend, 0)
	this.Runner.Debugf("Reader collection end succ!")
	if this.options.GetAutoClose() { //自动关闭
		go this.AutoClose(core.RunnerSuccAutoClose)
	}
}

func (this *Reader) AutoClose(msg string) {
	time.Sleep(time.Second * 10)
	this.Runner.Close(core.Runner_Runing, msg) //结束任务
}

//-------------------------------------------------------------------------------------------------------------------------------
//扫描数据库 扫描数据库下所有的表 以及表中的数据条数
func (this *Reader) scanddatabase() (tables map[string]uint64, err error) {
	var (
		data []string
	)
	tables = make(map[string]uint64)

	if data, err = this.redis.Keys(this.options.GetRedis_pattern()); err == nil {
		for _, v := range data {
			tables[v] = 0
		}
	}
	return
}

//采集数据表
func (this *Reader) collection_table(db redis.IRedisSys, table *msql.TableMeta) (err error) {
	var (
		ty     string
		result interface{}
		data   map[string]interface{}
	)
	if ty, err = db.Type(table.TableName); err == nil {
		switch ty {
		case "string":
			if result, err = db.Do(context.TODO(), "get", table.TableName).Result(); err == nil {
				data = map[string]interface{}{table.TableName: result}
				// this.Runner.Debugf("collection_table:%v", data)
				this.writeDataChan(table, data)
			}
			break
		case "list":
			if result, err = db.Do(context.TODO(), "LRANGE", table.TableName, 0, -1).Result(); err == nil {
				data = map[string]interface{}{table.TableName: result}
				// this.Runner.Debugf("collection_table:%v", data)
				this.writeDataChan(table, data)
			}
			break
		case "hash":
			if result, err = db.Do(context.TODO(), "HGETALL", table.TableName).Result(); err == nil {
				data = map[string]interface{}{table.TableName: result}
				// this.Runner.Debugf("collection_table:%v", data)
				this.writeDataChan(table, data)
			}
			break
		case "set":
			if result, err = db.Do(context.TODO(), "SMEMBERS", table.TableName).Result(); err == nil {
				data = map[string]interface{}{table.TableName: result}
				// this.Runner.Debugf("collection_table:%v", data)
				this.writeDataChan(table, data)
			}
			break
		case "zset":
			if result, err = db.Do(context.TODO(), "ZRANGE", table.TableName, 0, -1, "WITHSCORES").Result(); err == nil {
				data = map[string]interface{}{table.TableName: result}
				// this.Runner.Debugf("collection_table:%v", data)
				this.writeDataChan(table, data)
			}
			break
		default:
			this.Runner.Errorf("Reader redis unknown key:%s type:%s", table.TableName, ty)
			break
		}
	}

	return
}

//发送数据 数据量打的时候会有堵塞
func (this *Reader) writeDataChan(table *msql.TableMeta, data map[string]interface{}) {
	this.Input() <- core.NewCollData(this.addr, data)
	table.TableAlreadyReadOffset = 1
	return
}
