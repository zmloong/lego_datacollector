package folder_test

import (
	"fmt"
	"io/ioutil"
	"regexp"
	"testing"
)

func Test_regexp(t *testing.T) {
	matched, err := regexp.MatchString("\\.(log|txt)$", "liwei1dao.vip")
	fmt.Printf("matched:%v err:%v", matched, err)
}

func Test_readDir(t *testing.T) {
	fs, err := ioutil.ReadDir("/Users/liwei1dao/work/temap/logkit_test/reder/")
	fmt.Printf("fs:%+v err:%v", fs, err)
}
