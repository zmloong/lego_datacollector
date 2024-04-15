package comm

import (
	"github.com/liwei1dao/lego/core"
)

const (
	ErrorCode_StopRunnerError core.ErrorCode = 10001 //停止采集器错误
)

var ErrorCodeMsg = map[core.ErrorCode]string{
	ErrorCode_StopRunnerError: "停止采集器错误",
}

func GetErrorCodeMsg(code core.ErrorCode) string {
	if v, ok := ErrorCodeMsg[code]; ok {
		return v
	} else {
		return core.GetErrorCodeMsg(code)
	}
}
