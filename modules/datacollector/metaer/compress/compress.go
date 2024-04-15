package folder

import (
	"fmt"
	"lego_datacollector/modules/datacollector/core"
	"lego_datacollector/modules/datacollector/metaer"
	"sync"
)

func NewMeta(name string) (meta IFolderMetaData) {
	meta = &FolderMetaData{
		MetaDataBase: metaer.MetaDataBase{
			Name: name,
		},
		files: make(map[string]*CopresFileMeta),
	}
	return
}

type IFolderMetaData interface {
	core.IMetaerData
	GetFile(filename string) (file *CopresFileMeta, ok bool)
	GetFiles() (files map[string]*CopresFileMeta)
	SetFile(filename string, file *CopresFileMeta) (err error)
}
type FolderMetaData struct {
	metaer.MetaDataBase
	lock  sync.RWMutex
	files map[string]*CopresFileMeta
}
type CopresFileMeta struct {
	//Compres
	CompresName           string
	CompresSize           uint64
	CompresTotalFiles     uint64 //文件总数
	CompresAlreadyCollect uint64 //已采集完成文件数
	//file
	FileName               string
	FileSize               uint64
	FileLastModifyTime     int64
	FileAlreadyReadOffset  uint64
	FileLastCollectionTime int64
	FileCacheData          []byte
}

//map 结构对象序列化需要 map的指针
func (this *FolderMetaData) GetMetae() interface{} {
	return &this.files
}

func (this *FolderMetaData) GetFile(filename string) (file *CopresFileMeta, ok bool) {
	this.lock.RLock()
	defer this.lock.RUnlock()
	file, ok = this.files[filename]
	return
}

func (this *FolderMetaData) GetFiles() (files map[string]*CopresFileMeta) {
	this.lock.RLock()
	files = make(map[string]*CopresFileMeta)
	for k, v := range this.files {
		files[k] = v
	}
	this.lock.RUnlock()
	return
}

func (this *FolderMetaData) SetFile(filename string, file *CopresFileMeta) (err error) {
	this.lock.Lock()
	defer this.lock.Unlock()
	if _, ok := this.files[filename]; !ok {
		this.files[filename] = file
	} else {
		err = fmt.Errorf("Meta:%s is Repeat %s", this.Name, filename)
	}
	return
}
