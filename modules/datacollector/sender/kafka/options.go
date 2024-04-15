package kafka

import (
	"compress/gzip"
	"fmt"
	"lego_datacollector/modules/datacollector/core"
	"time"

	"github.com/Shopify/sarama"
	"github.com/mitchellh/mapstructure"
)

var (
	compressionModes = map[string]sarama.CompressionCodec{
		"none":   sarama.CompressionNone,
		"gzip":   sarama.CompressionGZIP,
		"snappy": sarama.CompressionSnappy,
		"lz4":    sarama.CompressionLZ4,
	}

	compressionLevelModes = map[string]int{
		"仅打包不压缩": gzip.NoCompression,
		"最快压缩速度": gzip.BestSpeed,
		"最高压缩比":  gzip.BestCompression,
		"默认压缩比":  gzip.DefaultCompression,
		"哈夫曼压缩":  gzip.HuffmanOnly,
	}
)

type (
	IOptions interface {
		core.ISenderOptions //继承基础配置
		GetKafka_host() []string
		GetKafka_topic() string
		GetKafka_client_id() string
		GetProducer_Compression() sarama.CompressionCodec
		GetProducer_CompressionLevel() int
		GetNet_DialTimeout() time.Duration
		GetNet_KeepAlive() time.Duration
		GetKafka_retry_max() int
	}
	Options struct {
		core.SenderOptions                                //继承基础配置
		Kafka_host                []string                `json:"kafka_host"`        //主机地址,可以有多个
		Kafka_topic               string                  `json:"kafka_topic"`       //topic 1.填一个值,则topic为所填值 2.天两个值: %{[字段名]}, defaultTopic :根据每条event,以指定字段值为topic,若无,则用默认值
		Kafka_client_id           string                  `json:"kafka_client_id"`   //客户端ID
		Kafka_compression         string                  `json:"kafka_compression"` //压缩模式,有none, gzip, snappy
		Producer_Compression      sarama.CompressionCodec //解析 Kafka_compression 字段使用
		Gzip_compression_level    string                  `json:"gzip_compression_level"` //GZIP压缩日志的策略
		Producer_CompressionLevel int                     //解析 Gzip_compression_level 字段使用
		Kafka_timeout             string                  `json:"kafka_timeout"` //连接超时时间
		Net_DialTimeout           time.Duration           //解析 Kafka_timeout 字段使用
		Kafka_keep_alive          string                  `json:"kafka_keep_alive"` //保持连接时长
		Net_KeepAlive             time.Duration           //解析 Kafka_keep_alive 字段使用
		Kafka_retry_max           int                     `json:"kafka_retry_max"` //最多充实多少次
	}
)

func newOptions(config map[string]interface{}) (opt IOptions, err error) {
	options := &Options{
		Producer_Compression:      sarama.CompressionGZIP,
		Producer_CompressionLevel: gzip.BestSpeed,
		Kafka_retry_max:           3,
		Net_DialTimeout:           time.Second * 5,
		Net_KeepAlive:             0,
	}
	if config != nil {
		if err = mapstructure.Decode(config, options); err == nil {
			err = mapstructure.Decode(config, &options.SenderOptions)
		}
	}
	var ok bool
	if options.Producer_Compression, ok = compressionModes[options.Kafka_compression]; !ok {
		err = fmt.Errorf("Sender kafka newOptions err:Kafka_compression%s", options.Kafka_compression)
		return
	}
	if options.Producer_CompressionLevel, ok = compressionLevelModes[options.Gzip_compression_level]; !ok {
		err = fmt.Errorf("Sender kafka newOptions err:Gzip_compression_level%s", options.Gzip_compression_level)
		return
	}
	if len(options.Kafka_timeout) != 0 {
		if options.Net_DialTimeout, err = time.ParseDuration(options.Kafka_timeout); err != nil {
			return
		}
	}
	if len(options.Kafka_keep_alive) != 0 {
		if options.Net_KeepAlive, err = time.ParseDuration(options.Kafka_keep_alive); err != nil {
			return
		}
	}
	opt = options
	return
}

func (this *Options) GetKafka_host() []string {
	return this.Kafka_host
}

func (this *Options) GetKafka_topic() string {
	return this.Kafka_topic
}

func (this *Options) GetKafka_client_id() string {
	return this.Kafka_client_id
}

func (this *Options) GetProducer_Compression() sarama.CompressionCodec {
	return this.Producer_Compression
}
func (this *Options) GetNet_DialTimeout() time.Duration {
	return this.Net_DialTimeout
}
func (this *Options) GetNet_KeepAlive() time.Duration {
	return this.Net_KeepAlive
}
func (this *Options) GetProducer_CompressionLevel() int {
	return this.Producer_CompressionLevel
}
func (this *Options) GetKafka_retry_max() int {
	return this.Kafka_retry_max
}
