package nacos_test

import (
	"fmt"
	"testing"

	"github.com/liwei1dao/lego/sys/sdks/aliyun/nacos"
	"github.com/nacos-group/nacos-sdk-go/vo"
)

func Test_NacosPushConfig(t *testing.T) {
	if sys, err := nacos.NewSys(
		nacos.SetNacosClientType(nacos.ConfigClient),
		nacos.SetNacosAddr("172.20.27.145"),
		nacos.SetPort(10005),
	); err != nil {
		fmt.Printf("启动nacos sys err：%v", err)
	} else {
		// sys.Config_PublishConfig(vo.ConfigParam{
		// 	DataId:  "dataId",
		// 	Group:   "group",
		// 	Content: "hello world!222222",
		// })
		content, err := sys.Config_GetConfig(vo.ConfigParam{
			DataId: "datacollector",
			Group:  "dev",
		})
		fmt.Printf("Config_GetConfig content:%v err：%v", content, err)
	}
}
