package webservice

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"lego_datacollector/modules/datacollector/core"
	"lego_datacollector/modules/datacollector/reader"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/beevik/etree"
	lgcron "github.com/liwei1dao/lego/sys/cron"
	"github.com/ricardolonga/jsongo"
	"github.com/robfig/cron/v3"
)

type Reader struct {
	reader.Reader
	options      IOptions
	httpClient   *http.Client
	runId        cron.EntryID
	rungoroutine int32
	wg           sync.WaitGroup //用于等待采集器完全关闭
}

func (this *Reader) Init(runner core.IRunner, reader core.IReader, meta core.IMetaerData, options core.IReaderOptions) (err error) {
	if err = this.Reader.Init(runner, reader, meta, options); err != nil {
		return
	}
	this.options = options.(IOptions)
	var t = &http.Transport{
		Dial: (&net.Dialer{
			Timeout:   this.options.GetDialTimeout(),
			KeepAlive: 30 * time.Second,
		}).Dial,
		ResponseHeaderTimeout: this.options.GetRespTimeout(),
		TLSClientConfig:       &tls.Config{},
	}
	this.httpClient = &http.Client{Transport: t}

	return
}

func (this *Reader) Start() (err error) {
	err = this.Reader.Start()
	if this.runId, err = lgcron.AddFunc(this.options.GetWebservice_interval(), this.Fetch); err != nil {
		err = fmt.Errorf("lgcron.AddFunc corn:%s err:%v", this.options.GetWebservice_interval(), err)
		return
	}
	if this.options.GetWebservice_exec_onstart() {
		go this.Fetch()
	}
	return nil
}

///外部调度器 驱动执行  此接口 不可阻塞
func (this *Reader) Drive() (err error) {
	err = this.Reader.Drive()
	return
}
func (this *Reader) Close() (err error) {
	lgcron.Remove(this.runId)
	this.wg.Wait()
	err = this.Reader.Close()
	return nil
}
func (this *Reader) Fetch() {
	for i := 1; i <= 10; i++ {
		content, err := this.curl()
		if err == nil {
			this.Reader.Input() <- core.NewCollData("", content)
			return
		}
		if i == 10 {
			this.Runner.Debugf("Reader webservice Fetch 请求重试【10】次，最后一次失败，err：%v", err)
		}
	}
	return
}
func (this *Reader) curl() (data interface{}, err error) {
	data, err = this.webervice_post()
	if err != nil {
		return
	}
	return
}
func (this *Reader) webervice_post() (data interface{}, err error) {
	//用户体检报告请求体
	payload := &bytes.Buffer{}
	//vid:="6917000067"
	payload.WriteString(this.options.GetWebservice_reqbody())
	client := &http.Client{
		Timeout: this.options.GetRespTimeout(),
	}
	req, err := http.NewRequest("POST", this.options.GetWebservice_address(), payload)
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", "text/xml; charset=utf-8")
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}
	err, data = this.xml_parse(content)
	if err != nil {
		return
	}
	return

}

func (this *Reader) xml_parse(content []byte) (err error, data interface{}) {
	doc := etree.NewDocument()
	if err = doc.ReadFromBytes(content); err != nil {
		return
	}
	root := doc.FindElements("//" + this.options.GetWebservice_xml_body_node() + "/*")
	jA := jsongo.Array()

	for _, v := range root {
		jO := jsongo.Object()
		//fmt.Printf("node:%v,value:%v\n", v.Tag, v.Text())
		jO.Put(v.Tag, v.Text())
		//fmt.Printf("Element:%v", v.SelectElement("string").Text())
		jA.Put(jO)
	}
	data = jA
	return
}
