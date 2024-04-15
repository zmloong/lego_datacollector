package tranfroms

import (
	"fmt"
	"lego_datacollector/modules/datacollector/core"
	"strings"
	"unicode"

	"github.com/liwei1dao/lego/sys/log"
)

type (
	TFactoryFunc func(index int, runner core.IRunner, conf map[string]interface{}) (rder core.ITransforms, err error)
	//工厂
	ITransformsFactory interface {
		RegisterTransforms(stype string, f TFactoryFunc) (err error)
		NewTransforms(index int, runner core.IRunner, conf map[string]interface{}) (trans core.ITransforms, err error)
	}
	TransformInfo struct {
		CurData core.ICollData
		Index   int
	}
	TransformResult struct {
		Index   int
		CurData core.ICollData
		Err     error
		ErrNum  int
	}
)

var (
	factory ITransformsFactory
)

func init() {
	factory = newFactory()
	return
}

func NewTransforms(runner core.IRunner, conf []map[string]interface{}) (trans []core.ITransforms, err error) {
	trans = make([]core.ITransforms, len(conf))
	if factory != nil {
		for i, v := range conf {
			if trans[i], err = factory.NewTransforms(i, runner, v); err != nil {
				trans = nil
				break
			}
		}
	} else {
		err = fmt.Errorf("Parser Factory No Init")
	}
	return
}
func RegisterTransforms(rtype string, f TFactoryFunc) (err error) {
	if factory != nil {
		if err = factory.RegisterTransforms(rtype, f); err != nil {
			log.Errorf("RegisterTransforms rtype:%s err:%v", rtype, err)
		}
	} else {
		log.Errorf("RegisterTransforms Factory No Init")
	}
	return
}

//-------------------------------------------------------------------------------------------------------------------------------
//根据key字符串,拆分出层级keys数据
func GetKeys(keyStr string) []string {
	keys := strings.FieldsFunc(keyStr, isSeparator)
	return keys
}
func isSeparator(separator rune) bool {
	return separator == '.' || unicode.IsSpace(separator)
}
