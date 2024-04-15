package folder

import (
	"fmt"
	"lego_datacollector/modules/datacollector/core"
	"lego_datacollector/modules/datacollector/metaer"
	"lego_datacollector/modules/datacollector/metaer/file"
	"sync"
)

func NewMeta(name string) (meta IFolderMetaData) {
	meta = &FolderMetaData{
		MetaDataBase: metaer.MetaDataBase{
			Name: name,
		},
		files: make(map[string]*file.FileMeta),
	}
	return
}

type IFolderMetaData interface {
	core.IMetaerData
	GetFile(filename string) (file *file.FileMeta, ok bool)
	GetFiles() (files map[string]*file.FileMeta)
	SetFile(filename string, file *file.FileMeta) (err error)
}
type FolderMetaData struct {
	metaer.MetaDataBase
	lock  sync.RWMutex
	files map[string]*file.FileMeta
}

//map 结构对象序列化需要 map的指针
func (this *FolderMetaData) GetMetae() interface{} {
	return &this.files
}

func (this *FolderMetaData) GetFile(filename string) (file *file.FileMeta, ok bool) {
	this.lock.RLock()
	defer this.lock.RUnlock()
	file, ok = this.files[filename]
	return
}

func (this *FolderMetaData) GetFiles() (files map[string]*file.FileMeta) {
	this.lock.RLock()
	files = make(map[string]*file.FileMeta)
	for k, v := range this.files {
		files[k] = v
	}
	this.lock.RUnlock()
	return
}

func (this *FolderMetaData) SetFile(filename string, file *file.FileMeta) (err error) {
	this.lock.Lock()
	defer this.lock.Unlock()
	if _, ok := this.files[filename]; !ok {
		this.files[filename] = file
	} else {
		err = fmt.Errorf("Meta:%s is Repeat %s", this.Name, filename)
	}
	return
}
