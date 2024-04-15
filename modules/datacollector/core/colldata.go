package core

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"time"
	"unsafe"
)

//采集数据桶--------------------------------------------------------------------------------------------------------------------------------
func NewCollDataBucket(maxcap int, maxsize uint64) (bucket *CollDataBucket) {
	bucket = &CollDataBucket{
		bucket:   make([]ICollData, maxcap),
		maxcap:   maxcap,
		currcap:  0,
		maxsize:  maxsize,
		currsize: 0,
		full:     false,
		err:      nil,
	}
	return
}

type CollDataBucket struct {
	bucket   []ICollData //数据容器
	err      error       //全局错误
	maxcap   int         //最大数据容量 数据条数
	currcap  int         //当前容量
	maxsize  uint64      //最大内存大小
	currsize uint64      //当前内存大小
	full     bool        //是否就绪
}

//添加采集数据
func (this *CollDataBucket) AddCollData(data ICollData) (full bool) {
	this.bucket[this.currcap] = data
	this.currcap++
	this.currsize += data.GetSize()
	if this.currcap >= this.maxcap || this.currsize >= this.maxsize {
		this.full = true
	}
	return this.full
}

//是否是空
func (this *CollDataBucket) IsEmplty() bool {
	return this.currcap == 0
}

//是否满了
func (this *CollDataBucket) IsFull() bool {
	return this.full
}

//当前容量
func (this *CollDataBucket) CurrCap() int {
	return this.currcap
}

func (this *CollDataBucket) Items() (reslut []ICollData) {
	if this.IsEmplty() {
		return
	}
	reslut = this.bucket[:this.currcap]
	return
}

func (this *CollDataBucket) SuccItemsCount() (reslut int) {
	if this.IsEmplty() {
		return 0
	}
	reslut = 0
	for _, v := range this.bucket[:this.currcap] {
		if v.GetError() == nil {
			reslut++
		}
	}
	return
}

func (this *CollDataBucket) SuccItems() (reslut []ICollData) {
	if this.IsEmplty() {
		return
	}
	var n = 0
	reslut = make([]ICollData, this.currcap)
	for _, v := range this.bucket[:this.currcap] {
		if v.GetError() == nil {
			reslut[n] = v
			n++
		}
	}
	return reslut[:n]
}

func (this *CollDataBucket) ErrItems() (reslut []ICollData) {
	if this.IsEmplty() {
		return
	}
	var n = 0
	reslut = make([]ICollData, this.currcap)
	for _, v := range this.bucket[:this.currcap] {
		if v.GetError() != nil {
			reslut[n] = v
			n++
		}
	}
	return reslut[:n]
}

func (this *CollDataBucket) ErrItemsCount() (reslut int) {
	if this.IsEmplty() {
		return 0
	}
	reslut = 0
	for _, v := range this.bucket[:this.currcap] {
		if v.GetError() != nil {
			reslut++
		}
	}
	return
}

//编码到string
func (this *CollDataBucket) ToString(onlysucc bool) (value string, err error) {
	if this.IsEmplty() {
		err = errors.New("This Bucket is Emplty")
		return
	}
	buff := bytes.NewBuffer([]byte{})
	jsonEncoder := json.NewEncoder(buff)
	jsonEncoder.SetEscapeHTML(false)
	if !onlysucc {
		if err = jsonEncoder.Encode(this.SuccItems()); err != nil {
			return
		}
	} else {
		if err = jsonEncoder.Encode(this.Items()); err != nil {
			return
		}
	}
	value = buff.String()
	return
}

//重构json序列化接口
func (this *CollDataBucket) MarshalJSON() (value []byte, err error) {
	buff := bytes.NewBuffer([]byte{})
	jsonEncoder := json.NewEncoder(buff)
	jsonEncoder.SetEscapeHTML(false)
	err = jsonEncoder.Encode(this.Items())
	value = buff.Bytes()
	return
}

//重置
func (this *CollDataBucket) Reset() {
	this.full = false
	this.currcap = 0
	this.currsize = 0
}

//克隆
func (this *CollDataBucket) Clone() (result *CollDataBucket) {
	result = &CollDataBucket{
		bucket:   make([]ICollData, len(this.bucket)),
		err:      this.err,
		maxcap:   this.maxcap,
		currcap:  this.currcap,
		maxsize:  this.maxsize,
		currsize: this.currsize,
		full:     this.full,
	}
	for i, v := range this.bucket {
		result.bucket[i] = v
	}
	return
}

func (this *CollDataBucket) SetError(err error) {
	this.err = err
}

//放回当前错误
func (this *CollDataBucket) Error() (err error) {
	if this.err != nil {
		return this.err
	}
	var buf bytes.Buffer
	for i, v := range this.bucket[:this.currcap] {
		if v.GetError() != nil {
			buf.WriteString(fmt.Sprintf("index:%d,err:%v,value:%v\n", i, v.GetError(), v.GetValue()))
		}
	}
	if buf.Len() > 0 {
		return errors.New(buf.String())
	}
	return nil
}

//采集数据结构----------------------------------------------------------------------------------------------------------------------
func NewCollData(source string, value interface{}) ICollData {
	return &CollData{
		source: source,
		time:   time.Now().UnixNano() / 1e6,
		data:   make(map[string]interface{}),
		value:  value,
	}
}

//采集 String 数据
type CollData struct {
	source string                 //数据来源
	time   int64                  //采集时间
	data   map[string]interface{} //整理数据
	value  interface{}            //采集数据
	err    error                  //错误信息
}

///错误信息
func (this *CollData) SetError(err error) {
	this.err = err
}

///错误信息
func (this *CollData) GetError() error {
	return this.err
}

///数据来源
func (this *CollData) GetSource() string {
	return this.source
}

///采集时间
func (this *CollData) GetTime() int64 {
	return this.time
}

///采集整理数据
func (this *CollData) GetData() map[string]interface{} {
	return this.data
}

///采集数据
func (this *CollData) GetValue() interface{} {
	return this.value
}

func (this *CollData) encodeMessage() {
	if v, ok := this.data[KeyRaw]; ok {
		switch v.(type) {
		case map[string]interface{}:
			buff := bytes.NewBuffer([]byte{})
			jsonEncoder := json.NewEncoder(buff)
			jsonEncoder.SetEscapeHTML(false)
			jsonEncoder.Encode(v)
			this.data[KeyRaw] = buff.String()
			break
		case []interface{}:
			buff := bytes.NewBuffer([]byte{})
			jsonEncoder := json.NewEncoder(buff)
			jsonEncoder.SetEscapeHTML(false)
			jsonEncoder.Encode(v)
			this.data[KeyRaw] = buff.String()
			break
		case string:
			this.data[KeyRaw] = v
		default:
			buff := bytes.NewBuffer([]byte{})
			jsonEncoder := json.NewEncoder(buff)
			jsonEncoder.SetEscapeHTML(false)
			jsonEncoder.Encode(v)
			this.data[KeyRaw] = buff.String()
			break
		}
	}
}

//编码到string
func (this *CollData) ToString() (value string, err error) {
	this.encodeMessage() //先序列化Message
	buff := bytes.NewBuffer([]byte{})
	jsonEncoder := json.NewEncoder(buff)
	jsonEncoder.SetEscapeHTML(false)
	jsonEncoder.Encode(this.data)
	value = buff.String()
	return
}

//重构json序列化接口
func (this *CollData) MarshalJSON() (value []byte, err error) {
	this.encodeMessage() //先序列化Message
	buff := bytes.NewBuffer([]byte{})
	jsonEncoder := json.NewEncoder(buff)
	jsonEncoder.SetEscapeHTML(false)
	err = jsonEncoder.Encode(this.data)
	value = buff.Bytes()
	return
}

///数据对象大小
func (this *CollData) GetSize() uint64 {
	siez := int(unsafe.Sizeof(this))
	return uint64(siez)
}
