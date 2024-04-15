package core_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"lego_datacollector/modules/datacollector/core"
	"testing"
)

func Test_JslionCollData(t *testing.T) {
	bucket := core.NewCollDataBucket(2, 1024)
	data := core.NewCollData("testdata", "liwei1dao")
	data.GetData()["hahah"] = "liwei1dao"
	data.GetData()["message"] = "<liwei2dao>"
	if bytes, err := data.ToString(); err == nil {
		fmt.Printf("data:%s\n", string(bytes))
		bucket.AddCollData(data)
		if bytes, err := bucket.ToString(true); err == nil {
			fmt.Printf("bucket:%s\n", string(bytes))
		}
	}

}

func Test_Json(t *testing.T) {
	buff := bytes.NewBuffer([]byte{})
	jsonEncoder2 := json.NewEncoder(buff)
	jsonEncoder2.SetEscapeHTML(false)
	jsonEncoder2.Encode("<liwei1dao>")
	fmt.Printf("bucket:%s\n", buff.String())
}
