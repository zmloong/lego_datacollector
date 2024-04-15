package hbase_test

import (
	"context"
	"fmt"
	"io"
	"log"
	"testing"

	"github.com/tsuna/gohbase"
	"github.com/tsuna/gohbase/filter"
	"github.com/tsuna/gohbase/hrpc"
)

func Test_hbase(t *testing.T) {

}
func getHbaseConnect(zk_host string) (client gohbase.Client) {
	client = gohbase.NewClient(zk_host)
	return
}
func getHbaseAdminConnect(zk_host string) (adminClient gohbase.AdminClient) {
	adminClient = gohbase.NewAdminClient(zk_host)
	return
}
func getROW(host string) {
	client := getHbaseConnect(host)
	getRequest, err := hrpc.NewGetStr(context.Background(), "student", "1001")
	if err != nil {
		fmt.Println(err)
	}
	getRes, err := client.Get(getRequest)
	fmt.Printf("getRes:%v", getRes)
}
func getList(host string) {
	adminClient := getHbaseAdminConnect(host)
	getRequest1, err := hrpc.NewListTableNames(context.Background())
	if err != nil {
		fmt.Println(err)
	}
	getRes1, err := adminClient.ListTableNames(getRequest1)
	var namelist []string
	for _, v := range getRes1 {
		namelist = append(namelist, string(v.Qualifier))
	}
	fmt.Println(namelist)
}
func scan01(host, tableName, startRow, stopRow string) {
	client := getHbaseConnect(host)
	getRequest, err := hrpc.NewScanRangeStr(context.Background(), tableName, startRow, stopRow)
	if err != nil {
		fmt.Println(err)
	}
	scan := client.Scan(getRequest)
	var res []*hrpc.Result
	for {
		getRsp, err := scan.Next()
		if err == io.EOF || getRsp == nil {
			break
		}
		if err != nil {
			log.Print(err, "scan.Next")
		}
		res = append(res, getRsp)
	}
	fmt.Println(res)
}
func scan02(host, tableName, startRow, stopRow string) {
	client := getHbaseConnect(host)
	getRequest, err := hrpc.NewScanRangeStr(context.Background(), tableName, startRow, stopRow)
	if err != nil {
		fmt.Println(err)
	}
	scan := client.Scan(getRequest)
	var res []*hrpc.Result
	for {
		getRsp, err := scan.Next()
		if err == io.EOF || getRsp == nil {
			break
		}
		if err != nil {
			log.Print(err, "scan.Next")
		}
		res = append(res, getRsp)
	}
	fmt.Println(res)
}
func filterscan(host string) {
	client := getHbaseConnect(host)
	pFilter := filter.NewPrefixFilter([]byte(""))
	scanRequest, err := hrpc.NewScanStr(context.Background(), "test", hrpc.Filters(pFilter))
	if err != nil {
		fmt.Println(err)
	}
	scanRsp := client.Scan(scanRequest)
	var res []*hrpc.Result
	for {
		getRsp, err := scanRsp.Next()
		if err == io.EOF || getRsp == nil {
			break
		}
		if err != nil {
			log.Print(err, "scan.Next")
		}
		res = append(res, getRsp)
		fmt.Println(getRsp)
	}
	fmt.Printf("len:%v,%v", len(res), res)
}
func putstring(host string) {
	client := getHbaseConnect(host)
	rowKey := "1004" // RowKey
	value := map[string]map[string][]byte{"info": {"name": []byte("zml"), "age": []byte("33"), "sex": []byte("male")}}
	putRequest, err := hrpc.NewPutStr(context.Background(), "test", rowKey, value)
	if err != nil {
		fmt.Println(err)
	}
	res, err := client.Put(putRequest)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(res)
}
