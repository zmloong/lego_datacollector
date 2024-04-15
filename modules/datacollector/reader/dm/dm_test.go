package dm_test

import (
	"database/sql"
	"fmt"
	"testing"

	_ "github.com/liwei1dao/dm"
)

func Test_DM_sql(t *testing.T) {
	db, err := sql.Open("dm", "dm://SYSDBA:SYSDBA@172.20.27.145:5236")
	if err != nil {
		fmt.Printf("初始化失败=%v\n", err)
	} else {
		sqlstr := fmt.Sprintf(`select table_name from all_tables`)
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
			fmt.Printf("err:%v\n", err)
		}
	}
}
