package sql

type ReadInfo struct {
	Bytes int64
	Json  string //排序去重时使用，其他时候无用
}
