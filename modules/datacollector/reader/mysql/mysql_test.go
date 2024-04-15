package mysql_test

import (
	"database/sql"
	"fmt"
	"testing"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/liwei1dao/dm"
)

func Test_Oracle_sql(t *testing.T) {
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

func Test_DM_sql_01(t *testing.T) {
	db, err := sql.Open("dm", "dm://SYSDBA:SYSDBA@172.20.27.145:5236")
	if err != nil {
		fmt.Printf("初始化失败=%v\n", err)
	} else {
		sqlstr := fmt.Sprintf("select table_name from all_tables")
		if data, err := db.Query(sqlstr); err == nil {
			for data.Next() {
				tablename := ""
				if err := data.Scan(&tablename); err != nil {
					fmt.Printf("err:%v\n", err)
				} else {
					fmt.Printf("tablename:%s\n", tablename)
				}
			}
		} else {
			fmt.Printf("Test_DM_sql err:%v\n", err)
		}
	}
}

func Test_DM_sql_02(t *testing.T) {
	db, err := sql.Open("dm", "dm://SYSDBA:SYSDBA@172.20.27.145:5236")
	if err != nil {
		fmt.Printf("初始化失败=%v\n", err)
	} else {
		sqlstr := fmt.Sprintf("SELECT * FROM test.TAB2")
		if data, err := db.Query(sqlstr); err == nil {
			if cols, err := data.Columns(); err == nil {
				fmt.Printf("cols:%v", cols)
			} else {
				fmt.Printf("err:%v", err)
			}

			// for data.Next() {
			// 	tablename := ""
			// 	if err := data.Scan(&tablename); err != nil {
			// 		fmt.Printf("err:%v\n", err)
			// 	} else {
			// 		fmt.Printf("tablename:%s\n", tablename)
			// 	}
			// }
		} else {
			fmt.Printf("Test_DM_sql err:%v\n", err)
		}
	}
}
