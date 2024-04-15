package httpfetch

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"lego_datacollector/comm"
	"lego_datacollector/modules/datacollector/core"
	"lego_datacollector/modules/datacollector/reader"
	"mime/multipart"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	lgcron "github.com/liwei1dao/lego/sys/cron"
	"github.com/liwei1dao/lego/sys/event"
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

	if !this.options.GetAutoClose() {
		if this.runId, err = lgcron.AddFunc(this.options.GetHttpfetch_interval(), this.Fetch); err != nil {
			err = fmt.Errorf("lgcron.AddFunc corn:%s err:%v", this.options.GetHttpfetch_interval(), err)
			return
		}
		if this.options.GetHttpfetch_exec_onstart() {
			go this.Fetch()
		}
	} else {
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
	if !this.options.GetAutoClose() {
		lgcron.Remove(this.runId)
	}
	this.wg.Wait()
	err = this.Reader.Close()
	return
}
func (this *Reader) AutoClose(msg string) {
	time.Sleep(time.Second * 10)
	this.Runner.Close(core.Runner_Runing, msg) //结束任务
}
func (this *Reader) Fetch() {
	for i := 1; i <= 10; i++ {
		content, err := this.curl()
		if err == nil {
			if len(content) > int(this.Runner.MaxCollDataSzie()) {
				this.Runner.Debugf("Reader http Fetch data size 大于消息大小限制，直接丢弃")
				event.TriggerEvent(comm.Event_WriteLog, this.Runner.Name(), this.Runner.InstanceId(), "http Fetch data大于消息大小限制，直接丢弃，关闭任务!", 1, time.Now().Unix())
				if this.options.GetAutoClose() { //自动关闭
					go this.AutoClose(core.RunnerFailAutoClose)
				}
				return
			} else {
				this.Reader.Input() <- core.NewCollData("", content)
				if this.options.GetAutoClose() { //自动关闭
					go this.AutoClose(core.RunnerSuccAutoClose)
				}
				return
			}
		}
		if i == 10 {
			this.Runner.Debugf("Reader http Fetch 请求重试【10】次，最后一次失败，err：%v", err)
			event.TriggerEvent(comm.Event_WriteLog, this.Runner.Name(), this.Runner.InstanceId(), "http Fetch 请求数据重试10次失败 关闭任务!", 1, time.Now().Unix())
			if this.options.GetAutoClose() { //自动关闭
				go this.AutoClose(core.RunnerFailAutoClose)
			}
		}
	}

	return
}
func (this *Reader) curl() (string, error) {
	isform_data := strings.Contains(this.options.GetHttpfetch_headersMap()["Content-Type"], "form-data")
	if isform_data {
		content, err := this.curl_formdata()
		if err != nil {
			return "", err
		}
		return string(content), nil
	} else {
		content, err := this.curl_get_post()
		if err != nil {
			return "", err
		}
		return string(content), nil

	}
}
func (this *Reader) curl_formdata() (string, error) {
	var dat map[string]interface{}
	if err := json.Unmarshal([]byte(this.options.GetHttpfetch_body()), &dat); err != nil {
		return "", fmt.Errorf("Reader httpfetch curl_formdata json.Unmarshal:%v", err)
	}
	payload := &bytes.Buffer{}
	writer := multipart.NewWriter(payload)
	for k, v := range dat {
		_ = writer.WriteField(k, fmt.Sprintf("%v", v))
	}
	err := writer.Close()
	if err != nil {
		fmt.Println(err)
		return "", fmt.Errorf("Reader httpfetch curl_formdata multipart.NewWriter:%v", err)
	}
	client := &http.Client{}
	req, err := http.NewRequest("POST", this.options.GetHttpfetch_address(), payload)
	if err != nil {
		fmt.Println(err)
		return "", fmt.Errorf("Reader httpfetch curl_formdata NewRequest:%v", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("Reader httpfetch curl_formdata client.Do:%v", err)
	}
	defer resp.Body.Close()

	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if resp.StatusCode != 200 {
		return "", errors.New(string(content))
	}
	return string(content), nil

}
func (this *Reader) curl_get_post() (string, error) {
	var body io.Reader
	if this.options.GetHttpfetch_body() != "" {
		body = bytes.NewBufferString(this.options.GetHttpfetch_body())
	}
	req, err := http.NewRequest(this.options.GetHttpfetch_method(), this.options.GetHttpfetch_address(), body)
	if err != nil {
		return "", fmt.Errorf("Reader httpfetch curl_get_post NewRequest err:%v", err)
	}
	if this.options.GetHttpfetch_headersMap() != nil {
		for k, v := range this.options.GetHttpfetch_headersMap() {
			req.Header.Add(k, v)
		}
	}

	resp, err := this.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("Reader httpfetch curl_get_post httpClient.Do err:%v", err)
	}
	defer resp.Body.Close()
	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if resp.StatusCode != 200 {
		return "", errors.New(string(content))
	}
	return string(content), nil
}
