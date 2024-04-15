package doris

import (
	"lego_datacollector/modules/datacollector/core"

	"github.com/liwei1dao/lego/utils/mapstructure"
)

const (
	CollectionSender SenderDorisType = iota
	SynchronizeSender
)

type (
	SenderDorisType int8
	IOptions        interface {
		core.ISenderOptions //继承基础配置
		GetDoris_sendertype() SenderDorisType
		GetDoris_column_separator() string
		GetDoris_ip() string
		GetDoris_sql_port() int
		GetDoris_stream_port() int
		GetDoris_user() string
		GetDoris_password() string
		GetDoris_dbname() string
		GetDoris_tableName() string
		GetDoris_mapping() map[string]string
	}

	Options struct {
		core.SenderOptions     //继承基础配置
		Doris_sendertype       SenderDorisType
		Doris_ip               string
		Doris_sql_port         int
		Doris_stream_port      int
		Doris_user             string
		Doris_password         string
		Doris_dbname           string
		Doris_tableName        string
		Doris_column_separator string
		Doris_mapping          map[string]string //数据库同步映射表
	}
)

func newOptions(config map[string]interface{}) (opt IOptions, err error) {
	options := &Options{
		Doris_sendertype:       CollectionSender,
		Doris_column_separator: "|",
	}
	if config != nil {
		if err = mapstructure.Decode(config, options); err == nil {
			err = mapstructure.Decode(config, &options.SenderOptions)
		}
		if err != nil {
			return
		}
	}

	opt = options
	return
}
func (this *Options) GetDoris_sendertype() SenderDorisType {
	return this.Doris_sendertype
}

func (this *Options) GetDoris_column_separator() string {
	return this.Doris_column_separator
}

func (this *Options) GetDoris_ip() string {
	return this.Doris_ip
}

func (this *Options) GetDoris_sql_port() int {
	return this.Doris_sql_port
}
func (this *Options) GetDoris_stream_port() int {
	return this.Doris_stream_port
}

func (this *Options) GetDoris_user() string {
	return this.Doris_user
}
func (this *Options) GetDoris_password() string {
	return this.Doris_password
}
func (this *Options) GetDoris_dbname() string {
	return this.Doris_dbname
}
func (this *Options) GetDoris_tableName() string {
	return this.Doris_tableName
}
func (this *Options) GetDoris_mapping() map[string]string {
	return this.Doris_mapping
}
