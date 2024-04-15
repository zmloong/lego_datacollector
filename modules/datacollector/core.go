package datacollector

import (
	"errors"

	"github.com/liwei1dao/lego/base"
	"github.com/liwei1dao/lego/core"
	"github.com/liwei1dao/lego/lib/modules/http"
)

type (
	RunnerState   uint8
	IDatCollector interface {
		http.IHttp
		GetService() base.IClusterService
		GetOptions() (options IOptions)
		ParamSign(param map[string]interface{}) (sign string)
		HttpStatusOK(c *http.Context, code core.ErrorCode, data interface{})
		RPC_StartRunner(rId, iid string) (result string, errstr string)
		RPC_DriveRunner(rId, iid string) (result string, errstr string)
		RPC_StopRunner(rId, iid string) (result string, errstr string)
		// RPC_DelRunner(rId string) (result string, errstr string)
	}
)

var (
	Error_NoRunner      = errors.New("runner is found")
	Error_RunnerStoping = errors.New("runner stoping")
	Error_RunnerStoped  = errors.New("runner stoped")
)

const (
	Runner_Stoped RunnerState = iota
	Runner_Starting
	Runner_Runing
)
