package kafka

import (
	"fmt"
	"lego_datacollector/modules/datacollector/core"

	"github.com/Shopify/sarama"
	"github.com/liwei1dao/lego/utils/mapstructure"
)

type (
	IOptions interface {
		core.IReaderOptions //继承基础配置
		GetRead_from() string
		GetKafka_version() string
		GetKafka_clientid() string
		GetKafka_groupid() string
		GetKafka_topic() []string
		GetKafka_zookeeper() []string
		Getconsumer_Offsets_Initial() int64
		GetKafka_Kerberos_Enable() bool
		GetKafka_Kerberos_Realm() string
		GetKafka_Kerberos_ServiceName() string
		GetKafka_Kerberos_Username() string
		GetKafka_Kerberos_KeyTabPath() string
		GetKafka_Kerberos_KerberosConfigPath() string
	}
	Options struct {
		core.ReaderOptions                //继承基础配置
		Read_from                         string
		Kafka_version                     string
		Kafka_clientid                    string
		Kafka_groupid                     string
		Kafka_topic                       []string
		Kafka_zookeeper                   []string
		Consumer_Offsets_Initial          int64  //解析Read_from的字段
		Kafka_Kerberos_Enable             bool   //是否开启 Kerberos 认证
		Kafka_Kerberos_Realm              string //Kerberos 认证 Realm 字段
		Kafka_Kerberos_ServiceName        string //Kerberos 认证 ServiceName 字段
		Kafka_Kerberos_Username           string //Kerberos 认证 Username 字段
		Kafka_Kerberos_KeyTabPath         string //Kerberos 认证 KeyTabPath 文件路径
		Kafka_Kerberos_KerberosConfigPath string //Kerberos 认证 KerberosConfigPath 文件路径
	}
)

func newOptions(config map[string]interface{}) (opt IOptions, err error) {
	options := &Options{
		Kafka_version:            "2.1.1",
		Kafka_clientid:           "logangent",
		Consumer_Offsets_Initial: sarama.OffsetOldest,
		Kafka_Kerberos_Enable:    false,
	}
	if config != nil {
		if err = mapstructure.Decode(config, options); err == nil {
			err = mapstructure.Decode(config, &options.ReaderOptions)
		}
	}
	if options.Read_from == "newest" { //需要兼容老版本 所以只能这样写了
		options.Consumer_Offsets_Initial = sarama.OffsetNewest
	}
	if len(options.Kafka_zookeeper) == 0 {
		err = fmt.Errorf("newOptions missing Kafka_zookeeper")
		return
	}

	if len(options.Kafka_topic) == 0 {
		err = fmt.Errorf("newOptions missing Kafka_topic")
		return
	}

	opt = options
	return
}
func (this *Options) GetRead_from() string {
	return this.Read_from
}
func (this *Options) GetKafka_version() string {
	return this.Kafka_version
}
func (this *Options) Getconsumer_Offsets_Initial() int64 {
	return this.Consumer_Offsets_Initial
}
func (this *Options) GetKafka_clientid() string {
	return this.Kafka_clientid
}
func (this *Options) GetKafka_groupid() string {
	return this.Kafka_groupid
}

func (this *Options) GetKafka_topic() []string {
	return this.Kafka_topic
}
func (this *Options) GetKafka_zookeeper() []string {
	return this.Kafka_zookeeper
}

func (this *Options) GetKafka_Kerberos_Enable() bool {
	return this.Kafka_Kerberos_Enable
}

func (this *Options) GetKafka_Kerberos_Realm() string {
	return this.Kafka_Kerberos_Realm
}

func (this *Options) GetKafka_Kerberos_ServiceName() string {
	return this.Kafka_Kerberos_ServiceName
}

func (this *Options) GetKafka_Kerberos_Username() string {
	return this.Kafka_Kerberos_Username
}

func (this *Options) GetKafka_Kerberos_KeyTabPath() string {
	return this.Kafka_Kerberos_KeyTabPath
}

func (this *Options) GetKafka_Kerberos_KerberosConfigPath() string {
	return this.Kafka_Kerberos_KerberosConfigPath
}
