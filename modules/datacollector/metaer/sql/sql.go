package sql

import (
	"fmt"
	"lego_datacollector/modules/datacollector/core"
	"lego_datacollector/modules/datacollector/metaer"
	"sync"
)

func NewMeta(name string) (meta ITableMetaData) {
	meta = &SqlMetaData{
		MetaDataBase: metaer.MetaDataBase{
			Name: name,
		},
		sqls: make(map[string]*TableMeta),
	}
	return
}

type ITableMetaData interface {
	core.IMetaerData
	GetTableMeta(tablename string) (table *TableMeta, ok bool)
	SetTableMeta(tablename string, table *TableMeta) (err error)
}

type TableMeta struct {
	TableName              string //表名
	TableDataCount         uint64 //表的数据长度
	TableAlreadyReadOffset uint64 //已采集数据长度
}

type SqlMetaData struct {
	metaer.MetaDataBase
	lock sync.RWMutex
	sqls map[string]*TableMeta
}

//map 结构对象序列化需要 map的指针
func (this *SqlMetaData) GetMetae() interface{} {
	return &this.sqls
}

func (this *SqlMetaData) GetTableMeta(tablename string) (table *TableMeta, ok bool) {
	this.lock.RLock()
	defer this.lock.RUnlock()
	table, ok = this.sqls[tablename]
	return
}

func (this *SqlMetaData) SetTableMeta(tablename string, table *TableMeta) (err error) {
	this.lock.Lock()
	defer this.lock.Unlock()
	if _, ok := this.sqls[tablename]; !ok {
		this.sqls[tablename] = table
	} else {
		err = fmt.Errorf("Meta:%s is Repeat %s", this.Name, tablename)
	}
	return
}
