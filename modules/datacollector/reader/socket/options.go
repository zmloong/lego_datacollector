package socket

import (
	"fmt"
	"lego_datacollector/modules/datacollector/core"

	"github.com/liwei1dao/lego/utils/mapstructure"
)

const (
	SocketRulePacket      = "按原始包读取"
	SocketRuleJson        = "按json格式读取"
	SocketRuleLine        = "按换行符读取"
	SocketRuleHeadPattern = "按行首正则读取"
)

type (
	IOptions interface {
		core.IReaderOptions //继承基础配置
		GetSocket_service_address() string
		GetSocket_read_buffer_size() int
		GetSocket_keep_alive_period() int
		GetSocket_read_timeout() int
		GetSocket_max_connections() int
		GetSocket_split_by_line() bool
		GetSocket_rule() string
		// GetEncoding() string
	}
	Options struct {
		core.ReaderOptions              //继承基础配置
		Socket_service_address   string //监听地址
		Socket_max_connections   int    //最大并发连接数 默认1024  应该是int类型的 由于需要兼容老版本只能这样
		Socket_read_timeout      int    //读的超时时间 默认30s
		Socket_read_buffer_size  int    //Socket的Buffer大小，默认65535
		Socket_keep_alive_period int    //TCP连接的keep_alive时长 默认5分钟
		Socket_split_by_line     bool   //分隔符
		Socket_rule              string //读取规则
		Encoding                 string //编码
		// IsSplitByLine            bool
	}
)

func newOptions(config map[string]interface{}) (opt IOptions, err error) {
	options := &Options{
		Socket_max_connections:   1024,
		Socket_read_buffer_size:  1024 * 64,
		Socket_keep_alive_period: 5 * 60,
		Socket_read_timeout:      30,
	}
	if config != nil {
		if err = mapstructure.Decode(config, options); err == nil {
			err = mapstructure.Decode(config, &options.ReaderOptions)
		}
	}
	if options.Socket_read_buffer_size == 0 {
		err = fmt.Errorf("Reader socket newOptions err:Socket_read_buffer_size:%d", options.Socket_read_buffer_size)
		return
	}
	// if len(options.Socket_read_buffer_size) != 0 {
	// 	if options.Read_Buffer_Szie, err = strconv.Atoi(options.Socket_read_buffer_size); err != nil {
	// 		err = fmt.Errorf("Reader socket newOptions err:Socket_read_buffer_size%s", options.Socket_read_buffer_size)
	// 		return
	// 	}
	// }
	// if len(options.Socket_keep_alive_period) != 0 {
	// 	if options.KeepAlivePeriod, err = time.ParseDuration(options.Socket_keep_alive_period); err != nil {
	// 		err = fmt.Errorf("Reader socket newOptions err:Socket_keep_alive_period%s", options.Socket_keep_alive_period)
	// 		return
	// 	}
	// }
	// if len(options.Socket_max_connections) != 0 {
	// 	if options.MaxConnections, err = strconv.Atoi(options.Socket_max_connections); err != nil {
	// 		err = fmt.Errorf("Reader socket newOptions err:Socket_max_connections%s", options.Socket_max_connections)
	// 		return
	// 	}
	// }
	// if len(options.Socket_split_by_line) != 0 {
	// 	if options.IsSplitByLine, err = strconv.ParseBool(options.Socket_split_by_line); err != nil {
	// 		err = fmt.Errorf("Reader socket newOptions err:Socket_split_by_line%s", options.Socket_split_by_line)
	// 		return
	// 	}
	// }
	// if len(options.Socket_read_timeout) != 0 {
	// 	if options.ReadTimeout, err = time.ParseDuration(options.Socket_read_timeout); err != nil {
	// 		err = fmt.Errorf("Reader socket newOptions err:Socket_read_timeout%s", options.Socket_read_timeout)
	// 		return
	// 	}
	// }
	opt = options
	return
}

func (this *Options) GetSocket_service_address() string {
	return this.Socket_service_address
}
func (this *Options) GetSocket_read_buffer_size() int {
	return this.Socket_read_buffer_size
}

func (this *Options) GetSocket_keep_alive_period() int {
	return this.Socket_keep_alive_period
}

func (this *Options) GetSocket_max_connections() int {
	return this.Socket_max_connections
}

func (this *Options) GetSocket_split_by_line() bool {
	return this.Socket_split_by_line
}

func (this *Options) GetSocket_rule() string {
	return this.Socket_rule
}
func (this *Options) GetSocket_read_timeout() int {
	return this.Socket_read_timeout
}
