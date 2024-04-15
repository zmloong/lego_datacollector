package datacollector

import (
	"fmt"
	"lego_datacollector/comm"
	"lego_datacollector/modules/datacollector/core"
	"lego_datacollector/modules/datacollector/metaer"
	"lego_datacollector/modules/datacollector/parser"
	"lego_datacollector/modules/datacollector/reader"
	"lego_datacollector/modules/datacollector/sender"
	"lego_datacollector/modules/datacollector/tranfroms"
	"lego_datacollector/sys/db"
	"sync"
	"sync/atomic"

	_ "lego_datacollector/modules/datacollector/parser/builtin"
	_ "lego_datacollector/modules/datacollector/reader/builtin"
	_ "lego_datacollector/modules/datacollector/sender/builtin"
	_ "lego_datacollector/modules/datacollector/tranfroms/builtin"

	"time"

	"github.com/liwei1dao/lego"
	"github.com/liwei1dao/lego/sys/event"
	"github.com/liwei1dao/lego/sys/log"
	"github.com/liwei1dao/lego/sys/redis"
)

func newRunner(conf *comm.RunnerConfig, sId, sIp string) (runner core.IRunner, err error) {
	// var (
	// 	rlog log.ILog
	// )
	// if rlog, err = log.NewSys(
	// 	log.SetFileName(fmt.Sprintf("./log/%s.log", conf.Name)),
	// 	log.SetLoglayer(2),
	// 	log.SetLoglevel(loglevel),   //与服务端log配置保持一致
	// 	log.SetDebugMode(debugmode), //与服务端log配置保持一致
	// ); err != nil {
	// 	return
	// }
	runner = &Runner{
		conf:      conf,
		serviceId: sId,
		serviceIp: sIp,
		// log:         rlog,
		readerPope:     make(chan core.ICollData, conf.DataChanSize),
		parserPope:     make(chan core.ICollDataBucket),
		transformsPope: make([]chan core.ICollDataBucket, len(conf.TranfromssConfig)),
		sendersPope:    make(map[string]chan core.ICollDataBucket),
		closesignal:    make(chan struct{}),
		statisticsIn:   make(chan *core.Statistics, 10),
		statisticsOut:  make(chan *core.Statistics, 10),
		ticker:         time.NewTicker(time.Duration(conf.MaxBatchInterval) * time.Second),
		readerCnt:      0,
		// statisticinfo: &comm.RunnerRuninfo{
		// 	RName:      conf.Name,
		// 	RunService: map[string]struct{}{sId: {}},
		// 	ReaderCnt:  0,
		// 	SenderCnt:  make(map[string]int64),
		// },
	}
	if err = runner.Init(); err != nil {
		runner.Close(core.Runner_Initing, fmt.Sprintf("Init err:%v", err))
	}
	return
}

type Runner struct {
	conf *comm.RunnerConfig
	// log         log.ILog
	serviceId      string
	serviceIp      string
	readerPope     chan core.ICollData
	parserPope     chan core.ICollDataBucket
	transformsPope []chan core.ICollDataBucket
	sendersPope    map[string]chan core.ICollDataBucket
	statisticsIn   chan *core.Statistics
	statisticsOut  chan *core.Statistics
	metaer         core.IMetaer
	reader         core.IReader
	parser         core.IParser
	transforms     []core.ITransforms
	senders        []core.ISender
	state          int32 //采集器状态
	closesignal    chan struct{}
	ticker         *time.Ticker //定时器
	currbucket     core.ICollDataBucket
	readerCnt      int64
	// slock          sync.RWMutex
	// statisticinfo  *comm.RunnerRuninfo
}

func (this *Runner) Name() (name string) {
	name = this.conf.Name
	return
}
func (this *Runner) BusinessType() comm.BusinessType {
	return this.conf.BusinessType
}
func (this *Runner) InstanceId() string {
	return this.conf.InstanceId
}

func (this *Runner) ServiceId() string {
	return this.serviceId
}
func (this *Runner) ServiceIP() string {
	return this.serviceIp
}
func (this *Runner) MaxProcs() (maxprocs int) {
	return this.conf.MaxProcs
}
func (this *Runner) MaxCollDataSzie() (msgmaxsize uint64) {
	return this.conf.MaxCollDataSzie
}

//获取采集器状态 原子操作
func (this *Runner) GetRunnerState() core.RunnerState {
	return core.RunnerState(atomic.LoadInt32(&this.state))
}

func (this *Runner) Init() (err error) {
	atomic.StoreInt32(&this.state, int32(core.Runner_Initing))
	this.currbucket = core.NewCollDataBucket(this.conf.MaxBatchLen, uint64(this.conf.MaxBatchSize))
	//初始化元数据管理
	if this.metaer, err = metaer.NewMetaer(this); err != nil {
		this.Errorf("NewRunner NewMetaer fail:%v", err)
		return
	}
	if this.reader, err = reader.NewReader(this, this.conf.ReaderConfig); err != nil {
		this.Errorf("NewRunner NewReader fail:%v", err)
		return
	}
	if this.parser, err = parser.NewParser(this, this.conf.ParserConf); err != nil {
		this.Errorf("NewRunner NewParser fail:%v", err)
		return
	}
	if this.transforms, err = tranfroms.NewTransforms(this, this.conf.TranfromssConfig); err != nil {
		this.Errorf("NewRunner NewTransforms fail:%v", err)
		return
	}
	for i, _ := range this.transforms {
		this.transformsPope[i] = make(chan core.ICollDataBucket)
	}
	if this.senders, err = sender.NewSender(this, this.conf.SendersConfig); err != nil {
		this.Errorf("NewRunner NewSender fail:%v", err)
		return
	}
	for _, v := range this.senders {
		this.sendersPope[v.GetType()] = make(chan core.ICollDataBucket)
		// this.statisticinfo.SenderCnt[v.GetType()] = 0
	}
	return
}

func (this *Runner) Start() (err error) {
	atomic.StoreInt32(&this.state, int32(core.Runner_Starting))
	for i, v := range this.senders {
		if err = v.Start(); err != nil {
			this.Errorf("启动Runner Senders:%v fail:%v", i, err)
			return
		}
	}
	for i, v := range this.transforms {
		if err = v.Start(); err != nil {
			this.Errorf("启动Runner Transforms:%v fail:%v", i, err)
			return
		}
	}
	if err = this.parser.Start(); err != nil {
		this.Errorf("启动Runner Parser fail:%v", err)
		return
	}
	if err = this.reader.Start(); err != nil {
		this.Errorf("启动Runner Reader fail:%v", err)
		return
	}
	if err = db.WriteRunnerRuninfo(&comm.RunnerRuninfo{
		RName:      this.conf.Name,
		InstanceId: this.conf.InstanceId,
		RunService: map[string]struct{}{this.serviceId: {}},
		RecordTime: time.Now().Unix(),
		ReaderCnt:  0,
		ParserCnt:  0,
		SenderCnt:  make(map[string]int64),
	}); err != nil { //写入运行信息
		return
	}
	this.Infof("启动Runner:%s", this.conf.Name)
	atomic.StoreInt32(&this.state, int32(core.Runner_Runing))
	go this.statisticRunner()
	go this.run()
	return
}

///驱动工作 提供外部调度器使用
func (this *Runner) Drive() (err error) {
	err = this.reader.Drive()
	return
}

///关闭逻辑  先关闭读取器 在关闭调度 最后关闭解析以及发送
func (this *Runner) Close(_state core.RunnerState, closemsg string) (err error) {
	state := atomic.LoadInt32(&this.state)
	if state == int32(core.Runner_Stoped) {
		err = Error_RunnerStoped
		return
	}
	if state == int32(core.Runner_Stoping) {
		err = Error_RunnerStoping
		return
	}
	atomic.StoreInt32(&this.state, int32(core.Runner_Stoping))
	defer lego.Recover(fmt.Sprintf("%s Close:%s", this.conf.Name, closemsg))
	this.Infof("Runner Close : %s", closemsg)
	if this.reader != nil { //初始化启动失败 也需要走一次close
		if err = this.reader.Close(); err != nil {
			this.Errorf("Runner reader Close err: %v", err)
			return
		}
	}
	//采集器已启动 关闭处理流
	if _state == core.Runner_Runing {
		this.Infof("Runner Close reader succ")
		this.closesignal <- struct{}{}
		close(this.closesignal)
		this.Infof("Runner Close closesignal succ")
		close(this.readerPope)
		this.Infof("Runner readerPope close")
		time.Sleep(time.Second) //等待一秒钟 清理剩余数据
	}

	close(this.parserPope)
	if this.parser != nil {
		if err = this.parser.Close(); err != nil {
			this.Errorf("Runner parser Close err: %v", err)
			return
		}
	}
	for i, v := range this.transforms {
		if v != nil {
			close(this.transformsPope[i])
			if err = v.Close(); err != nil {
				this.Errorf("Runner transforms Close err: %v", err)
				return
			}
		}
	}
	for _, v := range this.senders {
		if v != nil {
			close(this.sendersPope[v.GetType()])
			if err = v.Close(); err != nil {
				this.Errorf("Runner sender Close err: %v", err)
				return
			}
		}
	}
	if this.metaer != nil {
		if err = this.metaer.Close(); err != nil {
			this.Errorf("Runner metaer Close err: %v", err)
			return
		}
	}
	close(this.statisticsIn)
	close(this.statisticsOut)
	db.CleanRunnerRuninfo(this.conf.Name)
	atomic.StoreInt32(&this.state, int32(core.Runner_Stoped))
	event.TriggerEvent(comm.Event_UpdatRunner) //通知更新采集服务列表
	if _state == core.Runner_Runing {
		if closemsg == core.RunnerSuccAutoClose {
			event.TriggerEvent(comm.Event_RunnerAutoClose, this.conf.Name, this.conf.InstanceId, true)
		} else if closemsg == core.RunnerFailAutoClose {
			event.TriggerEvent(comm.Event_RunnerAutoClose, this.conf.Name, this.conf.InstanceId, false)
		}
	}
	return
}

func (this *Runner) Debug(msg string, fields ...log.Field) {
	log.Debug(msg, fields...)
}
func (this *Runner) Info(msg string, fields ...log.Field) {
	log.Info(msg, fields...)
}
func (this *Runner) Warn(msg string, fields ...log.Field) {
	log.Warn(msg, fields...)
}
func (this *Runner) Error(msg string, fields ...log.Field) {
	log.Error(msg, fields...)
}
func (this *Runner) Panic(msg string, fields ...log.Field) {
	log.Panic(msg, fields...)
}
func (this *Runner) Fatal(msg string, fields ...log.Field) {
	log.Fatal(msg, fields...)
}
func (this *Runner) Debugf(format string, a ...interface{}) {
	// this.log.Debugf(format, a...)
	this.Debug(fmt.Sprintf(format, a...), log.Field{Key: "rid", Value: this.Name()})
}
func (this *Runner) Infof(format string, a ...interface{}) {
	// this.log.Infof(format, a...)
	this.Info(fmt.Sprintf(format, a...), log.Field{Key: "rid", Value: this.Name()})
}
func (this *Runner) Warnf(format string, a ...interface{}) {
	// this.log.Warnf(format, a...)
	this.Warn(fmt.Sprintf(format, a...), log.Field{Key: "rid", Value: this.Name()})
}
func (this *Runner) Errorf(format string, a ...interface{}) {
	// this.log.Errorf(format, a...)
	this.Error(fmt.Sprintf(format, a...), log.Field{Key: "rid", Value: this.Name()})
}
func (this *Runner) Panicf(format string, a ...interface{}) {
	// this.log.Panicf(format, a...)
	this.Panic(fmt.Sprintf(format, a...), log.Field{Key: "rid", Value: this.Name()})
}
func (this *Runner) Fatalf(format string, a ...interface{}) {
	// this.log.Fatalf(format, a...)
	this.Fatal(fmt.Sprintf(format, a...), log.Field{Key: "rid", Value: this.Name()})
}
func (this *Runner) Metaer() core.IMetaer {
	return this.metaer
}
func (this *Runner) Reader() core.IReader {
	return this.reader
}

func (this *Runner) ReaderPipe() chan<- core.ICollData {
	return this.readerPope
}
func (this *Runner) StatisticsIn() chan<- *core.Statistics {
	return this.statisticsIn
}
func (this *Runner) StatisticsOut() chan<- *core.Statistics {
	return this.statisticsOut
}
func (this *Runner) Push_ParserPipe(bucket core.ICollDataBucket) {
	this.parserPope <- bucket
}
func (this *Runner) ParserPipe() <-chan core.ICollDataBucket {
	return this.parserPope
}
func (this *Runner) Push_TransformsPipe(bucket core.ICollDataBucket) {
	if len(this.transformsPope) > 0 {
		this.transformsPope[0] <- bucket
	} else {
		this.Push_SenderPipe(bucket)
	}
}
func (this *Runner) TransformsPipe(index int) (pipe <-chan core.ICollDataBucket, ok bool) {
	if len(this.transformsPope) > index {
		pipe = this.transformsPope[index]
		ok = true
	} else {
		ok = false
	}
	return
}
func (this *Runner) Push_NextTransformsPipe(index int, bucket core.ICollDataBucket) {
	if index >= len(this.transformsPope) {
		this.Push_SenderPipe(bucket)
	} else {
		this.transformsPope[index] <- bucket
	}
}
func (this *Runner) Push_SenderPipe(bucket core.ICollDataBucket) {
	for _, v := range this.sendersPope {
		v <- bucket
	}
}
func (this *Runner) SenderPipe(stype string) (pipe <-chan core.ICollDataBucket, ok bool) {
	pipe, ok = this.sendersPope[stype]
	return
}

//保活
func (this *Runner) SyncRunnerInfo() {
	readerCnt := atomic.SwapInt64(&this.readerCnt, 0)
	scnts := make(map[string]int64)
	for _, v := range this.senders {
		scnts[v.GetType()] = v.ReadAnResetCnt()
	}
	// this.slock.Lock()
	// this.statisticinfo.ReaderCnt += readerCnt
	// for k, _ := range this.statisticinfo.SenderCnt {
	// 	this.statisticinfo.SenderCnt[k] += scnts[k]
	// }
	// this.slock.Unlock()
	for i := 0; i < 3; i++ {
		if err := db.WriteRunnerRuninfo(&comm.RunnerRuninfo{
			RName:      this.conf.Name,
			InstanceId: this.conf.InstanceId,
			RunService: map[string]struct{}{this.serviceId: {}},
			ReaderCnt:  readerCnt,
			SenderCnt:  scnts,
		}); err == nil { //写入运行信息
			return
		} else if err == redis.TxFailedErr { //乐观锁丢失 重试
			time.Sleep(time.Second * 1)
			continue
		} else { //其他错误输出
			this.Errorf("SyncRunnerInfo err:%v", err)
			return
		}
	}
}

//统计
func (this *Runner) StatisticRunner() {
	// for i := 0; i < 3; i++ {
	// 	if err := db.SyncRunnerStatistics(this.statisticinfo); err == nil { //写入运行信息
	// 		this.slock.Lock()
	// 		this.statisticinfo.ReaderCnt = 0
	// 		for k, _ := range this.statisticinfo.SenderCnt {
	// 			this.statisticinfo.SenderCnt[k] = 0
	// 		}
	// 		this.slock.Unlock()
	// 		return
	// 	} else if err == redis.TxFailedErr { //乐观锁丢失 重试
	// 		time.Sleep(time.Second * 1)
	// 		continue
	// 	} else { //其他错误输出
	// 		this.Errorf("StatisticRunner err:%v", err)
	// 		return
	// 	}
	// }
}

//---------------------------------------------------------------------------------------------------------------
func (this *Runner) run() {
	defer this.ticker.Stop()
locp:
	for {
		select {
		case <-this.closesignal:
			if !this.currbucket.IsEmplty() { //清理数据
				this.submitbucket()
			}
			break locp
		case data := <-this.readerPope:
			if this.currbucket.AddCollData(data) { //满了
				this.submitbucket()
			}
		case <-this.ticker.C:
			if !this.currbucket.IsEmplty() { //当前数据桶 不为空
				this.submitbucket()
			}
		}
	}
	this.Debugf("Runner exit run")
}

///统计数据异步存储
func (this *Runner) statisticRunner() {
	var (
		in      map[string]int64            = make(map[string]int64)
		out     map[string]map[string]int64 = make(map[string]map[string]int64)
		tempin  map[string]int64            = make(map[string]int64)
		tempout map[string]map[string]int64 = make(map[string]map[string]int64)
		inlock  sync.Mutex
		outlock sync.Mutex
	)
	go func() {
		for v := range this.statisticsIn {
			inlock.Lock()
			if _, ok := in[v.Name]; !ok {
				in[v.Name] = 0
			}
			in[v.Name] += v.Statistics
			inlock.Unlock()
		}
	}()
	go func() {
		for v := range this.statisticsOut {
			outlock.Lock()
			if _, ok := out[v.Name]; !ok {
				out[v.Name] = map[string]int64{}
			}
			if _, ok := out[v.Name][v.Type]; !ok {
				out[v.Name][v.Type] = 0
			}
			out[v.Name][v.Type] += v.Statistics
			outlock.Unlock()
		}
	}()
	for {
		if atomic.LoadInt32(&this.state) != int32(core.Runner_Stoped) {
			tempin = make(map[string]int64)
			tempout = make(map[string]map[string]int64)
			inlock.Lock()
			outlock.Lock()
			for k, v := range in {
				if v > 0 {
					tempin[k] = v
					in[k] = 0
				}
			}
			for k, v := range out {
				tempout[k] = map[string]int64{}
				for k1, v1 := range v {
					if v1 > 0 {
						tempout[k][k1] = v1
						out[k][k1] = 0
					}
				}
			}
			inlock.Unlock()
			outlock.Unlock()
			for k, v := range tempin {
				db.RealTimeStatisticsIn(k, v)
			}
			for k, v := range tempout {
				for k1, v1 := range v {
					db.RealTimeStatisticsOut(k, k1, v1)
				}
			}
			time.Sleep(time.Second)
		} else {
			break
		}
	}
	this.Debugf("Runner exit statisticRunner")
}

//这里不再做额外判断 直接处理
func (this *Runner) submitbucket() {
	var (
		temp core.ICollDataBucket
	)
	temp = this.currbucket
	atomic.AddInt64(&this.readerCnt, int64(temp.CurrCap()))
	if this.conf.BusinessType == comm.Collect {
		// db.RealTimeStatisticsIn(this.conf.Name, int64(temp.CurrCap()))
		this.StatisticsIn() <- &core.Statistics{Name: this.conf.Name, Statistics: int64(temp.CurrCap())}
	}
	this.currbucket = core.NewCollDataBucket(this.conf.MaxBatchLen, uint64(this.conf.MaxBatchSize)) //重新申请桶
	this.Push_ParserPipe(temp)
}
