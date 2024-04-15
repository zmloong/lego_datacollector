package builtin

import (
	_ "lego_datacollector/modules/datacollector/sender/doris"
	_ "lego_datacollector/modules/datacollector/sender/elasticsearch"
	_ "lego_datacollector/modules/datacollector/sender/file"
	_ "lego_datacollector/modules/datacollector/sender/hdfs"
	_ "lego_datacollector/modules/datacollector/sender/hive"
	_ "lego_datacollector/modules/datacollector/sender/kafka"
	_ "lego_datacollector/modules/datacollector/sender/mysql"
	_ "lego_datacollector/modules/datacollector/sender/pulsar"
)
