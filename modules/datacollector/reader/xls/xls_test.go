package xls_test

import (
	"fmt"
	"testing"

	"github.com/extrame/xls"
	// "github.com/shakinm/xlsReader/xls"
)

func Test_xml_create(t *testing.T) {
	// //OpenFile打开成功返回的是workBook对象
	// wb, err := xls.OpenFile("/Users/liwei1dao/work/temap/text.xlsx")
	// if err != nil {
	// 	//logger是我使用的内部日志记录器，下同
	// 	fmt.Printf("打开xls文件「%s」失败，格式有误？", "")
	// 	return
	// }
	// fmt.Println(wb.GetNumberSheets())

	// sheet, err := wb.GetSheet(0)

	// if err != nil {
	// 	fmt.Printf(err.Error())
	// }

	// // Имя листа
	// // Print sheet name
	// println(sheet.GetName())

	// // Вывести кол-во строк в листе
	// // Print the number of rows in the sheet
	// println(sheet.GetNumberRows())

	// for i := 0; i <= sheet.GetNumberRows(); i++ {
	// 	if row, err := sheet.GetRow(i); err == nil {
	// 		if cell, err := row.GetCol(1); err == nil {

	// 			// Значение ячейки, тип строка
	// 			// Cell value, string type
	// 			fmt.Println(cell.GetString())

	// 			//fmt.Println(cell.GetInt64())
	// 			//fmt.Println(cell.GetFloat64())

	// 			// Тип ячейки (записи)
	// 			// Cell type (records)
	// 			fmt.Println(cell.GetType())

	// 			// Получение отформатированной строки, например для ячеек с датой или проценты
	// 			// Receiving a formatted string, for example, for cells with a date or a percentage
	// 			xfIndex := cell.GetXFIndex()
	// 			formatIndex := wb.GetXFbyIndex(xfIndex)
	// 			format := wb.GetFormatByIndex(formatIndex.GetFormatIndex())
	// 			fmt.Println(format.GetFormatString(cell))

	// 		}

	// 	}
	// }
}

func Test_xml_read(t *testing.T) {
	res := make([][]string, 0)
	if xlFile, err := xls.Open("/Users/liwei1dao/work/temap/test001.xls", "utf-8"); err == nil {
		fmt.Println(xlFile.Author)
		//第一个sheet
		sheet := xlFile.GetSheet(0)
		if sheet.MaxRow != 0 {
			temp := make([][]string, sheet.MaxRow)
			for i := 0; i < int(sheet.MaxRow); i++ {
				row := sheet.Row(i)
				data := make([]string, 0)
				if row.LastCol() > 0 {
					for j := 0; j < row.LastCol(); j++ {
						col := row.Col(j)
						data = append(data, col)
					}
					temp[i] = data
				}
			}
			res = append(res, temp...)
		}
	}
	fmt.Printf("res:%v", res)
}
