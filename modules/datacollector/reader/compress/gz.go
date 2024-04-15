package compress

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"io"
	compress "lego_datacollector/modules/datacollector/metaer/compress"
	"regexp"
)

//解压文件
func (this *Reader) gz_Unpack(r io.Reader) (tr *tar.Reader, err error) {
	// gzip read
	gr, err := gzip.NewReader(r)
	if err != nil {
		return
	}
	defer gr.Close()
	// tar read
	tr = tar.NewReader(gr)
	return
}

//解压文件
func (this *Reader) unpackGzip(fr bytes.Buffer) (tr *tar.Reader, err error) {
	gr, err := gzip.NewReader(&fr)
	if err != nil {
		return
	}
	defer gr.Close()
	tr = tar.NewReader(gr)
	return
}

//获取压缩包文件meta记录信息
func (this *Reader) gz_GetCompressMeta(filepath string, cpsf *compress.CopresFileMeta, tr *tar.Reader) (filecount uint64, err error) {
	for {
		var h *tar.Header
		h, err = tr.Next()
		if err != nil {
			return
		}
		//文件结尾
		if err == io.EOF {
			break
		}
		//过滤掉文件夹+正则过滤文件名
		matched, err := regexp.MatchString(this.options.GetCompress_inter_regularrules(), h.Name)
		if h.Size == 0 || (!matched && err == nil) {
			continue
		}

		filemeta := &compress.CopresFileMeta{
			FileName:              filepath + "/" + h.Name,
			FileSize:              uint64(h.Size),
			FileLastModifyTime:    h.ModTime.Unix(), //只有采集完毕才同步时间
			FileAlreadyReadOffset: 0,
			FileCacheData:         make([]byte, 0),
		}
		this.meta.SetFile(filemeta.FileName, filemeta)
		filecount++
	}

	return
}
