package file

import (
	"fmt"
	"testing"
)

func Test_Options(t *testing.T) {
	if opt, err := newOptions(map[string]interface{}{
		"type":           "file",
		"file_send_path": "liwei1dao",
		"file_partition": 10,
	}); err == nil {
		fmt.Printf("opt:%+v", opt)
	} else {
		fmt.Printf("err:%v", err)
	}
}
