package redis_test

import (
	"fmt"
	"testing"
	"time"
)

func Test_redis(t *testing.T) {
	colltask := make(chan int, 10) //文件采集任务
	for i := 0; i < 9; i++ {
		colltask <- i
	}
	for i := 0; i < 10; i++ {
		go func() {
			for v := range colltask {
				time.Sleep(time.Second * 10)
				fmt.Printf("结束:%d\n", v)
			}
			fmt.Printf("退出 \n")
		}()
	}
	fmt.Printf("启动 01\n")
	time.Sleep(time.Second * 1)
	fmt.Printf("启动 02\n")
	close(colltask)
	fmt.Printf("启动 03\n")
	time.Sleep(time.Second * 20)
}
