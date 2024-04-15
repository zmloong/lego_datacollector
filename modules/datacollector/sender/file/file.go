package file

import (
	"fmt"
	"io"
	"lego_datacollector/modules/datacollector/core"
	"lego_datacollector/modules/datacollector/sender"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/lestrrat-go/strftime"
	"github.com/liwei1dao/lego/utils/flietools"
)

type (
	writer struct {
		inUseStatus int32 // Note: 原子操作
		lastWrote   int32
		wc          io.WriteCloser
	}
	Sender struct {
		sender.Sender
		options IOptions
		pattern *strftime.Strftime
		size    int
		lock    sync.RWMutex
		writers map[string]*writer
	}
)

//writer--------------------------------------------------------------------------------------------
// IsBusy 当任务数为 0 时返回 false 表示可以被安全回收，否则返回 true
func (w *writer) IsBusy() bool {
	return atomic.LoadInt32(&w.inUseStatus) > 0
}

// SetBusy 将 writer 的正在处理任务数添加 1，使对象表示正忙，防止因 lastWrote 过期被意外回收
func (w *writer) SetBusy() {
	atomic.AddInt32(&w.inUseStatus, 1)
}

// SetIdle 将 writer 的正在处理任务数减去 1
func (w *writer) SetIdle() {
	atomic.AddInt32(&w.inUseStatus, -1)
}

func (w *writer) Write(b []byte) (int, error) {
	atomic.StoreInt32(&w.lastWrote, int32(time.Now().Unix()))
	return w.wc.Write(b)
}

func (w *writer) Close() error {
	return w.wc.Close()
}

//Sender----------------------------------------------------------------------------------------------------------------
func (this *Sender) Type() string {
	return SenderType
}
func (this *Sender) Init(rer core.IRunner, ser core.ISender, options core.ISenderOptions) (err error) {
	this.Sender.Init(rer, ser, options)
	this.options = options.(IOptions)
	dir := filepath.Dir(this.options.GetSendfilePath())
	if !flietools.IsExist(dir) {
		if err = os.MkdirAll(dir, os.ModePerm); err != nil {
			err = fmt.Errorf("Sender file SendfilePath:%s MkdirErr:%v", this.options.GetSendfilePath(), err)
			return
		}
	}
	if this.pattern, err = strftime.New(this.options.GetSendfilePath()); err != nil {
		return
	}
	this.writers = make(map[string]*writer)
	return
}

func (this *Sender) Close() (err error) {
	this.lock.Lock()
	defer this.lock.Unlock()

	for name, w := range this.writers {
		if err = w.Close(); err != nil {
			return
		}
		delete(this.writers, name)
	}
	return
}

func (this *Sender) Send(pipeId int, bucket core.ICollDataBucket, params ...interface{}) {
	var (
		filename  string
		databytes string
		err       error
	)
	filename = this.pattern.FormatString(time.Now())
	if databytes, err = bucket.ToString(true); err != nil {
		bucket.SetError(err)
		return
	}
	if _, err = this.Write(filename, []byte(databytes)); err != nil {
		bucket.SetError(err)
	}
	this.Sender.Send(pipeId, bucket)
	return
}

// Write 向指定文件写入数据，调用者不需要关心文件是否存在，函数自身会协调打开的文件句柄
func (this *Sender) Write(filename string, b []byte) (int, error) {
	w, err := this.NewWriter(filename)
	if err != nil {
		return 0, fmt.Errorf("new writer: %v", err)
	}
	defer w.SetIdle()
	return w.Write(b)
}

// Get 根据文件名称返回对应的 writer，如果不存在则返回 nil，调用者在用完对象后必须调用 SetIdle 便于回收
func (this *Sender) Get(filename string) *writer {
	this.lock.RLock()
	defer this.lock.RUnlock()

	w := this.writers[filename]
	if w != nil {
		w.SetBusy()
	}
	return w
}

// NewWriter 会根据文件名新建一个文件句柄，如果打开的句柄数已经达到限制，则会关闭最不活跃的句柄
func (this *Sender) NewWriter(filename string) (w *writer, _ error) {
	w = this.Get(filename)
	if w != nil {
		return w, nil
	}

	dir := filepath.Dir(filename)
	if !flietools.IsExist(dir) {
		if err := os.MkdirAll(dir, os.ModePerm); err != nil {
			return nil, fmt.Errorf("create parent directory: %v", err)
		}
	}
	f, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("open file: %v", err)
	}
	w = &writer{
		wc: f,
	}
	w.SetBusy()

	this.lock.Lock()
	defer this.lock.Unlock()

	this.writers[filename] = w

	// 关闭并删除 (total - size) 个最不活跃的句柄，因为可能同时创建多个句柄而当时无法确认谁是最不活跃的
	for i := 1; i <= len(this.writers)-this.size; i++ {
		var expiredFilename string
		lastWrite := int32(time.Now().Unix())
		for name, w := range this.writers {
			currentLastWrite := atomic.LoadInt32(&w.lastWrote)
			if currentLastWrite == 0 || w.IsBusy() {
				continue // 暂时忽略刚创建和正忙的句柄
			}

			if currentLastWrite < lastWrite {
				expiredFilename = name
				lastWrite = currentLastWrite
			}
		}

		if len(expiredFilename) > 0 {
			if err := this.writers[expiredFilename].Close(); err != nil {
				this.Runner.Errorf("Failed to close expired file writer %q: %v", expiredFilename, err)
			}
			delete(this.writers, expiredFilename)
		}
	}
	return w, nil
}

func (this *Sender) getPartitionFolder(nowStr string, idx int) string {
	base := filepath.Base(nowStr)
	dir := filepath.Dir(nowStr)
	return filepath.Join(dir, "partition"+strconv.Itoa(idx), base)
}
