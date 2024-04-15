package mysql_test

import (
	"database/sql"
	"fmt"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

func Test_Mysql_sql(t *testing.T) {
	db, err := sql.Open("mysql", "root:Idss@sjzt2021@tcp(172.20.27.125:3306)/mysql")
	if err != nil {
		fmt.Printf("初始化失败=%v\n", err)
	} else {
		// sqlstr := fmt.Sprintf("INSERT INTO test_liwei(name,age,boy,time) VALUES ('liwei3dao',22,true,'%s')", time.Now().Format("2006-01-02 15:04:05"))
		// if _, err := db.Exec(sqlstr); err == nil {
		// 	fmt.Printf("gettablecount %s succ:\n", "test1")
		// } else {
		// 	fmt.Printf("gettablecount %s err:%v\n", "test1", err)
		// }

		table := "test_liwei"
		keys := []string{}
		mapping := map[string]string{"name": "name", "age": "age", "boy": "boy", "time": "time"}
		originaldata := []map[string]interface{}{
			{"message": map[string]interface{}{"name": "liwei1dao", "age": 23, "boy": true, "time": time.Now()}},
			{"message": map[string]interface{}{"name": "liwei2dao", "age": 23, "boy": true, "time": time.Now()}},
			{"message": map[string]interface{}{"name": "liwei3dao", "age": 23, "boy": true, "time": time.Now()}},
			{"message": map[string]interface{}{"name": "liwei4dao", "age": 23, "boy": true, "time": time.Now()}},
			{"message": map[string]interface{}{"name": "liwei5dao", "age": 23, "boy": true, "time": time.Now()}},
			{"message": map[string]interface{}{"name": "liwei6dao", "age": 23, "boy": true, "time": time.Now()}},
			{"message": map[string]interface{}{"name": "liwei7dao", "age": 23, "boy": true, "time": time.Now()}},
			{"message": map[string]interface{}{"name": "liwei8dao", "age": 24, "boy": true, "time": time.Now()}},
			{"message": map[string]interface{}{"name": "liwei9dao", "age": 23, "boy": true, "time": time.Now()}},
			{"message": map[string]interface{}{"name": "liwei10dao", "age": 23, "boy": true, "time": time.Now()}},
			{"message": map[string]interface{}{"name": "liwei11dao", "age": 23, "boy": true, "time": time.Now()}},
			{"message": map[string]interface{}{"name": "liwei12dao", "age": 23, "boy": false, "time": time.Now()}},
			{"message": map[string]interface{}{"name": "liwei13dao", "age": 23, "boy": true, "time": time.Now()}},
			{"message": map[string]interface{}{"name": "liwei14dao", "age": 23, "boy": true, "time": time.Now()}},
			{"message": map[string]interface{}{"name": "liwei15dao", "age": 23, "boy": true, "time": time.Now()}},
			{"message": map[string]interface{}{"name": "liwei16dao", "age": 23, "boy": false, "time": time.Now()}},
		}
		getemptycontainer := func() (result map[string]interface{}) {
			result = make(map[string]interface{})
			for _, v := range mapping {
				result[v] = nil
			}
			return
		}
		transIntface := func(data map[string]interface{}) (result string) {
			result = "("
			for _, v := range keys {
				switch data[v].(type) {
				case string:
					result += fmt.Sprintf("'%s',", data[v])
					break
				case time.Time:
					result += fmt.Sprintf("'%s',", data[v].(time.Time).Format("2006-01-02 15:04:05"))
					break
				default:
					result += fmt.Sprintf("%v,", data[v])
					break
				}
			}
			result = result[0 : len(result)-1]
			result += ")"
			return
		}

		datastr := ""
		sqlstr := "INSERT INTO " + table + "("

		for k, _ := range getemptycontainer() {
			sqlstr += k + ","
			keys = append(keys, k)
		}
		sqlstr = sqlstr[0 : len(sqlstr)-1]
		sqlstr += ")VALUES"
		//在写入数据
		for _, d := range originaldata {
			data := getemptycontainer()
			for k, v := range d {
				if v1, ok := mapping[k]; ok { //在映射表中提取
					data[v1] = v
				}
			}
			if m, ok := d["message"].(map[string]interface{}); ok {
				for k, v := range m {
					if v1, ok := mapping[k]; ok { //在映射表中提取
						data[v1] = v
					}
				}
			}
			datastr += transIntface(data) + ","
		}
		datastr = datastr[0 : len(datastr)-1]
		fmt.Printf("sql:%s", sqlstr+datastr)
		if _, err := db.Exec(sqlstr + datastr); err != nil {
			fmt.Printf("Send Mysql err:%v", err)
			return
		}
	}
}
