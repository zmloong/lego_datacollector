package elastic

import (
	"context"
	"fmt"
	"io"
	"lego_datacollector/comm"
	"lego_datacollector/modules/datacollector/core"
	msql "lego_datacollector/modules/datacollector/metaer/elastic"
	"lego_datacollector/modules/datacollector/reader"
	. "lego_datacollector/utils/sql"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/liwei1dao/lego"
	lgcron "github.com/liwei1dao/lego/sys/cron"
	"github.com/liwei1dao/lego/sys/event"
	"github.com/olivere/elastic/v7"
	"github.com/robfig/cron/v3"
)

type Reader struct {
	reader.Reader
	options      IOptions            //以接口对象传递参数 方便后期继承扩展
	meta         msql.ITableMetaData //愿数据
	client       *elastic.Client
	runId        cron.EntryID
	rungoroutine int32                //运行采集携程数
	isend        int32                //采集任务结束
	lock         sync.Mutex           //执行锁
	wg           sync.WaitGroup       //用于等待采集器完全关闭
	colltask     chan *msql.TableMeta //文件采集任务
	schemas      map[string]string    //sql 类型处理
}

func (this *Reader) Type() string {
	return ReaderType
}

func (this *Reader) Init(runner core.IRunner, reader core.IReader, meta core.IMetaerData, options core.IReaderOptions) (err error) {
	if err = this.Reader.Init(runner, reader, meta, options); err != nil {
		return
	}
	this.options = options.(IOptions)
	this.meta = meta.(msql.ITableMetaData)
	this.colltask = make(chan *msql.TableMeta)
	if this.schemas, err = SchemaCheck(this.options.GetElastic_schema()); err != nil {
		return
	}
	this.client, err = this.getClient(this.options.GetElastic_addr())
	return
}
func (this *Reader) getClient(addr string) (client *elastic.Client, err error) {
	//这个地方有个小坑 不加上elastic.SetSniff(false) 会连接不上
	client, err = elastic.NewClient(
		elastic.SetSniff(false),
		elastic.SetHealthcheck(false),
		elastic.SetURL(addr))
	if err != nil {
		err = fmt.Errorf("elastic getClient err:%v", err)
		return
	}
	return
}
func (this *Reader) Start() (err error) {
	err = this.Reader.Start()
	atomic.StoreInt32(&this.rungoroutine, 0)
	atomic.StoreInt32(&this.isend, 0)
	if !this.options.GetAutoClose() {
		if this.runId, err = lgcron.AddFunc(this.options.GetElastic_batch_intervel(), this.scanSql); err != nil {
			err = fmt.Errorf("lgcron.AddFunc corn:%s err:%v", this.options.GetElastic_batch_intervel(), err)
			return
		}
		if this.options.GetElastic_exec_onstart() {
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
	this.client.Stop()
	err = this.Reader.Close()
	return
}

//定时扫描sql
func (this *Reader) scanSql() {
	// this.lock.Lock()
	// defer this.lock.Unlock()
	var (
		tables           map[string]uint64
		collectiontables []*msql.TableMeta
		table            *msql.TableMeta
		tablecount       uint64
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
		if this.options.GetElastic_collection_type() == StaticTableCollection { //静态表采集
			for _, k := range this.options.GetElastic_tables() {
				if v, ok := tables[k]; ok {
					if table, ok = this.meta.GetTableMeta(k); !ok {
						table = &msql.TableMeta{
							TableName:              k,
							TableDataCount:         v,
							TableAlreadyReadOffset: 0,
						}
						this.meta.SetTableMeta(k, table)
					}
					tablecount, _ = this.gettablecount(k)
					if table.TableAlreadyReadOffset < tablecount { //有新的数据
						table.TableDataCount = tablecount
						collectiontables = append(collectiontables, table)
					}
				} else {
					this.Runner.Errorf("Reader Elastic not found table：%s", k)
				}
			}
		} else if this.options.GetElastic_collection_type() == DynamicTableCollection { //动态表
			var (
				nextDate    string
				canIncrease bool
				dynamictame string
			)
			for _, k := range this.options.GetElastic_tables() {
				var (
					dynamictable *msql.TableMeta
				)
				if dynamictable, ok = this.meta.GetTableMeta(k); !ok {
					dynamictable = &msql.TableMeta{
						TableName:              k,
						TableDataCount:         0,
						TableAlreadyReadOffset: 0,
					}
					this.meta.SetTableMeta(k, dynamictable)
				}

			locp:
				for {
					nextDate, canIncrease = GetNextDate(this.options.GetElastic_datetimeFormat(), this.options.GetElastic_startDate(), int(dynamictable.TableAlreadyReadOffset), this.options.GetElastic_dateUnit())
					if canIncrease {
						dynamictable.TableAlreadyReadOffset += uint64(this.options.GetElastic_dateGap())
						dynamictame = k + nextDate
						if v, ok := tables[dynamictame]; ok {
							if table, ok = this.meta.GetTableMeta(dynamictame); !ok {
								table = &msql.TableMeta{
									TableName:              dynamictame,
									TableDataCount:         v,
									TableAlreadyReadOffset: 0,
								}
								this.meta.SetTableMeta(dynamictame, table)
							}
							tablecount, _ = this.gettablecount(dynamictame)
							if table.TableAlreadyReadOffset < tablecount { //有新的数据
								table.TableDataCount = tablecount
								collectiontables = append(collectiontables, table)
							}
						} else {
							this.Runner.Errorf("Reader Elastic not found table：%s", dynamictame)
						}
					} else {
						this.SyncMeta()
						break locp
					}
				}

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
			this.Runner.Debugf("Reader Elastic start new collection:%d", atomic.AddInt32(&this.rungoroutine, int32(procs)))
			this.wg.Add(procs)
			for i := 0; i < procs; i++ {
				go this.asyncollection(this.colltask)
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
			event.TriggerEvent(comm.Event_WriteLog, this.Runner.Name(), this.Runner.InstanceId(), fmt.Sprintf("采集数据失败 错误:%v", err), 1, time.Now().Unix())
			go this.AutoClose(core.RunnerFailAutoClose)
		} else {
			atomic.StoreInt32(&this.isend, 0)
			if strings.Contains(err.Error(), "database is closed") && this.Runner.GetRunnerState() == core.Runner_Runing {
				this.client.Stop()
				if this.client, err = this.getClient(this.options.GetElastic_addr()); err != nil {
					this.Runner.Errorf("Reader recomment sql err:%v", err)
				} else { //重启成功 重写扫描
					go this.scanSql()
				}
			}
		}
	} else {
		this.Runner.Debugf("Reader scan no found table!")
		if this.options.GetAutoClose() { //自动关闭
			event.TriggerEvent(comm.Event_WriteLog, this.Runner.Name(), this.Runner.InstanceId(), "未找到目标数据表!", 1, time.Now().Unix())
			go this.AutoClose(core.RunnerFailAutoClose)
		} else {
			atomic.StoreInt32(&this.isend, 0)
		}
	}
}

func (this *Reader) asyncollection(fmeta <-chan *msql.TableMeta) {
	defer lego.Recover(fmt.Sprintf("%s Reader", this.Runner.Name()))
	defer atomic.AddInt32(&this.rungoroutine, -1)
locp:
	for v := range fmeta {
	clocp:
		for {
			if ok, err := this.collection_table(v); ok || err != nil {
				this.Runner.Debugf("Reader collection_table ok:%v,err:%v", ok, err)
				break clocp
			}

			if this.Runner.GetRunnerState() == core.Runner_Stoping { //采集器进入停止过程中
				this.Runner.Debugf("Reader asyncollection exit")
				break locp
			}

		}
		this.Runner.Debugf("Reader SyncMeta:%s count:%d offset:%d ", v.TableName, v.TableDataCount, v.TableAlreadyReadOffset)
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

//静态表模式 全量采集结束后 延时关闭
func (this *Reader) sleepClose() {
	time.Sleep(10 * time.Second)
	this.Runner.Close(core.Runner_Runing, "静态表模式 全量采集结束，任务关闭！")
}

//-------------------------------------------------------------------------------------------------------------------------------
//扫描数据库 扫描数据库下所有的表 以及表中的数据条数
func (this *Reader) scanddatabase() (tables map[string]uint64, err error) {

	tables = make(map[string]uint64)
	if IndexNames, err := this.client.IndexNames(); err == nil {
		for _, v := range IndexNames {
			total, err := this.gettablecount(v)
			if err != nil {
				this.Runner.Errorf("Reader scanddatabase table【%s】 err:%v", v, err)
				continue
			}
			tables[v] = uint64(total)
		}
	}
	return
}

//采集数据表
func (this *Reader) collection_table(table *msql.TableMeta) (isend bool, err error) {
	//时间范围过滤
	if this.options.GetElastic_query_starttime() != "" &&
		this.options.GetElastic_query_endtime() != "" {
		isend, err = this.getAllDatasBytimeRange(table)
		return
	}
	//静态表id增量
	if this.options.GetElastic_collection_type() == StaticTableCollection &&
		this.options.GetElastic_offset_key() != "" {
		isend, err = this.getAllDatasByoffsetKey(table)
		return
	}
	isend, err = this.getAllDatas(table)
	return
}

//采集数据表
/*
func (this *Reader) collection_table(table *msql.TableMeta) (isend bool, err error) {
	//过滤条件
	//1.静态表+时间范围
	StaticTimeRange := this.options.GetElastic_collection_type() == StaticTableCollection &&
		this.options.GetAutoClose() &&
		this.options.GetElastic_query_starttime() != "" &&
		this.options.GetElastic_query_endtime() != ""
	//2.动态表+时间范围
	DynamicTimeRange := this.options.GetElastic_collection_type() == DynamicTableCollection &&
		!this.options.GetAutoClose() &&
		this.options.GetElastic_query_starttime() != "" &&
		this.options.GetElastic_query_endtime() != ""
	//讨论
	if StaticTimeRange || DynamicTimeRange {
		isend, err = this.getAllDatasBytimeRange(table)
		return
	}
	//静态表id增量
	if this.options.GetElastic_collection_type() == StaticTableCollection &&
		!this.options.GetAutoClose() &&
		this.options.GetElastic_offset_key() != "" {
		isend, err = this.getAllDatasByoffsetKey(table)
		return
	}
	isend, err = this.getAllDatas(table)
	return
}
*/
//*获取表长度
func (this *Reader) gettablecount(tablename string) (count uint64, err error) {

	if this.options.GetElastic_timestamp_key() != "" {
		var results *elastic.SearchResult
		results, err = this.querydataCountBytime(tablename)
		if err != nil {
			this.Runner.Errorf("Reader getdataBytime 时间范围查询 table【%s】 err:%v", tablename, err)
			return
		}
		count = uint64(results.Hits.TotalHits.Value)
		return
	}
	//无过滤条件
	total, err := this.client.Count(tablename).Do(context.Background())
	if err != nil {
		this.Runner.Errorf("Reader gettablecount table【%s】 err:%v", tablename, err)
		return
	}
	count = uint64(total)
	return
}

//*获取表长度
/*
func (this *Reader) gettablecount(tablename string) (count uint64, err error) {
	//有过滤条件
	//1.静态表+时间范围
	StaticTimeRange := this.options.GetElastic_collection_type() == StaticTableCollection &&
		this.options.GetAutoClose() &&
		this.options.GetElastic_query_starttime() != "" &&
		this.options.GetElastic_query_endtime() != ""
	//2.动态表+时间范围
	DynamicTimeRange := this.options.GetElastic_collection_type() == DynamicTableCollection &&
		!this.options.GetAutoClose() &&
		this.options.GetElastic_query_starttime() != "" &&
		this.options.GetElastic_query_endtime() != ""
	if StaticTimeRange || DynamicTimeRange {
		var results *elastic.SearchResult
		results, err = this.querydataCountBytime(tablename)
		if err != nil {
			this.Runner.Errorf("Reader getdataBytime 时间范围查询 table【%s】 err:%v", tablename, err)
			return
		}
		count = uint64(results.Hits.TotalHits.Value)
	}
	//无过滤条件
	total, err := this.client.Count(tablename).Do(context.Background())
	if err != nil {
		this.Runner.Errorf("Reader gettablecount table【%s】 err:%v", tablename, err)
		return
	}
	count = uint64(total)
	return
}
*/
//*获取表长度+加时间范围
func (this *Reader) querydataCountBytime(tablename string) (results *elastic.SearchResult, err error) {
	boolQ := elastic.NewBoolQuery()
	if this.options.GetElastic_timestamp_type() == TimestampTypeEsDate {
		boolQ.Filter(elastic.NewRangeQuery(this.options.GetElastic_timestamp_key()).Gte(this.options.GetElastic_query_starttime()).Lte(this.options.GetElastic_query_endtime()).Format("yyyy-MM-dd HH:mm:ss"))
	} else if this.options.GetElastic_timestamp_type() == TimestampTypeUnix {
		boolQ.Filter(elastic.NewRangeQuery(this.options.GetElastic_timestamp_key()).Gte(this.options.GetQuery_starttime()).Lte(this.options.GetQuery_endtime()))
	} else {
		err = fmt.Errorf(`querydataCountBytime 指定日期搜索类型 "timestamp_type" 不存在 `)
		return
	}
	results, err = this.client.Search(tablename).Query(boolQ).Do(context.Background())

	return
}

//*指定时间字段 按时间范围读取数据
func (this *Reader) getAllDatasBytimeRange(table *msql.TableMeta) (isend bool, err error) {
	var num int
	for !isend {
		boolQ := elastic.NewBoolQuery()
		if this.options.GetElastic_timestamp_type() == 1 {
			boolQ.Filter(elastic.NewRangeQuery(this.options.GetElastic_timestamp_key()).Gte(this.options.GetElastic_query_starttime()).Lte(this.options.GetElastic_query_endtime()))
		} else if this.options.GetElastic_timestamp_type() == 2 {
			boolQ.Filter(elastic.NewRangeQuery(this.options.GetElastic_timestamp_key()).Gte(this.options.GetQuery_starttime()).Lte(this.options.GetQuery_endtime()))
		} else {
			err = fmt.Errorf(`getAllDatasBytimeRange 指定日期搜索类型 "timestamp_type" 不存在 `)
			return
		}
		var offset string
		scroll := this.client.Scroll(table.TableName).Query(boolQ).Sort(this.options.GetElastic_timestamp_key(), true).Size(this.options.GetElastic_limit_batch()).KeepAlive("10m")
		for {
			var results *elastic.SearchResult
			if offset != "" {
				scroll = scroll.ScrollId(offset)
			}
			results, err = scroll.Do(context.Background())
			if err == io.EOF {
				this.Runner.Debugf("Reader getAllDatasBytimeRange table【%s】 all results retrieved:%v", table.TableName, err)
				return // all results retrieved
			}
			if err != nil {
				// something went wrong
				this.Runner.Errorf("Reader getAllDatasBytimeRange table【%s】 err:%v", table.TableName, err)
				return
			}
			for _, hit := range results.Hits.Hits {
				num++
				if num > int(table.TableAlreadyReadOffset) {
					data, _ := hit.Source.MarshalJSON()
					isend = this.writeDataChan(table, string(data))
				}

			}
			offset = results.ScrollId
		}
	}
	return
}

//*指定  offset  增量读取数据
func (this *Reader) getAllDatasByoffsetKey(table *msql.TableMeta) (isend bool, err error) {

	for !isend {
		fmt.Printf("getAllDatasByoffsetKey：%v,%v,%v",
			this.options.GetElastic_offset_key(),
			table.TableAlreadyReadOffset,
			table.TableAlreadyReadOffset+uint64(this.options.GetElastic_limit_batch()))
		boolQ := elastic.NewBoolQuery()
		boolQ.Filter(elastic.NewRangeQuery(this.options.GetElastic_offset_key()).
			Gte(table.TableAlreadyReadOffset).
			Lt(table.TableAlreadyReadOffset + uint64(this.options.GetElastic_limit_batch())))

		results, err := this.client.Search(table.TableName).Query(boolQ).Size(this.options.GetElastic_limit_batch()).Do(context.Background())
		if err != nil {
			println(err.Error())
		}
		for _, hit := range results.Hits.Hits {
			data, _ := hit.Source.MarshalJSON()
			isend = this.writeDataChan(table, string(data))
		}
	}
	return
}

//*无指定 offset 全表读取数据
func (this *Reader) getAllDatas(table *msql.TableMeta) (isend bool, err error) {
	for !isend {
		var offset string
		scroll := this.client.Scroll(table.TableName).Size(this.options.GetElastic_limit_batch()).KeepAlive("1h")
		for {
			var results *elastic.SearchResult
			if offset != "" {
				scroll = scroll.ScrollId(offset)
			}
			results, err = scroll.Do(context.Background())
			if err == io.EOF {
				this.Runner.Debugf("Reader getAllDatas table【%s】 all results retrieved:%v", table.TableName, err)
				return // all results retrieved
			}
			if err != nil {
				// something went wrong
				this.Runner.Errorf("Reader getAllDatas table【%s】 err:%v", table.TableName, err)
				return
			}
			for _, hit := range results.Hits.Hits {
				data, _ := hit.Source.MarshalJSON()
				isend = this.writeDataChan(table, string(data))
			}
			offset = results.ScrollId
		}
	}
	return
}

//发送数据 数据量打的时候会有堵塞
func (this *Reader) writeDataChan(table *msql.TableMeta, data string) (isend bool) {
	// this.Runner.Debugf("reder mysql table:%s writeDataChan start", table.TableName)
	// defer this.Runner.Debugf("reder mysql table:%s writeDataChan end", table.TableName)
	this.Input() <- core.NewCollData(this.options.GetElastic_ip(), data)
	table.TableAlreadyReadOffset += 1
	if table.TableAlreadyReadOffset >= table.TableDataCount {
		isend = true
	}
	return
}
