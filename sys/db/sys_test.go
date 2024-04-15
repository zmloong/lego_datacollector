package db_test

import (
	"fmt"
	"lego_datacollector/comm"
	"lego_datacollector/modules/datacollector/metaer/file"
	"lego_datacollector/sys/db"
	"sync"
	"testing"
)

func Test_WriteSqlMeta(t *testing.T) {
	err := db.OnInit(map[string]interface{}{
		"RedisUrl":        "127.0.0.1:10001",
		"RedisDB":         1,
		"RedisPassword":   "li13451234",
		"MongodbUrl":      "mongodb://127.0.0.1:10002",
		"MongodbDatabase": "datacollector",
	})
	if err != nil {
		fmt.Printf("初始化系统失败:%v", err)
	}
	db.WriteMetaData("runner_test", "reder",
		map[string]*file.FileMeta{"liwei1dao.text": {
			FileName:              "liwei1dao.text",
			FileSize:              123,
			FileLastModifyTime:    12312312374,
			FileAlreadyReadOffset: 78,
			FileCacheData:         []byte{1, 2, 3},
		}},
	)
}

func Test_DelRunner(t *testing.T) {
	err := db.OnInit(map[string]interface{}{
		"RedisUrl":        "172.20.27.145:10001",
		"RedisDB":         1,
		"RedisPassword":   "li13451234",
		"MongodbUrl":      "mongodb://172.20.27.145:10002",
		"MongodbDatabase": "datacollector",
	})
	if err != nil {
		fmt.Printf("初始化系统失败:%v", err)
	}
	// rediskey := fmt.Sprintf(string(Cache_RunnerMeta), rId, "reader")
	db.DelRunner("oracle_001")
}

func Test_ReadSqlMeta(t *testing.T) {
	err := db.OnInit(map[string]interface{}{
		"RedisUrl":      "172.20.27.145:10001",
		"RedisDB":       0,
		"RedisPassword": "li13451234",
	})
	if err != nil {
		fmt.Printf("初始化系统失败:%v", err)
	}
	meta := map[string]*file.FileMeta{}
	if err = db.ReadMetaData("runner_test", "reder", meta); err != nil {
		fmt.Printf("ReadMetaData Error:%v", err)
	} else {
		fmt.Printf("ReadMetaData meta:%+v", meta)
	}
}

func Test_QueryAndRunNoRunRunnerConfig(t *testing.T) {
	err := db.OnInit(map[string]interface{}{
		"RedisUrl":        "172.20.27.145:10001",
		"RedisDB":         1,
		"RedisPassword":   "li13451234",
		"MongodbUrl":      "mongodb://172.20.27.145:10002",
		"MongodbDatabase": "datacollector",
	})
	if err != nil {
		fmt.Printf("初始化系统失败:%v", err)
	}
	rconf := &comm.RunnerConfig{}
	if rconf, err = db.QueryAndRunNoRunRunnerConfig("172.20.27.145"); err != nil {
		fmt.Printf("QueryAndRunNoRunRunnerConfig Error:%v", err)
	} else {
		fmt.Printf("QueryAndRunNoRunRunnerConfig rconf:%+v", rconf)
	}
}

func Test_QueryRunnerRuninfo(t *testing.T) {
	err := db.OnInit(map[string]interface{}{
		"RedisUrl":        "172.20.27.145:10001",
		"RedisDB":         1,
		"RedisPassword":   "li13451234",
		"MongodbUrl":      "mongodb://172.20.27.145:10002",
		"MongodbDatabase": "datacollector",
	})
	if err != nil {
		fmt.Printf("初始化系统失败:%v", err)
	}
	if data, err := db.QueryRunnerRuninfo("test_liwei1dao"); err != nil {
		fmt.Printf("执行 QueryRunnerRuninfo err:%v", err)
	} else {
		fmt.Printf("执行 QueryRunnerRuninfo succ:%v", data)
	}
}
func Test_UpdateRunnerRuninfo(t *testing.T) {
	err := db.OnInit(map[string]interface{}{
		"RedisUrl":      "172.20.27.145:10001",
		"RedisDB":       0,
		"RedisPassword": "li13451234",
	})
	if err != nil {
		fmt.Printf("初始化系统失败:%v", err)
	}
	wg := new(sync.WaitGroup)
	wg.Add(10)
	for i := 0; i < 10; i++ {
		go func() {
			if err = db.WriteRunnerRuninfo(&comm.RunnerRuninfo{
				RName:      "test_liwei1dao",
				RunService: map[string]struct{}{"service1": {}},
				RecordTime: 197263,
				ReaderCnt:  100,
				ParserCnt:  1,
				SenderCnt:  map[string]int64{"kafka": 1},
			}); err != nil {
				fmt.Printf("执行 UpdateRunnerRuninfo err:%v", err)
			} else {
				fmt.Printf("执行 UpdateRunnerRuninfo succ")
			}
			wg.Done()
		}()
	}
	wg.Wait()
}

func Test_INCRBY(t *testing.T) {
	err := db.OnInit(map[string]interface{}{
		"RedisUrl":      "172.20.27.145:10001",
		"RedisDB":       1,
		"RedisPassword": "li13451234",
	})
	if err != nil {
		fmt.Printf("初始化系统失败:%v", err)
	}
	wg := new(sync.WaitGroup)
	wg.Add(20)
	for i := 0; i < 20; i++ {
		go func() {
			if err = db.RealTimeStatisticsIn("liwei1dao", 1); err != nil {
				fmt.Printf("执行错误:%v", err)
			}
			wg.Done()
		}()
	}
	wg.Wait()
}

func Test_UpDateServiceNodeInfo(t *testing.T) {
	err := db.OnInit(map[string]interface{}{
		"RedisUrl":      "172.20.27.145:10001",
		"RedisDB":       1,
		"RedisPassword": "li13451234",
	})
	if err != nil {
		fmt.Printf("初始化系统失败:%v", err)
	}
	db.UpDateServiceNodeInfo(&comm.ServiceNode{
		Id:        "data_1",
		IP:        "127.9.0.1",
		Port:      123,
		Wight:     0.134,
		RunnerNum: 1,
		State:     1,
	})
}
