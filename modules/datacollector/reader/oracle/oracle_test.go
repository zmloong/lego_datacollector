package oracle_test

import (
	"database/sql"
	"fmt"
	"testing"

	_ "github.com/godror/godror"
	lsql "github.com/liwei1dao/lego/sys/sql"
)

func Test_Oracle(t *testing.T) {
	sys, err := lsql.NewSys(
		lsql.SetSqlType(lsql.Oracle),
		//jdbc:oracle:thin:@192.168.1.100:1521:oracle
		lsql.SetSqlUrl("idss_sjzt/idss1234@172.20.27.125:1521/nek"),
	)
	if err != nil {
		fmt.Printf("初始化失败=%v\n", err)
	} else {
		fmt.Printf("初始化成功\n")
		if data, err := sys.Query(`select t.table_name,t.num_rows from all_tables t`); err == nil {
			if coluns, err := data.Columns(); err == nil {
				fmt.Printf("coluns:%v\n", coluns)
			} else {
				fmt.Printf("coluns err:%v\n", err)
			}
			table_name := ""
			table_count := 0
			for data.Next() {
				if err := data.Scan(&table_name, &table_count); err == nil {
					fmt.Printf("table_name:%s table_count:%d\n", table_name, table_count)
				} else {
					fmt.Printf("tablename err :%v\n", err)
				}
			}
		} else {
			fmt.Printf("Query err:%v\n", err)
		}
	}
}

func Test_Oracle_sql(t *testing.T) {
	db, err := sql.Open("godror", "idss_sjzt/idss1234@172.20.27.125:1521/nek")
	if err != nil {
		fmt.Printf("初始化失败=%v\n", err)
	} else {
		sqlstr := fmt.Sprintf(`select count(*) from "test1"`)
		if data, err := db.Query(sqlstr); err == nil {
			for data.Next() {
				count := 0
				if err := data.Scan(&count); err != nil {
					fmt.Printf("gettablecount %s err:%v\n", "test1", err)
				} else {
					fmt.Printf("gettablecount %s count:%d\n", "test1", count)
				}
			}
		}
	}
}
