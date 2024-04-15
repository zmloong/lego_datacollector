package doris_test

import (
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"lego_datacollector/modules/datacollector/sender/doris"
	"net/http"
	"strings"
	"testing"

	_ "github.com/go-sql-driver/mysql"
	"github.com/liwei1dao/lego/utils/container/id"
)

func Test_Doris_sql(t *testing.T) {
	db, err := sql.Open("mysql", "root@tcp(172.20.27.148:9030)/doris")
	if err != nil {
		fmt.Printf("初始化失败=%v\n", err)
	} else {
		sqlstr := "CREATE TABLE IF NOT EXISTS doris.collect \n" +
			"(\n" +
			"	`idss_collect_time` DATETIME COMMENT \"idss_collect_time\"," +
			"	`idss_collect_ip` VARCHAR(20) COMMENT \"idss_collect_ip\"," +
			"	`eqpt_ip` VARCHAR(20) COMMENT \"eqpt_ip\"," +
			"	`idss_collect_expmatch` BOOLEAN COMMENT \"idss_collect_expmatch\"," +
			"	`message` VARCHAR(1000) COMMENT \"message\"" +
			")\n" +
			"ENGINE=olap\n" +
			"DUPLICATE KEY(`idss_collect_time`)\n" +
			"PARTITION BY RANGE(`idss_collect_time`) ()\n" +
			"DISTRIBUTED BY HASH(`idss_collect_time`) BUCKETS 16\n" +
			"PROPERTIES\n" +
			"(\n" +
			"    \"replication_num\" = \"3\",\n" +
			"    \"storage_medium\" = \"SSD\",\n" +
			"    \"dynamic_partition.enable\" = \"true\",\n" +
			"    \"dynamic_partition.time_unit\" = \"DAY\",\n" +
			"    \"dynamic_partition.end\" = \"3\",\n" +
			"    \"dynamic_partition.prefix\" = \"p\",\n" +
			"    \"dynamic_partition.buckets\" = \"16\"\n" +
			");"
		if result, err := db.Exec(sqlstr); err == nil {
			fmt.Printf("Test_DM_sql result:%v\n", result)
		} else {
			fmt.Printf("Test_DM_sql err:%v\n", err)
		}
	}
}

func Test_Doris_column(t *testing.T) {
	db, err := sql.Open("mysql", "root@tcp(172.20.27.148:9030)/doris")
	if err != nil {
		fmt.Printf("初始化失败=%v\n", err)
	} else {
		sqlstr := "select COLUMN_NAME from information_schema.COLUMNS where table_name = 'collect';"
		if result, err := db.Query(sqlstr); err == nil {
			for result.Next() {
				column := ""
				if err := result.Scan(&column); err != nil {
					fmt.Printf("gettablecount %s err:%v\n", "collect", err)
				} else {
					fmt.Printf("gettablecount %s column:%s\n", "collect", column)
				}
			}
		} else {
			fmt.Printf("Test_DM_sql err:%v\n", err)
		}
	}
}

func Test_Doris_http(t *testing.T) {
	var load StreamLoad
	load.userName = "root"
	load.password = ""
	// data := "2022-01-26 00:00:00,127.0.0.1,127.0.0.1,0,liwei5dao"
	// data := `2022-01-26 00:00:00,,,,<165>Sep 18 16:58:30 cngb-login-b13-3 Auditd type=SYSCALL msg=audit(1505725108.802:15396685): arch=c000003e syscall=59 success=yes exit=0 a0=1f11bb0 a1=1f0a800 a2=1f06ca0 a3=10 items=2 ppid=25936 pid=25937 auid=36374 uid=36374 gid=3004 euid=36374 suid=36374 fsuid=36374 egid=3004 sgid=3004 fsgid=3004 tty=(none) ses=69813 comm="qstat" idss_source_ip=10 exe="/opt/gridengine/bin/linux-x64/qstat" key="exec"2019-09-18 16:58:23.168`
	data := "2022-01-26 11:51:00||||{\"idss_collect_time\":\"2022-01-26 00:00:00\",\"idss_collect_ip\":\"3\",\"eqpt_ip\":\"6\",\"idss_collect_expmatch\":\"2\",\"message\":\"3\"}\n" +
		"2022-01-26 11:39:00||||<165>Sep 18 16:58:30 cngb-login-b13-3 Auditd type=SYSCALL msg=audit(1505725108.802:15396685): arch=c000003e syscall=59 success=yes exit=0 a0=1f11bb0 a1=1f0a800 a2=1f06ca0 a3=10 items=2 ppid=25936 pid=25937 auid=36374 uid=36374 gid=3004 euid=36374 suid=36374 fsuid=36374 egid=3004 sgid=3004 fsgid=3004 tty=(none) ses=69813 comm=\"qstat\" idss_source_ip=10 exe=\"/opt/gridengine/bin/linux-x64/qstat\" key=\"exec\"2019-09-18 16:58:23.168\n" +
		"2022-01-26 11:39:00||||<165>Sep 18 16:58:30 cngb-login-b13-3 Auditd type=SYSCALL msg=audit(1505725108.802:15396685): arch=c000003e syscall=59 success=yes exit=0 a0=1f11bb0 a1=1f0a800 a2=1f06ca0 a3=10 items=2 ppid=25936 pid=25937 auid=36374 uid=36374 gid=3004 euid=36374 suid=36374 fsuid=36374 egid=3004 sgid=3004 fsgid=3004 tty=(none) ses=69813 comm=\"qstat\" idss_source_ip=10 exe=\"/opt/gridengine/bin/linux-x64/qstat\" key=\"exec\"2019-09-18 16:58:23.168\n" +
		"2022-01-26 11:39:00||||<165>Sep 18 16:58:30 cngb-login-b13-3 Auditd type=SYSCALL msg=audit(1505725108.802:15396685): arch=c000003e syscall=59 success=yes exit=0 a0=1f11bb0 a1=1f0a800 a2=1f06ca0 a3=10 items=2 ppid=25936 pid=25937 auid=36374 uid=36374 gid=3004 euid=36374 suid=36374 fsuid=36374 egid=3004 sgid=3004 fsgid=3004 tty=(none) ses=69813 comm=\"qstat\" idss_source_ip=10 exe=\"/opt/gridengine/bin/linux-x64/qstat\" key=\"exec\"2019-09-18 16:58:23.168\n" +
		"2022-01-26 11:39:00||||<165>Sep 18 16:58:30 cngb-login-b13-3 Auditd type=SYSCALL msg=audit(1505725108.802:15396685): arch=c000003e syscall=59 success=yes exit=0 a0=1f11bb0 a1=1f0a800 a2=1f06ca0 a3=10 items=2 ppid=25936 pid=25937 auid=36374 uid=36374 gid=3004 euid=36374 suid=36374 fsuid=36374 egid=3004 sgid=3004 fsgid=3004 tty=(none) ses=69813 comm=\"qstat\" idss_source_ip=10 exe=\"/opt/gridengine/bin/linux-x64/qstat\" key=\"exec\"2019-09-18 16:58:23.168\n" +
		"2022-01-26 11:39:00||||<165>Sep 18 16:58:30 cngb-login-b13-3 Auditd type=SYSCALL msg=audit(1505725108.802:15396685): arch=c000003e syscall=59 success=yes exit=0 a0=1f11bb0 a1=1f0a800 a2=1f06ca0 a3=10 items=2 ppid=25936 pid=25937 auid=36374 uid=36374 gid=3004 euid=36374 suid=36374 fsuid=36374 egid=3004 sgid=3004 fsgid=3004 tty=(none) ses=69813 comm=\"qstat\" idss_source_ip=10 exe=\"/opt/gridengine/bin/linux-x64/qstat\" key=\"exec\"2019-09-18 16:58:23.168\n" +
		"2022-01-26 11:39:00||||<165>Sep 18 16:58:30 cngb-login-b13-3 Auditd type=SYSCALL msg=audit(1505725108.802:15396685): arch=c000003e syscall=59 success=yes exit=0 a0=1f11bb0 a1=1f0a800 a2=1f06ca0 a3=10 items=2 ppid=25936 pid=25937 auid=36374 uid=36374 gid=3004 euid=36374 suid=36374 fsuid=36374 egid=3004 sgid=3004 fsgid=3004 tty=(none) ses=69813 comm=\"qstat\" idss_source_ip=10 exe=\"/opt/gridengine/bin/linux-x64/qstat\" key=\"exec\"2019-09-18 16:58:23.168\n" +
		"2022-01-26 11:39:00||||<165>Sep 18 16:58:30 cngb-login-b13-3 Auditd type=SYSCALL msg=audit(1505725108.802:15396685): arch=c000003e syscall=59 success=yes exit=0 a0=1f11bb0 a1=1f0a800 a2=1f06ca0 a3=10 items=2 ppid=25936 pid=25937 auid=36374 uid=36374 gid=3004 euid=36374 suid=36374 fsuid=36374 egid=3004 sgid=3004 fsgid=3004 tty=(none) ses=69813 comm=\"qstat\" idss_source_ip=10 exe=\"/opt/gridengine/bin/linux-x64/qstat\" key=\"exec\"2019-09-18 16:58:23.168\n" +
		"2022-01-26 11:39:00||||<165>Sep 18 16:58:30 cngb-login-b13-3 Auditd type=SYSCALL msg=audit(1505725108.802:15396685): arch=c000003e syscall=59 success=yes exit=0 a0=1f11bb0 a1=1f0a800 a2=1f06ca0 a3=10 items=2 ppid=25936 pid=25937 auid=36374 uid=36374 gid=3004 euid=36374 suid=36374 fsuid=36374 egid=3004 sgid=3004 fsgid=3004 tty=(none) ses=69813 comm=\"qstat\" idss_source_ip=10 exe=\"/opt/gridengine/bin/linux-x64/qstat\" key=\"exec\"2019-09-18 16:58:23.168\n"
	batch_load_data(load, data)
}

type StreamLoad struct {
	url       string
	dbName    string
	tableName string
	data      string
	userName  string
	password  string
}

//实现Doris用户认证信息
func auth(load StreamLoad) string {
	s := load.userName + ":" + load.password
	b := []byte(s)

	sEnc := base64.StdEncoding.EncodeToString(b)
	fmt.Printf("enc=[%s]\n", sEnc)

	sDec, err := base64.StdEncoding.DecodeString(sEnc)
	if err != nil {
		fmt.Printf("base64 decode failure, error=[%v]\n", err)
	} else {
		fmt.Printf("dec=[%s]\n", sDec)
	}
	return sEnc
}

//内存流数据，通过Stream Load导入Doris表中
func batch_load_data(load StreamLoad, data string) {
	client := &http.Client{}
	//生成要访问的url
	url := "http://172.20.27.148:8040/api/doris/txt/_stream_load"

	record := strings.NewReader(data)
	//提交请求
	reqest, err := http.NewRequest(http.MethodPut, url, record)

	//增加header选项
	reqest.Header.Add("Authorization", "Basic "+auth(load))
	reqest.Header.Add("EXPECT", "100-continue")
	reqest.Header.Add("label", id.NewXId())
	reqest.Header.Add("column_separator", "|")

	if err != nil {
		panic(err)
	}
	//处理返回结果
	response, err := client.Do(reqest)
	if err == nil {
		defer response.Body.Close()
		body, _ := ioutil.ReadAll(response.Body)
		responseBody := doris.ResponseBody{}
		jsonStr := string(body)
		fmt.Printf("jsonStr : %s ", jsonStr)
		err := json.Unmarshal([]byte(jsonStr), &responseBody)
		if err != nil {
			fmt.Println(err.Error())
		}
		if responseBody.Status == "Success" {
			//如果有被过滤的数据，打印错误的URL
			if responseBody.NumberFilteredRows > 0 {
				fmt.Printf("Error Data : %s ", responseBody.ErrorURL)
			} else {
				fmt.Printf("Success import data : %d", responseBody.NumberLoadedRows)
			}
		} else {
			fmt.Printf("Error Message : %s \n", responseBody.Message)
			fmt.Printf("Error Data : %s ", responseBody.ErrorURL)
		}
		//fmt.Println(jsonStr)
	} else {
		fmt.Printf("Error : %v ", err)
	}
}
