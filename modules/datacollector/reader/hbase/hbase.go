package hbase

import (
	"context"
	"fmt"
	"io"
	"lego_datacollector/comm"
	"lego_datacollector/modules/datacollector/core"
	hbase "lego_datacollector/modules/datacollector/metaer/hbase"
	"lego_datacollector/modules/datacollector/reader"
	. "lego_datacollector/utils/sql"
	"log"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	_ "github.com/lib/pq"
	"github.com/liwei1dao/lego"
	lgcron "github.com/liwei1dao/lego/sys/cron"
	"github.com/liwei1dao/lego/sys/event"
	"github.com/robfig/cron/v3"
	"github.com/tsuna/gohbase"
	"github.com/tsuna/gohbase/filter"
	"github.com/tsuna/gohbase/hrpc"
)

type Reader struct {
	reader.Reader
	options      IOptions             //以接口对象传递参数 方便后期继承扩展
	meta         hbase.ITableMetaData //愿数据
	client       gohbase.Client
	adminClient  gohbase.AdminClient
	runId        cron.EntryID
	rungoroutine int32                 //运行采集携程数
	isend        int32                 //采集任务结束
	lock         sync.Mutex            //执行锁
	wg           sync.WaitGroup        //用于等待采集器完全关闭
	colltask     chan *hbase.TableMeta //文件采集任务
	schemas      map[string]string     //sql 类型处理
}

func (this *Reader) Type() string {
	return ReaderType
}
func (this *Reader) Init(runner core.IRunner, reader core.IReader, meta core.IMetaerData, options core.IReaderOptions) (err error) {
	if err = this.Reader.Init(runner, reader, meta, options); err != nil {
		return
	}
	this.options = options.(IOptions)
	this.meta = meta.(hbase.ITableMetaData)
	this.colltask = make(chan *hbase.TableMeta)
	if this.schemas, err = SchemaCheck(this.options.GetHbase_schema()); err != nil {
		return
	}
	if this.client, err = this.connectionClient(); err != nil {
		return
	}
	if this.adminClient, err = this.connectionAdminClient(); err != nil {
		return
	}
	return
}
func (this *Reader) connectionClient() (client gohbase.Client, err error) {
	client = gohbase.NewClient(this.options.GetHbase_host())
	return
}
func (this *Reader) connectionAdminClient() (adminclient gohbase.AdminClient, err error) {
	adminclient = gohbase.NewAdminClient(this.options.GetHbase_host())
	return
}
func (this *Reader) Start() (err error) {
	err = this.Reader.Start()
	atomic.StoreInt32(&this.rungoroutine, 0)
	atomic.StoreInt32(&this.isend, 0)
	if !this.options.GetAutoClose() {
		if this.runId, err = lgcron.AddFunc(this.options.GetHbase_batch_intervel(), this.scanSql); err != nil {
			err = fmt.Errorf("lgcron.AddFunc corn:%s err:%v", this.options.GetHbase_batch_intervel(), err)
			return
		}
		if this.options.GetHbase_exec_onstart() {
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
	this.client.Close()
	err = this.Reader.Close()
	return
}

//定时扫描sql
func (this *Reader) scanSql() {
	// this.lock.Lock()
	// defer this.lock.Unlock()
	var (
		tables           map[string]uint64
		collectiontables []*hbase.TableMeta
		table            *hbase.TableMeta
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
		collectiontables = make([]*hbase.TableMeta, 0)
		if this.options.GetHbase_collection_type() == StaticTableCollection { //静态表采集
			for _, k := range this.options.GetHbase_tables() {
				if v, ok := tables[k]; ok {
					if table, ok = this.meta.GetTableMeta(k); !ok {
						table = &hbase.TableMeta{
							TableName:              k,
							TableDataCount:         v,
							TableAlreadyReadOffset: 0,
							TableAlreadyKey:        make(map[string]string),
						}
						this.meta.SetTableMeta(k, table)
					}
					tablecount = this.gettablecount(k)
					if table.TableAlreadyReadOffset < tablecount { //有新的数据
						table.TableDataCount = tablecount
						collectiontables = append(collectiontables, table)
					}
				} else {
					this.Runner.Errorf("Reader Hbase not found table：%s", k)
				}
			}
		} else if this.options.GetHbase_collection_type() == DynamicTableCollection { //动态表
			var (
				nextDate    string
				canIncrease bool
				dynamictame string
			)
			for _, k := range this.options.GetHbase_tables() {
				var (
					dynamictable *hbase.TableMeta
				)
				if dynamictable, ok = this.meta.GetTableMeta(k); !ok {
					dynamictable = &hbase.TableMeta{
						TableName:              k,
						TableDataCount:         0,
						TableAlreadyReadOffset: 0,
						TableAlreadyKey:        make(map[string]string),
					}
					this.meta.SetTableMeta(k, dynamictable)
				}

			locp:
				for {
					nextDate, canIncrease = GetNextDate(this.options.GetHbase_datetimeFormat(), this.options.GetHbase_startDate(), int(dynamictable.TableAlreadyReadOffset), this.options.GetHbase_dateUnit())
					if canIncrease {
						dynamictable.TableAlreadyReadOffset += uint64(this.options.GetHbase_dateGap())
						dynamictame = k + nextDate
						if v, ok := tables[dynamictame]; ok {
							if table, ok = this.meta.GetTableMeta(dynamictame); !ok {
								table = &hbase.TableMeta{
									TableName:              dynamictame,
									TableDataCount:         v,
									TableAlreadyReadOffset: 0,
									TableAlreadyKey:        make(map[string]string),
								}
								this.meta.SetTableMeta(dynamictame, table)
							}
							tablecount = this.gettablecount(dynamictame)
							if table.TableAlreadyReadOffset < tablecount { //有新的数据
								table.TableDataCount = tablecount
								collectiontables = append(collectiontables, table)
							}
						} else {
							this.Runner.Errorf("Reader Hbase not found table：%s", dynamictame)
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
			this.colltask = make(chan *hbase.TableMeta, len(collectiontables))
			for _, v := range collectiontables {
				this.colltask <- v
			}
			this.Runner.Debugf("Reader Hbase start new collection:%d", atomic.AddInt32(&this.rungoroutine, int32(procs)))
			this.wg.Add(procs)
			for i := 0; i < procs; i++ {
				go this.asyncollection(this.client, this.colltask)
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
				this.client.Close()
				if this.client, err = this.connectionClient(); err != nil {
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

func (this *Reader) asyncollection(db gohbase.Client, fmeta <-chan *hbase.TableMeta) {
	defer lego.Recover(fmt.Sprintf("%s Reader", this.Runner.Name()))
	defer atomic.AddInt32(&this.rungoroutine, -1)
locp:
	for v := range fmeta {
	clocp:
		for {
			if ok, err := this.collection_table(db, v); ok || err != nil {
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

//-------------------------------------------------------------------------------------------------------------------------------
//扫描数据库 扫描数据库下所有的表 以及表中的数据条数
func (this *Reader) scanddatabase() (tables map[string]uint64, err error) {
	tables = make(map[string]uint64)
	res, err := hrpc.NewListTableNames(context.Background())
	if err != nil {
		fmt.Println(err)
	}
	if ResNames, err := this.adminClient.ListTableNames(res); err == nil {
		for _, v := range ResNames {
			count := uint64(0)
			tables[string(v.Qualifier)] = count
		}
	} else {
		this.Runner.Errorf("ListTableNames:%v", err)
	}
	return
}

//采集数据表
func (this *Reader) collection_table(db gohbase.Client, table *hbase.TableMeta) (isend bool, err error) {
	isend, err = this.getAllDatas(table)
	return
}

//获取表长度
func (this *Reader) gettablecount(tablename string) (count uint64) {
	pFilter := filter.NewPrefixFilter([]byte(""))
	scanRequest, err := hrpc.NewScanStr(context.Background(), tablename, hrpc.Filters(pFilter))
	if err != nil {
		fmt.Println(err)
	}
	scanRsp := this.client.Scan(scanRequest)
	for {
		getRsp, err := scanRsp.Next()
		if err == io.EOF || getRsp == nil {
			break
		}
		if err != nil {
			log.Print(err, "scan.Next")
		}
		count++
	}
	return
}

//读取数据
func (this *Reader) getAllDatas(table *hbase.TableMeta) (isend bool, err error) {
	pFilter := filter.NewPrefixFilter([]byte(""))
	scanRequest, err := hrpc.NewScanStr(context.Background(), table.TableName, hrpc.Filters(pFilter))
	if err != nil {
		fmt.Println(err)
	}
	scanRsp := this.client.Scan(scanRequest)
	for {
		getRsp, err := scanRsp.Next()
		if err == io.EOF || getRsp == nil {
			break
		}
		if err != nil {
			log.Print(err, "scan.Next")
		}
		if len(getRsp.Cells) < 0 {
			continue
		}
		rowkey := fmt.Sprintf("%s", getRsp.Cells[0].Row)
		if _, ok := table.TableAlreadyKey[rowkey]; ok {
			continue
		}
		var res []string
		for _, v := range getRsp.Cells {
			res = append(res, v.String())
		}
		data := make(map[string]interface{})
		data[rowkey] = res
		isend = this.writeDataChan(table, rowkey, data)

	}
	return
}

//发送数据 数据量打的时候会有堵塞
func (this *Reader) writeDataChan(table *hbase.TableMeta, rowkey string, data map[string]interface{}) (isend bool) {

	this.Input() <- core.NewCollData("", data)
	table.TableAlreadyReadOffset += 1
	table.TableAlreadyKey[rowkey] = rowkey
	if table.TableAlreadyReadOffset >= table.TableDataCount {
		isend = true
	}
	return
}
