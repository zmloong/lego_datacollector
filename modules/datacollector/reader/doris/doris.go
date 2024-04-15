package doris

import (
	"database/sql"
	"fmt"
	"lego_datacollector/comm"
	"lego_datacollector/modules/datacollector/core"
	msql "lego_datacollector/modules/datacollector/metaer/sql"
	"lego_datacollector/modules/datacollector/reader"
	. "lego_datacollector/utils/sql"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/liwei1dao/lego"
	lgcron "github.com/liwei1dao/lego/sys/cron"
	"github.com/liwei1dao/lego/sys/event"
	lgsql "github.com/liwei1dao/lego/sys/sql"
	"github.com/robfig/cron/v3"
)

type Reader struct {
	reader.Reader
	options      IOptions            //以接口对象传递参数 方便后期继承扩展
	meta         msql.ITableMetaData //愿数据
	sql          lgsql.ISys
	runId        cron.EntryID
	rungoroutine int32 //运行采集携程数
	isend        int32 //采集任务结束
	// lock         sync.Mutex           //执行锁
	wg        sync.WaitGroup       //用于等待采集器完全关闭
	colltask  chan *msql.TableMeta //文件采集任务
	schemas   map[string]string    //sql 类型处理
	source_ip string               //数据源地址
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
	if this.schemas, err = SchemaCheck(this.options.GetDoris_schema()); err != nil {
		return
	}
	this.source_ip = this.options.GetDoris_addr()[strings.Index(this.options.GetDoris_addr(), "(")+1 : strings.LastIndex(this.options.GetDoris_addr(), ":")]
	this.sql, err = lgsql.NewSys(lgsql.SetSqlType(lgsql.MySql), lgsql.SetSqlUrl(fmt.Sprintf("%s/%s", this.options.GetDoris_addr(), this.options.GetDoris_db_name())))
	return
}

func (this *Reader) Start() (err error) {
	err = this.Reader.Start()
	atomic.StoreInt32(&this.rungoroutine, 0)
	atomic.StoreInt32(&this.isend, 0)
	if !this.options.GetAutoClose() {
		if this.runId, err = lgcron.AddFunc(this.options.GetDoris_batch_intervel(), this.scanSql); err != nil {
			err = fmt.Errorf("lgcron.AddFunc corn:%s err:%v", this.options.GetDoris_batch_intervel(), err)
			return
		}
		if this.options.GetDoris_exec_onstart() {
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
	this.sql.Close()
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
		if this.options.GetDoris_collection_type() == StaticTableCollection { //静态表采集
			for _, k := range this.options.GetDoris_tables() {
				if v, ok := tables[k]; ok {
					if table, ok = this.meta.GetTableMeta(k); !ok {
						table = &msql.TableMeta{
							TableName:              k,
							TableDataCount:         v,
							TableAlreadyReadOffset: 0,
						}
						this.meta.SetTableMeta(k, table)
					}
					tablecount = this.gettablecount(k)
					if table.TableAlreadyReadOffset < tablecount { //有新的数据
						table.TableDataCount = tablecount
						collectiontables = append(collectiontables, table)
					}
				} else {
					this.Runner.Errorf("Reader Doris not found table：%s", k)
				}
			}
		} else if this.options.GetDoris_collection_type() == DynamicTableCollection { //动态表
			var (
				nextDate    string
				canIncrease bool
				dynamictame string
			)
			for _, k := range this.options.GetDoris_tables() {
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
					nextDate, canIncrease = GetNextDate(this.options.GetDoris_datetimeFormat(), this.options.GetDoris_startDate(), int(dynamictable.TableAlreadyReadOffset), this.options.GetDoris_dateUnit())
					if canIncrease {
						dynamictable.TableAlreadyReadOffset += uint64(this.options.GetDoris_dateGap())
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
							tablecount = this.gettablecount(dynamictame)
							if table.TableAlreadyReadOffset < tablecount { //有新的数据
								table.TableDataCount = tablecount
								collectiontables = append(collectiontables, table)
							}
						} else {
							this.Runner.Errorf("Reader Doris not found table：%s", dynamictame)
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
			this.Runner.Debugf("Reader Doris start new collection:%d", atomic.AddInt32(&this.rungoroutine, int32(procs)))
			this.wg.Add(procs)
			for i := 0; i < procs; i++ {
				go this.asyncollection(this.sql, this.colltask)
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
			event.TriggerEvent(comm.Event_WriteLog, this.Runner.Name(), this.Runner.InstanceId(), fmt.Sprintf("扫描目标数据源错误:%v!", err), 1, time.Now().Unix())
			go this.AutoClose(core.RunnerFailAutoClose)
		} else {
			atomic.StoreInt32(&this.isend, 0)
			if strings.Contains(err.Error(), "database is closed") && this.Runner.GetRunnerState() == core.Runner_Runing {
				this.sql.Close()
				if this.sql, err = lgsql.NewSys(lgsql.SetSqlType(lgsql.MySql), lgsql.SetSqlUrl(fmt.Sprintf("%s/%s", this.options.GetDoris_addr(), this.options.GetDoris_db_name()))); err != nil {
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

func (this *Reader) asyncollection(db lgsql.ISys, fmeta <-chan *msql.TableMeta) {
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
	var (
		data *sql.Rows
		sql  string
	)
	tables = make(map[string]uint64)
	sql = fmt.Sprintf("select table_name from information_schema.tables where table_schema = '%s' order by table_rows asc", this.options.GetDoris_db_name())
	if data, err = this.sql.Query(sql); err == nil {
		tablename := ""
		for data.Next() {
			if e := data.Scan(&tablename); e == nil {
				tables[tablename] = 0
			} else {
				this.Runner.Errorf("Reader scanddatabase %s err:%v", sql, e)
			}
		}
	}
	return
}

//采集数据表
func (this *Reader) collection_table(db lgsql.ISys, table *msql.TableMeta) (isend bool, err error) {
	var (
		sqlstr   string
		data     *sql.Rows
		columns  []string
		scanArgs []interface{}
		nochiced []bool
	)
	sqlstr = this.getsqlstr(table)
	// this.Runner.Debugf("sql:%s", sqlstr)
	if data, err = db.Query(sqlstr); err == nil {
		if columns, err = data.Columns(); err != nil {
			this.Runner.Errorf("collection_table err%v", err)
			return
		}
		schemas := make(map[string]string)
		for k, v := range this.schemas {
			schemas[k] = v
		}
		scanArgs, nochiced = GetInitScans(len(columns), data, schemas, this.Runner, table.TableName)
		isend, err = this.getAllDatas(table, data, scanArgs, columns, nochiced, schemas)
	}
	return
}

//获取表长度
func (this *Reader) gettablecount(tablename string) (count uint64) {
	var (
		sqlstr string
		err    error
		data   *sql.Rows
	)
	count = 0
	sqlstr = fmt.Sprintf("select count(*) from `%s`", tablename)
	if data, err = this.sql.Query(sqlstr); err != nil {
		this.Runner.Errorf("gettablecount %s sql:%s err:%v", tablename, sqlstr, err)
	} else {
		for data.Next() {
			if err := data.Scan(&count); err != nil {
				this.Runner.Errorf("gettablecount %s sql:%s err:%v", tablename, sqlstr, err)
			}
		}
	}
	return
}

//获取查询语句
func (this *Reader) getsqlstr(table *msql.TableMeta) (sqlstr string) {
	sqlstr = strings.Replace(this.options.GetDoris_sql(), "$TABLE$", table.TableName, -1)
	sqlstr = fmt.Sprintf("%s limit %d,%d;", sqlstr, table.TableAlreadyReadOffset, table.TableAlreadyReadOffset+uint64(this.options.GetDoris_limit_batch()))
	return
}

//读取数据
func (this *Reader) getAllDatas(table *msql.TableMeta, rows *sql.Rows, scanArgs []interface{}, columns []string, nochiced []bool, schemas map[string]string) (isend bool, err error) {
	for rows.Next() {
		err := rows.Scan(scanArgs...)
		if err != nil {
			this.Runner.Errorf("getAllDatas scan rows err:%v", err)
			continue
		}
		// this.Runner.Debugf("getAllDatas data:%v", scanArgs)
		var (
			data = make(map[string]interface{}, len(scanArgs))
		)
		for i := 0; i < len(scanArgs); i++ {
			_, err := ConvertScanArgs(data, scanArgs[i], columns[i], this.Runner, table.TableName, nochiced[i], this.schemas)
			if err != nil {
				this.Runner.Errorf("getAllDatas ConvertScanArgs err:%v", err)
			}
		}
		err = nil
		if len(data) <= 0 {
			continue
		}
		isend = this.writeDataChan(table, data)
	}
	return
}

//发送数据 数据量打的时候会有堵塞
func (this *Reader) writeDataChan(table *msql.TableMeta, data map[string]interface{}) (isend bool) {
	// this.Runner.Debugf("reder mysql table:%s writeDataChan start", table.TableName)
	// defer this.Runner.Debugf("reder mysql table:%s writeDataChan end", table.TableName)
	this.Input() <- core.NewCollData(this.source_ip, data)
	table.TableAlreadyReadOffset += 1
	if table.TableAlreadyReadOffset >= table.TableDataCount {
		isend = true
	}
	return
}
