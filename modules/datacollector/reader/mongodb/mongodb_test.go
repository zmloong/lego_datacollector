package mongodb_test

import (
	"database/sql"
	"fmt"
	"testing"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/liwei1dao/dm"
)

func Test_Mongodb(t *testing.T) {
	db, err := sql.Open("mysql", "root:Idss@sjzt2021@tcp(172.20.27.125:3306)/mysql")
	if err != nil {
		fmt.Printf("初始化失败=%v\n", err)
	} else {
		sqlstr := fmt.Sprintf("select count(*) from `test1`")
		if data, err := db.Query(sqlstr); err == nil {
			for data.Next() {
				count := 0
				if err := data.Scan(&count); err != nil {
					fmt.Printf("gettablecount %s err:%v\n", "test1", err)
				} else {
					fmt.Printf("gettablecount %s count:%d\n", "test1", count)
				}
			}
		} else {
			fmt.Printf("gettablecount %s err:%v\n", "test1", err)
		}
	}
}
