package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
)

func Test_RunnerAutoCloseNotice(t *testing.T) {
	client := &http.Client{}
	data := make(map[string]interface{})
	data["clientId"] = "datac_11"
	data["code"] = "runner_syslog"
	data["flowNum"] = "runner_syslog_1234"
	data["status"] = 4
	bytesData, _ := json.Marshal(data)
	if req, err := http.NewRequest("POST", "http://172.20.27.143:8713/dataSynTask/changeStatus", bytes.NewReader(bytesData)); err == nil {
		if resp, err := client.Do(req); err == nil {
			if body, err := ioutil.ReadAll(resp.Body); err == nil {
				fmt.Printf("RunnerAutoCloseNotice result:%s", body)
			} else {
				fmt.Printf("RunnerAutoCloseNotice err:%v", err)
			}
		} else {
			fmt.Printf("RunnerAutoCloseNotice err:%v", err)
		}
	} else {
		fmt.Printf("RunnerAutoCloseNotice err:%v", err)
	}
}

func Test_Post(t *testing.T) {
	if resp, err := http.Post("http://172.20.27.143:8713/dataSynTask/changeStatus",
		"application/x-www-form-urlencoded",
		strings.NewReader(fmt.Sprintf("clientId=%s&code=%s&flowNum=%s&status=%d", "datac_11", "SNP202204181047588301515884817093365760", "runner_syslog_1234", 4))); err == nil {
		body, err := ioutil.ReadAll(resp.Body)
		if err == nil {
			defer resp.Body.Close()
			fmt.Printf("RunnerAutoCloseNotice result:%s", string(body))
		} else {
			fmt.Printf("RunnerAutoCloseNotice err:%v", err)
		}
	} else {
		fmt.Printf("RunnerAutoCloseNotice err:%v", err)
	}
}

func Test_Post_(t *testing.T) {
	if resp, err := http.Post(fmt.Sprintf("http://172.20.27.143:8713/dataSynTask/changeStatus?clientId=%s&code=%s&flowNum=%s&status=%d", "datac_11", "SNP202204181047588301515884817093365760", "runner_syslog_1234", 4),
		"application/json",
		nil); err == nil {
		body, err := ioutil.ReadAll(resp.Body)
		if err == nil {
			defer resp.Body.Close()
			fmt.Printf("RunnerAutoCloseNotice result:%s", string(body))
		} else {
			fmt.Printf("RunnerAutoCloseNotice err:%v", err)
		}
	} else {
		fmt.Printf("RunnerAutoCloseNotice err:%v", err)
	}
}
