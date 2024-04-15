package compress

import (
	"archive/tar"
	"bufio"
	"bytes"
	"fmt"
	"io"
	"lego_datacollector/modules/datacollector/core"
	compress "lego_datacollector/modules/datacollector/metaer/compress"
	"regexp"
	"sync/atomic"
	"time"

	"github.com/beevik/etree"
	"github.com/jlaffaye/ftp"
	"github.com/liwei1dao/lego"
	"github.com/ricardolonga/jsongo"
)

//高并发采集服务
func (this *Reader) ftp_collection() {
	this.lock.Lock()
	defer this.lock.Unlock()
	var (
		cfiles map[string]*compress.CopresFileMeta
	)
	if !atomic.CompareAndSwapInt32(&this.isend, 0, 1) {
		this.Runner.Debugf("Reader is collectioning rungoroutine:%d", atomic.LoadInt32(&this.rungoroutine))
		this.SyncMeta()
		return
	}
	cfiles = make(map[string]*compress.CopresFileMeta)
	if conn, err := this.ftp_connectFtp(); err != nil {
		this.Runner.Errorf("Reader Compress Dial Ftp err:%v", err)
		atomic.StoreInt32(&this.isend, 0)
		return
	} else {
		fmetas := this.meta.GetFiles()
		this.ftp_scandirectory(conn, this.options.GetCompress_directory(), fmetas, cfiles)
		conn.Quit()
	}
	this.Runner.Debugf("Reader Compress  collection filenum:%d", len(cfiles))
	this.Runner.Debugf("Reader Compress  collection cfiles:%v", cfiles)
	if len(cfiles) > 0 {
		procs := this.Runner.MaxProcs()
		if len(cfiles) < this.Runner.MaxProcs() {
			procs = len(cfiles)
		}
		this.colltask = make(chan *compress.CopresFileMeta, len(cfiles))
		for _, v := range cfiles {
			this.colltask <- v
		}

		for i := 0; i < procs; i++ {
			if conn, err := this.ftp_connectFtp(); err == nil {
				atomic.AddInt32(&this.rungoroutine, 1)
				this.wg.Add(1)
				go this.ftp_asyncollection(conn, this.colltask)
			} else {
				this.Runner.Errorf("Reader Compress Ftp collection err:%v", err)
			}
		}
		if atomic.LoadInt32(&this.rungoroutine) == 0 { // 没有任务
			atomic.StoreInt32(&this.isend, 0)
		}
	} else {
		atomic.StoreInt32(&this.isend, 0)
	}
}
func (this *Reader) ftp_asyncollection(conn *ftp.ServerConn, fmeta <-chan *compress.CopresFileMeta) {
	defer lego.Recover(fmt.Sprintf("%s Reader Compress", this.Runner.Name()))
	defer atomic.AddInt32(&this.rungoroutine, -1)
	defer this.wg.Done()
	defer conn.Quit()
locp:
	for v := range fmeta {
		//读取压缩包
		resp, err := this.ftpRetrbypice(conn, v.CompresName, 1024)
		if this.Runner.GetRunnerState() == core.Runner_Stoping {
			break locp
		}
		if err != nil && err != io.EOF {
			this.Runner.Debugf("Reader Compress ftp_asyncollection ftpRetrbypice ,err:%v", err)
			continue
		}
		//解压处理
		tr, err := this.unpackGzip(resp)
		if this.Runner.GetRunnerState() == core.Runner_Stoping {
			break locp
		}
		if err != nil && err != io.EOF {
			this.Runner.Debugf("Reader Compress ftp_asyncollection gz_Unpack ,err:%v", err)
			continue
		}
	clocp:
		for {
			//处理单个文件
			_, err := this.ftp_collection_file(conn, tr, v)
			if err != nil || err == io.EOF {
				this.Runner.Debugf("Reader Compress ftp_asyncollection ,err:%v", err)
				break clocp
			}

			if this.Runner.GetRunnerState() == core.Runner_Stoping { //采集器进入停止过程中
				break locp
			}
		}
		this.Runner.Debugf("Reader Compress ftp_asyncollection SyncMeta")
		if this.Runner.GetRunnerState() == core.Runner_Stoping || len(fmeta) == 0 {
			break locp
		} else {
			this.SyncMeta()
		}
	}
	this.SyncMeta()
	if atomic.CompareAndSwapInt32(&this.isend, 1, 0) { //最后一个任务已经完成
		this.collectionend()
	}
	this.Runner.Debugf("Reader Compress ftp_asyncollection asyncollection exit")
}
func (this *Reader) ftp_connectFtp() (conn *ftp.ServerConn, err error) {
	if conn, err = ftp.Dial(fmt.Sprintf("%s:%d", this.options.GetCompress_server_addr(), this.options.GetCompress_server_port()), ftp.DialWithTimeout(5*time.Second)); err == nil {
		err = conn.Login(this.options.GetCompress_server_user(), this.options.GetCompress_server_password())
	}
	return
}
func (this *Reader) ftp_scandirectory(conn *ftp.ServerConn, dir string, fmetas map[string]*compress.CopresFileMeta, cfile map[string]*compress.CopresFileMeta) (err error) {
	var (
		entries          []*ftp.Entry
		entrie           *ftp.Entry
		matched_compress bool
	)
	//扫描路径下文件
	if entries, err = conn.List(dir); err == nil {
		for _, entrie = range entries {

			if entrie.Size > uint64(this.options.GetCompress_max_collection_size()) {
				continue
			}
			var cpsf *compress.CopresFileMeta
			filepath := dir + "/" + entrie.Name
			if entrie.Type == ftp.EntryTypeFile {
				//匹配指定正则规则的压缩文件
				matched_compress, err = regexp.MatchString(this.options.GetCompress_regularrules(), entrie.Name)
				if err == nil && matched_compress {
					//判断meta文件是否存在
					if _, ok := fmetas[filepath]; !ok {
						cpsf = &compress.CopresFileMeta{
							CompresName:           filepath,
							CompresSize:           entrie.Size,
							CompresAlreadyCollect: 0,
						}
						this.meta.SetFile(cpsf.CompresName, cpsf)
						cfile[filepath] = cpsf
					}
				}

			} else if entrie.Type == ftp.EntryTypeFolder {
				err = this.ftp_scandirectory(conn, filepath, fmetas, cfile)
			} else {
				cpsf, _ = fmetas[filepath]
				//判断压缩包是否采集完
				if cpsf.CompresAlreadyCollect < cpsf.CompresSize {
					cfile[filepath] = cpsf
				}
			}
		}
	}
	return
}
func (this *Reader) ftp_collection_file(conn *ftp.ServerConn, tr *tar.Reader, cmf *compress.CopresFileMeta) (isskip bool, err error) {

	hd, err1 := tr.Next()
	if err1 == io.EOF {
		//压缩包采集完
		fmt.Println(err1)
		cmf.CompresAlreadyCollect = cmf.CompresSize
		this.meta.SetFile(cmf.CompresName, cmf)
		err = err1
		return
	}
	if err1 != nil {
		err = err1
		return
	}
	//过滤掉文件夹+正则过滤文件名
	matched, err := regexp.MatchString(this.options.GetCompress_inter_regularrules(), hd.Name)
	if hd.Size == 0 || (!matched && err == nil) {
		return true, nil
	}
	//文件是否已采集过
	_, ok := this.meta.GetFile(cmf.CompresName + "/" + hd.Name)
	if ok {
		return true, nil
	}
	//判断文件内容格式，读取文件
	var data interface{}
	xmlmatched, err := regexp.MatchString(this.options.GetCompress_xml_regularrules(), hd.Name)
	if xmlmatched && err == nil {
		//xml 格式内容
		data, err = this.xml_Reader(tr)
		if err != nil {
			err = fmt.Errorf("Compress xml_Reader read err:%s", err)
		}
	} else {
		//文本格式内容
		data, err = this.text_Reader(tr)
		if err != nil {
			err = fmt.Errorf("Compress text_Reader read err:%s", err)
		}
	}

	//采集发送
	this.Reader.Input() <- core.NewCollData("", data)

	//记录meta
	filemeta := &compress.CopresFileMeta{
		FileName:              cmf.CompresName + "/" + hd.Name,
		FileSize:              uint64(hd.Size),
		FileLastModifyTime:    hd.ModTime.Unix(), //只有采集完毕才同步时间
		FileAlreadyReadOffset: uint64(hd.Size),
		FileCacheData:         make([]byte, 0),
	}
	this.meta.SetFile(filemeta.FileName, filemeta)
	return true, nil
}
func (this *Reader) xml_Reader(tr *tar.Reader) (data interface{}, err error) {
	content := bufio.NewReader(tr)
	doc := etree.NewDocument()
	var n int64
	if n, err = doc.ReadFrom(content); err != nil {
		return
	}
	if n <= 0 {
		err = fmt.Errorf("Compress xml_parse doc ReadFrom 0")
		return
	}
	root := doc.FindElements("//" + this.options.GetCompress_xml_root_node() + "/*")
	jA := jsongo.Array()

	for _, v := range root {
		jO := jsongo.Object()
		//fmt.Printf("node:%v,value:%v\n", v.Tag, v.Text())
		jO.Put(v.Tag, v.Text())
		//fmt.Printf("Element:%v", v.SelectElement("string").Text())
		jA.Put(jO)
	}
	//fmt.Printf("xml_Reader+++:%s\n", jA.String())
	data = jA
	return
}
func (this *Reader) text_Reader(tr *tar.Reader) (data interface{}, err error) {
	buf := bytes.NewBuffer([]byte{})
	content := bufio.NewReader(tr)
	cb := make([]byte, 1024*1024*4)
	var n int
	for {
		n, err = content.Read(cb)
		buf.Write(cb[:n])
		if err != nil && err == io.EOF {
			break
		}
		if n == 0 {
			break
		}
	}
	//fmt.Printf("text_Reader+++:%s\n", buf.String())
	data = buf.String()
	return
}
func (this *Reader) ftpRetrbypice(c *ftp.ServerConn, path string, limitsize uint64) (databuf bytes.Buffer, err error) {
	size, err := c.FileSize(path)
	if err != nil {
		return
	}
	resp, err := c.RetrFrom(path, 0)
	defer resp.Close()
	if err != nil {
		return
	}
	var num, total int
	for {
		tmpbuff := make([]byte, limitsize)
		num, err = resp.Read(tmpbuff)
		if num > 0 {
			total += num
			databuf.Write(tmpbuff[0:num])
		}
		if err == io.EOF {
			break
		}
		if this.Runner.GetRunnerState() == core.Runner_Stoping { //采集器进入停止过程中
			break
		}
	}
	fmt.Printf("Reader Compress ftpRetrbypice path:%s,size:%d,total:%d,len:%d", path, size, total, databuf.Len())
	return
}
