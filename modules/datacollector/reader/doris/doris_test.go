package doris_test

import (
	"database/sql"
	"fmt"
	"testing"

	_ "github.com/go-sql-driver/mysql"
)

func Test_Doris_sql(t *testing.T) {
	db, err := sql.Open("mysql", "root@tcp(172.20.27.148:9030)/doris")
	if err != nil {
		fmt.Printf("初始化失败=%v\n", err)
	} else {
		sqlstr := fmt.Sprintf("select table_name from information_schema.tables where table_schema = 'doris' order by table_rows asc")
		if data, err := db.Query(sqlstr); err == nil {
			for data.Next() {
				tablename := ""
				if err := data.Scan(&tablename); err != nil {
					fmt.Printf("err:%v\n", err)
				} else {
					fmt.Printf("tablename:%s \n", tablename)
				}
			}
		} else {
			fmt.Printf("Test_DM_sql err:%v\n", err)
		}
	}
}

func Test_Doris_data(t *testing.T) {
	db, err := sql.Open("mysql", "root@tcp(172.20.27.148:9030)/doris")
	if err != nil {
		fmt.Printf("初始化失败=%v\n", err)
	} else {
		sqlstr := fmt.Sprintf("Select * From `test`")
		if data, err := db.Query(sqlstr); err == nil {
			if cols, err := data.Columns(); err == nil {
				fmt.Printf("cols:%v", cols)
			} else {
				fmt.Printf("err:%v", err)
			}
		} else {
			fmt.Printf("Test_DM_sql err:%v\n", err)
		}
	}
}
