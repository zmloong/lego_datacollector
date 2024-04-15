package sql

import (
	"database/sql"
	"fmt"
	"lego_datacollector/modules/datacollector/core"
	"strings"
	"time"
)

func SchemaCheck(rawSchemas []string) (schemas map[string]string, err error) {
	schemas = make(map[string]string)
	for _, raw := range rawSchemas {
		rs := strings.Fields(raw)
		if len(rs) != 2 {
			err = fmt.Errorf("SQL schema %v not split by space, split lens is %v", raw, len(rs))
			return
		}
		key, vtype := rs[0], rs[1]
		vtype = strings.ToLower(vtype)
		switch vtype {
		case "string", "s":
			vtype = "string"
		case "float", "f":
			vtype = "float"
		case "long", "l":
			vtype = "long"
		default:
			err = fmt.Errorf("schema type %v not supported", vtype)
			return
		}
		schemas[key] = vtype
	}
	return
}

func GetInitScans(length int, rows *sql.Rows, schemas map[string]string, runner core.IRunner, table string) (scanArgs []interface{}, nochoiced []bool) {
	nochoice := make([]interface{}, length)
	nochoiced = make([]bool, length)
	for i := range scanArgs {
		nochoice[i] = new(interface{})
		nochoiced[i] = true
	}
	defer func() {
		if r := recover(); r != nil {
			runner.Errorf("Recovered in getInitScans", r)
			scanArgs = nochoice
			return
		}
	}()

	tps, err := rows.ColumnTypes()
	if err != nil {
		runner.Errorf("GetInitScans err :%v", err)
		scanArgs = nochoice
	}
	if len(tps) != length {
		runner.Errorf("%s getInitScans length is %d not equal to columetypes %d", table, length, len(tps))
		scanArgs = nochoice
	}
	scanArgs = make([]interface{}, length)
	for i, v := range tps {
		nochoiced[i] = false
		scantype := v.ScanType().String()
		dataBaseType := v.DatabaseTypeName()
		if setDataBaseType(schemas, dataBaseType, v) {
			scanArgs[i] = new(interface{})
			continue
		}
		switch scantype {
		case "int64", "int32", "int16", "int", "int8":
			scanArgs[i] = new(interface{})
			if _, ok := schemas[v.Name()]; !ok {
				schemas[v.Name()] = "long"
			}
		case "float32", "float64":
			scanArgs[i] = new(float64)
		case "uint", "uint8", "uint16", "uint32", "uint64":
			scanArgs[i] = new(uint64)
		case "bool":
			scanArgs[i] = new(interface{})
			if _, ok := schemas[v.Name()]; !ok {
				schemas[v.Name()] = "bool"
			}
		case "[]uint8":
			scanArgs[i] = new([]byte)
		case "string", "RawBytes", "NullTime":
			scanArgs[i] = new(interface{})
			if _, ok := schemas[v.Name()]; !ok {
				schemas[v.Name()] = "string"
			}
		case "time.Time":
			scanArgs[i] = new(interface{})
			if _, ok := schemas[v.Name()]; !ok {
				schemas[v.Name()] = "date"
			}
		case "sql.NullInt64":
			scanArgs[i] = new(interface{})
			if _, ok := schemas[v.Name()]; !ok {
				schemas[v.Name()] = "long"
			}
		case "sql.NullFloat64":
			scanArgs[i] = new(interface{})
			if _, ok := schemas[v.Name()]; !ok {
				schemas[v.Name()] = "float"
			}
		default:
			scanArgs[i] = new(interface{})
			//Postgres Float的ScanType为interface,使用dataBaseType进一步判断
			if strings.Contains(dataBaseType, "FLOAT") {
				if _, ok := schemas[v.Name()]; !ok {
					schemas[v.Name()] = "float"
				}
			} else {
				nochoiced[i] = true
			}
		}
	}
	return scanArgs, nochoiced
}

func setDataBaseType(schemas map[string]string, dataBaseType string, v *sql.ColumnType) bool {
	// mysql
	switch dataBaseType {
	case "DATE", "DATETIME", "TIMESTAMP", "TIME":
		schemas[v.Name()] = "date"
		return true
	case "UNIQUEIDENTIFIER": // sqlserver
		schemas[v.Name()] = "uniqueidentifier"
		return true
	default:
		return false
	}
}

func GetOffsetIndexWithTimeStamp(offsetKey, timestampKey string, columns []string) int {
	offsetKeyIndex := -1
	for idx, key := range columns {
		if len(offsetKey) > 0 && key == offsetKey {
			return idx
		}
		if len(timestampKey) > 0 && key == timestampKey {
			return idx
		}
	}
	return offsetKeyIndex
}

func GetNextDate(format string, stringDate string, dateGap int, dateUnit string) (string, bool) {
	upperFormat := strings.ToUpper(format)
	loc, _ := time.LoadLocation("Local")
	var retDateFormat string
	var isBefore bool
	switch upperFormat {
	case "YYYYMMDD":
		const Layout = "20060102"
		dateToTime, _ := time.ParseInLocation(Layout, stringDate, loc)
		retDate := dateAdd(dateUnit, dateGap, dateToTime)
		retDateFormat = retDate.Format(Layout)

		todayFormat := time.Now().Format(Layout)
		todayToTime, _ := time.Parse(Layout, todayFormat)
		isBefore = retDate.Before(todayToTime)

	case "YYYY-MM-DD":
		const Layout = "2006-01-02"
		dateToTime, _ := time.ParseInLocation(Layout, stringDate, loc)
		retDate := dateAdd(dateUnit, dateGap, dateToTime)
		retDateFormat = retDate.Format(Layout)

		todayFormat := time.Now().Format(Layout)
		todayToTime, _ := time.Parse(Layout, todayFormat)
		isBefore = retDate.Before(todayToTime)
	case "YYYY_MM_DD":
		const Layout = "2006_01_02"
		dateToTime, _ := time.ParseInLocation(Layout, stringDate, loc)
		retDate := dateAdd(dateUnit, dateGap, dateToTime)
		retDateFormat = retDate.Format(Layout)

		todayFormat := time.Now().Format(Layout)
		todayToTime, _ := time.Parse(Layout, todayFormat)
		isBefore = retDate.Before(todayToTime)

	case "YYYYMM":
		const Layout = "200601"
		dateToTime, _ := time.ParseInLocation(Layout, stringDate, loc)
		retDate := dateAdd(dateUnit, dateGap, dateToTime)
		retDateFormat = retDate.Format(Layout)

		todayFormat := time.Now().Format(Layout)
		todayToTime, _ := time.Parse(Layout, todayFormat)
		isBefore = retDate.Before(todayToTime)

	case "YYYY-MM":
		const Layout = "2006-01"
		dateToTime, _ := time.ParseInLocation(Layout, stringDate, loc)
		retDate := dateAdd(dateUnit, dateGap, dateToTime)
		retDateFormat = retDate.Format(Layout)

		todayFormat := time.Now().Format(Layout)
		todayToTime, _ := time.Parse(Layout, todayFormat)
		isBefore = retDate.Before(todayToTime)
	case "YYYY_MM":
		const Layout = "2006_01"
		dateToTime, _ := time.ParseInLocation(Layout, stringDate, loc)
		retDate := dateAdd(dateUnit, dateGap, dateToTime)
		retDateFormat = retDate.Format(Layout)

		todayFormat := time.Now().Format(Layout)
		todayToTime, _ := time.Parse(Layout, todayFormat)
		isBefore = retDate.Before(todayToTime)

	case "YYYY":
		const Layout = "2006"
		dateToTime, _ := time.ParseInLocation(Layout, stringDate, loc)
		retDate := dateAdd(dateUnit, dateGap, dateToTime)
		retDateFormat = retDate.Format(Layout)

		todayFormat := time.Now().Format(Layout)
		todayToTime, _ := time.Parse(Layout, todayFormat)
		isBefore = retDate.Before(todayToTime)
	}
	return retDateFormat, isBefore
}

func dateAdd(dateUnit string, dateGap int, toTime time.Time) time.Time {
	var retDate time.Time
	switch dateUnit {
	case "day":
		retDate = toTime.AddDate(0, 0, dateGap)
	case "month":
		retDate = toTime.AddDate(0, dateGap, 0)
	case "year":
		retDate = toTime.AddDate(dateGap, 0, 0)
	}
	return retDate
}
