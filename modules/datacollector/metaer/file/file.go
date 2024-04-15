package file

import (
	"lego_datacollector/modules/datacollector/core"
	"lego_datacollector/modules/datacollector/metaer"
	"sync"
)

func NewMeta(name string) (meta IFileMetaData) {
	meta = &FileMetaData{
		MetaDataBase: metaer.MetaDataBase{
			Name: name,
		},
		file: &FileMeta{
			FileCacheData: make([]byte, 0),
		},
	}
	return
}

type IFileMetaData interface {
	core.IMetaerData
	GetFileName() string
	SetFileName(v string)
	GetFileAlreadyReadOffset() uint64
	GetFileCacheData() []byte
	ResetCacheData()
	GetFileSize() uint64
	SetFileSize(size uint64)
	SetFileCacheData(buf []byte) uint64
	SetFileAlreadyReadOffset(offset uint64)
	SetFileLastModifyTime(t int64)
	GetFileLastModifyTime() (t int64)
	SetFileLastCollectionTime(t int64)
	GetFileLastCollectionTime() (t int64)
}

type FileMeta struct {
	FileName               string
	FileSize               uint64
	FileLastModifyTime     int64
	FileAlreadyReadOffset  uint64
	FileLastCollectionTime int64
	FileCacheData          []byte
}

type FileMetaData struct {
	metaer.MetaDataBase
	lock sync.RWMutex
	file *FileMeta
}

func (this *FileMetaData) GetMetae() interface{} {
	return this.file
}
func (this *FileMetaData) GetFileName() string {
	this.lock.RLock()
	defer this.lock.RUnlock()
	return this.file.FileName
}
func (this *FileMetaData) SetFileName(v string) {
	this.lock.RLock()
	defer this.lock.RUnlock()
	this.file.FileName = v
}

func (this *FileMetaData) GetFileAlreadyReadOffset() uint64 {
	this.lock.RLock()
	defer this.lock.RUnlock()
	return this.file.FileAlreadyReadOffset
}

func (this *FileMetaData) GetFileCacheData() []byte {
	this.lock.RLock()
	defer this.lock.RUnlock()
	return this.file.FileCacheData
}

func (this *FileMetaData) ResetCacheData() {
	this.lock.Lock()
	defer this.lock.Unlock()
	this.file.FileCacheData = make([]byte, 0)
}

func (this *FileMetaData) GetFileSize() uint64 {
	this.lock.Lock()
	defer this.lock.Unlock()
	return this.file.FileSize
}

func (this *FileMetaData) SetFileSize(size uint64) {
	this.lock.Lock()
	defer this.lock.Unlock()
	this.file.FileSize = size
}

func (this *FileMetaData) SetFileAlreadyReadOffset(offset uint64) {
	this.lock.Lock()
	defer this.lock.Unlock()
	this.file.FileAlreadyReadOffset = offset
}
func (this *FileMetaData) SetFileLastModifyTime(t int64) {
	this.lock.Lock()
	defer this.lock.Unlock()
	this.file.FileLastModifyTime = t
}
func (this *FileMetaData) SetFileCacheData(buf []byte) uint64 {
	this.lock.Lock()
	defer this.lock.Unlock()
	this.file.FileCacheData = buf
	return uint64(len(this.file.FileCacheData))
}
func (this *FileMetaData) GetFileLastModifyTime() (t int64) {
	this.lock.RLock()
	defer this.lock.RUnlock()
	return this.file.FileLastModifyTime
}

func (this *FileMetaData) SetFileLastCollectionTime(t int64) {
	this.lock.Lock()
	defer this.lock.Unlock()
	this.file.FileLastCollectionTime = t
}
func (this *FileMetaData) GetFileLastCollectionTime() (t int64) {
	this.lock.RLock()
	defer this.lock.RUnlock()
	return this.file.FileLastCollectionTime
}
