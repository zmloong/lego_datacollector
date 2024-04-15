@[TOC](datacollector  api接口文档)

### Api签名密钥
密钥:@88199%67g12q4*67m12#4l67!
签名方式: md5加密 k1=v1&k2=v2&key=@88199%67g12q4*67m12#4l67!
签名输出全小写:例 hduwqueqhieicuhiueqiueh

### 创建采集任务_kafka_接口Api
1. 请求地址:/datacollector/addrunner
2. 请求方式:Post
3. 请求参数方式:body-json
4. 请求参数:
```
{
    "name":"kafka_001",                                                                             //采集任务id 必填 string
    "instanceid":"kafka_001_1231231231",                             //任务的实例Id
    "ip": ["172.20.27.145"],                                                                           //任务运行地址 必填 string (默认 0.0.0.0 自动分配)
    “maxprocs”:4,                                                                                   //最大采集并发数据  选填 int (默认 4)
    “max_colldata_size”:4194304                                                                     //最大采集数据大小  选填 int (默认 4m)
    "reader":{
        "type":"kafka",                                                                             //读取器类型 必填 string
        "encoding":"utf-8",                                                                         //采集解析编码 必填 string 默认 utf-8
        "autoClose":false,                                                                          //是否自动关闭 自动关闭 -> 全量采集
        "kafka_groupid": "logagent",                                                                //kafaka消费组id 必填 string
        "kafka_topic": ["ETL-IN-CX202108051438145021423171470628098048"],                           //kafka消费topic 必填 []string
        "kafka_zookeeper": ["172.20.27.126:9092","172.20.27.127:9092","172.20.27.128:9092"],        //kafka集群地址 必填 []string
        "read_from": "oldest",                                                                       //kafka 消费方式 oldest｜newest
        "kafka_clientid": "logangent",
        "kafka_Kerberos_Enable":true,                                                               //是否开启 Kerberos 认证
        "kafka_Kerberos_Realm":"TUJL.COM",                                                          //Kerberos 认证 Realm 字段
        "kafka_Kerberos_ServiceName":"kafka",                                                       //Kerberos 认证 ServiceName 字段
        "kafka_Kerberos_Username":"kafka/sjzt-wuhan-13",                                            //Kerberos 认证 Username 字段
        "kafka_Kerberos_KeyTabPath":"./sjzt-wuhan-13.keytab",                                       //Kerberos 认证 KeyTabPath 文件路径 提前讲文件放在采集节点服务器上
        "kafka_Kerberos_KerberosConfigPath":"./krb5.conf"                                           //Kerberos 认证 KerberosConfigPath 文件路径 提前讲文件放在采集节点服务器上
    },
    "parser":{
        "type":"raw",                                                                               //解析器类型 必填 string 
        "idss_collect_time": true,                                                                  //是否添加采集时间 选填 bool 
        "idss_source_ip":true                                                                       //是否添加采集ip 选填 bool 
    },
    "transforms":[
        {     
            "type": "avro"                                      //avro 解码消息
            "schema":"",                                        //avro schema
        },
       {     
            "type": "datamask"                                                                      //数据脱敏 aes 对称加密
            "signkey":"123456781234567812345678",                                                   //加密密钥 长度必须固定 16, 24, 32
            "keys":["key1","key2","key3"]                                                           //需要加密的字段列表
        },
        { 
            "type": "expmatch"                                                                      //正则过滤
            "expmatch":"(H)ello",                                                                   //正则
            "isnegate":false,                                                                       //是否取反
            "isdiscard":false                                                                       //是否丢弃 
        }
    ],
    "senders":[
        {
            "type":"kafka",                                                                         //发送器类型 必填 string 
            "kafka_host":["192.168.0.1:9092","192.168.0.1:9093"],                                   //kafka集群地址 必填 []string 
            "kafka_topic":"ETL-IN-CX202108051412338611423165008711389184",                        //kafka生产topic 必填 []string 
            "kafka_client_id":"logangent",                                                          //kafka生产client 必填 string 
            "kafka_compression":"gzip",                                                             //kafka生产消息压缩方式 必填 string
            "gzip_compression_level":"最快压缩速度"                                                   //kafka生产压缩级别 必填 string
        }
    ],
    "isclean":true,                         //是否清理上一次的记录数据    
    "sign":"16daf080ddc15133ebe243a67db88fde"                                                       //签名 字段 [name]
}
``` 
5. 输出参数
```
{
    "code":0,                                                   //非 0 表示请求错误
    "message":"成功",
    "data":null,
}
```

### 创建采集任务_socket_接口Api
1. 请求地址:/datacollector/addrunner
2. 请求方式:Post
3. 请求参数方式:body-json
4. 请求参数:
```
{
    "name": "socket_001",
    "ip": ["172.20.27.145"], 
    “maxprocs”:4,                                       
    "max_colldata_size": 4194304,
    "reader": {
        "type": "socket",
        "encoding":"utf-8",                                                      //采集解析编码 必填 string 默认 utf-8
        "autoClose":false,                                                                          //是否自动关闭 自动关闭 -> 全量采集
        "socket_service_address": "tcp://0.0.0.0:6378",                          //socket监听地址 必填 string 
        default:tcp://0.0.0.0:6378 or udp://0.0.0.0:6378
        "socket_max_connections":0,                                             //最大连接数 选填 int default:1024 (0 表示没有限制)
        "socket_read_timeout":10,                                               //读取超时 选填 int default:10 (单位秒)
        "socket_keep_alive_period":60,                                          //读取超时 选填 int default:10 (单位秒)
        "socket_read_buffer_size":65535,                                        //读取缓存 选填 int default:4096
        "socket_split_by_line":true,                                            //分割符 选填 int
        "socket_rule":""                                                        //解析规则 选填 int
    },
    "parser": {
        "type": "raw",
        "idss_collect_time": true,
        "idss_source_ip": true
    },
    "senders": [
        {
            "type": "kafka",
            "kafka_host": [
                "172.20.27.126:9092",
                "172.20.27.127:9092",
                "172.20.27.128:9092"
            ],
            "kafka_topic": "ETL-IN-CX202108051412338611423165008711389184",
            "kafka_client_id": "logangent",
            "kafka_compression": "gzip",
            "gzip_compression_level": "最快压缩速度"
        }
    ],
    "sign": "c8c3ccd0d247818bb86e1e85209f9a4f"
}
``` 
5. 输出参数
```
{
    "code":0,                                                   //非 0 表示请求错误
    "message":"成功",
    "data":null,
}
```



### 创建采集任务_ftp_接口Api
1. 请求地址:/datacollector/addrunner
2. 请求方式:Post
3. 请求参数方式:body-json
4. 请求参数:
```
{
    "name": "ftp_001",
    "ip": ["172.20.27.145"], 
    “maxprocs”:4,                                                                           
    "max_colldata_size": 4194304,
    "reader": {
        "type": "ftp",
        "encoding":"utf-8",                                 //采集解析编码 必填 string 默认 utf-8
        "autoClose":false,                                                                          //是否自动关闭 自动关闭 -> 全量采集
        "ftp_server": "49.235.111.75",                      //ftp地址  必填 string
        "ftp_port":21,                                      //ftp端口  选填 int (默认 21)
        "ftp_user":"zmlftp",                                //ftp账号  必填 string 
        "ftp_password":"123456",                            //ftp密码  必填 string 
        "ftp_directory":"/var/ftp"                          //ftp采集目录  必填 string
        "ftp_regularrules":"\\.(log|txt)$",                 //ftp采集正则过滤 必填 string 默认 "\\.(log|txt)$"
        "ftp_read_buffer_size":4096,                        //单次采集的数据量
        “ftp_interval”：“0 */1 * * * ?”,                    //定时采集 默认一分钟执行一次
        “ftp_exec_onstart”：true,                           //启动时执行一次
    },
    "parser": {
        "type": "raw",
        "idss_collect_time": true,
        "idss_source_ip": true
    },
    "senders": [
        {
            "type": "kafka",
            "kafka_host": [
                "172.20.27.126:9092",
                "172.20.27.127:9092",
                "172.20.27.128:9092"
            ],
            "kafka_topic": "ETL-IN-CX202108051412338611423165008711389184",
            "kafka_client_id": "logangent",
            "kafka_compression": "gzip",
            "gzip_compression_level": "最快压缩速度"
        }
    ],
    "sign": "756f04a3a5df900b32774b2d21157425"
}
``` 
5. 输出参数
```
{
    "code":0,                                                   //非 0 表示请求错误
    "message":"成功",
    "data":null,
}
```


### 创建采集任务_sftp_接口Api
1. 请求地址:/datacollector/addrunner
2. 请求方式:Post
3. 请求参数方式:body-json
4. 请求参数:
```
{
    "name": "ftp_001",
    "ip": ["172.20.27.145"], 
    “maxprocs”:4,                                                                           
    "max_colldata_size": 4194304,
    "reader": {                                                                       
        "type": "sftp",
        "encoding":"utf-8",                                 //采集解析编码 必填 string 默认 utf-8
        "autoClose":false,                                  //是否自动关闭 自动关闭 -> 全量采集
        "sftp_server": "49.235.111.75",
        "sftp_port":22,
        "sftp_user":"root",
        "sftp_password":"*511215949a",
        "sftp_directory":"/var/ftp",                         //sftp采集目录  必填 string
        "sftp_read_buffer_size":4096,                         //单次采集数据量
        "sftp_regularrules":"\\.(log|txt)$",                 //sftp采集正则过滤 必填 string 默认 "\\.(log|txt)$"
        “sftp_interval”：“0 */1 * * * ?”,                    //定时采集 默认一分钟执行一次
        “sftp_exec_onstart”：true,                           //启动时执行一次
    },
    "parser": {
        "type": "raw",
        "idss_collect_time": true,
        "idss_source_ip": true
    },
    "senders": [
        {
            "type": "kafka",
            "kafka_host": [
                "172.20.27.126:9092",
                "172.20.27.127:9092",
                "172.20.27.128:9092"
            ],
            "kafka_topic": "ETL-IN-CX202108051412338611423165008711389184",
            "kafka_client_id": "logangent",
            "kafka_compression": "gzip",
            "gzip_compression_level": "最快压缩速度"
        }
    ],
    "sign": "756f04a3a5df900b32774b2d21157425"
}
``` 
5. 输出参数
```
{
    "code":0,                                                   //非 0 表示请求错误
    "message":"成功",
    "data":null,
}
```

### 创建采集任务_mysql静态表_接口Api
1. 请求地址:/datacollector/addrunner
2. 请求方式:Post
3. 请求参数方式:body-json
4. 请求参数:
```
{
    "name": "mysql_001",
    "ip": ["172.20.27.145"], 
    “maxprocs”:4,                                                                           
    "max_colldata_size": 4194304,
    "reader": {                                                                       
        "type": "mysql",
        "encoding":"utf-8",                                             //采集解析编码 必填 string 默认 utf-8
        "autoClose":false,                                              //是否自动关闭 自动关闭 -> 全量采集
        "Mysql_collection_type": 0,                                     //0 表示静态表，1 表示动态表
        "Mysql_datasource": "root:Idss@sjzt2021@tcp(172.20.27.125:3306)",
        "Mysql_database": "mysql",
        "Mysql_sql": "Select * From `$TABLE$`",    
        "Mysql_offset_key": "INCREMENTAL",                              //自增主键字段，offset记录依据，默认INCREMENTAL
        "Mysql_limit_batch": 100,                                       //每次查询多少条数据
        "Mysql_batch_intervel": "0 */1 * * * ?",                        //扫描间隔，cron定时任务查询，默认 1m 
        "Mysql_exec_onstart": true,                                     //是否启动时查询一次
        "Mysql_tables": [
            "test"
        ],
        "MySql_dateUnit": "day",                               //year 年表 month 月表 day日表
        "MySql_dateGap": 1,                                    //动态表 的是假间隔 默认 1 (几年一表 几月一表 几天一表)
        "MySql_datetimeFormat": "YYYY-MM-DD",
        "MySql_startDate": "2021-08-01"
    },
    "parser": {
        "type": "raw",
        "idss_collect_time": true,
        "idss_source_ip": true
    },
    "senders": [
        {
            "type": "mysql",
            "mysql_datasource": "root:Idss@sjzt2021@tcp(172.20.27.125:3306)",
            "mysql_database": "mysql",
            "mysql_table": "test001",
            "mysql_mapping": {"a1":"a1","b1":"b1"}
        }
    ],
    "sign": "54f2cbf8c903a843abc6a7eb9e9010ce"
}
``` 
5. 输出参数
```
{
    "code":0,                                                   //非 0 表示请求错误
    "message":"成功",
    "data":null,
}
```

### 创建采集任务_mysql动态表_接口Api
1. 请求地址:/datacollector/addrunner
2. 请求方式:Post
3. 请求参数方式:body-json
4. 请求参数:
```
{
    "name": "mysql_002",
    "instanceid":"mysql_002_1231231231",                             //任务的实例Id
    "ip": ["172.20.27.145"], 
    “maxprocs”:4,                                                                           
    "max_colldata_size": 4194304,
    "reader": {                                                                       
        "type": "mysql",
        "encoding":"utf-8",                                           //采集解析编码 必填 string 默认 utf-8
        "autoClose":false,                                            //是否自动关闭 自动关闭 -> 全量采集
        "Mysql_collection_type": 1,                                   //0 表示静态表，1 表示动态表
        "Mysql_datasource": "root:Idss@sjzt2021@tcp(172.20.27.125:3306)",
        "Mysql_database": "mysql",
        "Mysql_sql": "Select * From `$TABLE$`",
        "Mysql_offset_key": "INCREMENTAL",                            //自增主键字段，offset记录依据，默认INCREMENTAL
        "Mysql_limit_batch": 100,                                     //每次查询多少条数据
        "Mysql_batch_intervel": "0 */1 * * * ?",                      //扫描间隔，cron定时任务查询，默认 1m
        "Mysql_exec_onstart": true,                                   //是否启动时查询一次
        "Mysql_tables": [
            "test_"                                                   //动态表前缀字段 匹配test_2021-08-01
        ],
        "MySql_dateUnit": "day",
        "MySql_dateGap": 1,
        "MySql_datetimeFormat": "YYYY-MM-DD",  //时间格式对应动态表名匹配：
                                               //支持：YYYY-MM-DD,YYYYMMDD,YYYY_MM_DD,YYYYMM,YYYY-MM,YYYY_MM,YYYY
        "MySql_startDate": "2021-08-01"
    },
    "parser": {
        "type": "raw",
        "idss_collect_time": true,
        "idss_source_ip": true
    },
    "senders": [
        {
            "type": "kafka",
            "kafka_host": [
                "172.20.27.126:9092",
                "172.20.27.127:9092",
                "172.20.27.128:9092"
            ],
            "kafka_topic": "ETL-IN-CX202108051412338611423165008711389184",
            "kafka_client_id": "logangent",
            "kafka_compression": "gzip",
            "gzip_compression_level": "最快压缩速度"
        }
    ],
    "isclean":true,                         //是否清理上一次的记录数据                                        
    "sign": "0753f3e2517c89310fab7a94663c9634"
}
``` 
5. 输出参数
```
{
    "code":0,                                                   //非 0 表示请求错误
    "message":"成功",
    "data":null,
}
```

### 创建采集任务_oracle静态表_接口Api
1. 请求地址:/datacollector/addrunner
2. 请求方式:Post
3. 请求参数方式:body-json
4. 请求参数:
```
{
    "name": "oracle_001",
    "ip": ["172.20.27.145"], 
    “maxprocs”:4,                                                                           
    "max_colldata_size": 4194304,
    "reader": {                                                                       
        "type": "oracle",
        "autoClose":false,                                            //是否自动关闭 自动关闭 -> 全量采集
        "Oracle_collection_type": 0,
        "Oracle_datasource": "idss_sjzt/idss1234@172.20.27.125:1521",    //oracle连接字符串
        "Oracle_sql":"select "$TABLE$".*,ROWNUM as INCREMENTAL From "$TABLE$"",
                                                //sql语句中 
        "Oracle_database": "nek",               //oracle服务名
        "Oracle_offset_key": "INCREMENTAL",     //自增主键字段，offset记录依据，默认INCREMENTAL
        "Oracle_limit_batch": 100,
        "Oracle_batch_intervel": "0 */1 * * * ?",
        "Oracle_exec_onstart": true,
        "Oracle_tables": [
            "test"
        ],
        "Oracle_dateUnit": "day",
        "Oracle_dateGap": 1,
        "Oracle_datetimeFormat": "YYYY-MM-DD",
        "Oracle_startDate": "2021-08-26"
    },
    "parser": {
        "type": "raw",
        "idss_collect_time": true,
        "idss_source_ip": true
    },
    "senders": [
        {
            "type": "kafka",
            "kafka_host": [
                "172.20.27.126:9092",
                "172.20.27.127:9092",
                "172.20.27.128:9092"
            ],
            "kafka_topic": "ETL-IN-CX202108051412338611423165008711389184",
            "kafka_client_id": "logangent",
            "kafka_compression": "gzip",
            "gzip_compression_level": "最快压缩速度"
        }
    ],
    "sign": "0753f3e2517c89310fab7a94663c9634"
}
``` 
5. 输出参数
```
{
    "code":0,                                                   //非 0 表示请求错误
    "message":"成功",
    "data":null,
}
```

### 创建采集任务_oracle动态表_接口Api
1. 请求地址:/datacollector/addrunner
2. 请求方式:Post
3. 请求参数方式:body-json
4. 请求参数:
```
{
    "name": "oracle_002",
    "ip": ["172.20.27.145"], 
    “maxprocs”:4,                                                                           
    "max_colldata_size": 4194304,
    "reader": {                                                                       
        "type": "oracle",
        "autoClose":false,                                            //是否自动关闭 自动关闭 -> 全量采集
        "Oracle_collection_type": 1,
        "Oracle_datasource": "idss_sjzt/idss1234@172.20.27.125:1521",       //oracle连接字符串
        "Oracle_sql":"select "$TABLE$".*,ROWNUM as INCREMENTAL From "$TABLE$"",
                                                //sql语句中 IDSS_SJZT 为库名 
        "Oracle_database": "nek",               //oracle服务名
        "Oracle_offset_key": "INCREMENTAL",     //自增主键字段，offset记录依据，默认INCREMENTAL
        "Oracle_limit_batch": 100,
        "Oracle_batch_intervel": "0 */1 * * * ?",
        "Oracle_exec_onstart": true,
        "Oracle_tables": [
            "test_"
        ],
        "Oracle_dateUnit": "day",
        "Oracle_dateGap": 1,
        "Oracle_datetimeFormat": "YYYY-MM-DD",
        "Oracle_startDate": "2021-09-01"
    },
    "parser": {
        "type": "raw",
        "idss_collect_time": true,
        "idss_source_ip": true
    },
    "senders": [
        {
            "type": "kafka",
            "kafka_host": [
                "172.20.27.126:9092",
                "172.20.27.127:9092",
                "172.20.27.128:9092"
            ],
            "kafka_topic": "ETL-IN-CX202108051412338611423165008711389184",
            "kafka_client_id": "logangent",
            "kafka_compression": "gzip",
            "gzip_compression_level": "最快压缩速度"
        }
    ],
    "sign": "0753f3e2517c89310fab7a94663c9634"
}
``` 
5. 输出参数
```
{
    "code":0,                                                   //非 0 表示请求错误
    "message":"成功",
    "data":null,
}
```



### 创建采集任务_dm静态表_接口Api
1. 请求地址:/datacollector/addrunner
2. 请求方式:Post
3. 请求参数方式:body-json
4. 请求参数:
```
  "name": "dm_001",
    "ip": ["172.20.27.145"], 
    “maxprocs”:4,                                                                           
    "max_colldata_size": 4194304,
    "reader": {                                                                       
        "type": "dm",
        "encoding":"utf-8",                                          //采集解析编码 必填 string 默认 utf-8
        "autoClose":false,                                            //是否自动关闭 自动关闭 -> 全量采集
        "DM_collection_type": 0,                                     //0 表示静态表，1 表示动态表
        "DM_datasource": "dm://SYSDBA:SYSDBA@172.20.27.145:5236",
        "DM_database": "test",
        "DM_sql": "select * From $TABLE$",    
        "DM_offset_key": "INCREMENTAL",                              //自增主键字段，offset记录依据，默认INCREMENTAL
        "DM_limit_batch": 100,                                       //每次查询多少条数据
        "DM_batch_intervel": "0 */1 * * * ?",                        //扫描间隔，cron定时任务查询，默认 1m 
        "DM_exec_onstart": true,                                     //是否启动时查询一次
        "DM_tables": [
            "TAB2"
        ],
        "DM_dateUnit": "day",                               //year 年表 month 月表 day日表
        "DM_dateGap": 1,                                    //动态表 的是假间隔 默认 1 (几年一表 几月一表 几天一表)
        "DM_datetimeFormat": "YYYY-MM-DD",
        "DM_startDate": "2021-08-01"
    },
    "parser": {
        "type": "raw",
        "idss_collect_time": true,
        "idss_source_ip": true
    },
    "senders": [
        {
            "type": "kafka",
            "kafka_host": [
                "172.20.27.126:9092",
                "172.20.27.127:9092",
                "172.20.27.128:9092"
            ],
            "kafka_topic": "ETL-IN-CX202108051412338611423165008711389184",
            "kafka_client_id": "logangent",
            "kafka_compression": "gzip",
            "gzip_compression_level": "最快压缩速度"
        }
    ],
    "sign": "54f2cbf8c903a843abc6a7eb9e9010ce"
}
``` 
5. 输出参数
```
{
    "code":0,                                                   //非 0 表示请求错误
    "message":"成功",
    "data":null,
}
```


### 创建采集任务_mongodb静态表_接口Api
1. 请求地址:/datacollector/addrunner
2. 请求方式:Post
3. 请求参数方式:body-json
4. 请求参数:
```
  "name": "mongodb_001",
    "ip": ["172.20.27.145"], 
    “maxprocs”:4,                                                                           
    "max_colldata_size": 4194304,
    "reader": {                                                                       
      "type": "mongodb",
        "encoding":"utf-8",
        "autoClose":false,                                            //是否自动关闭 自动关闭 -> 全量采集
        "Mongodb_collection_type": 0,
        "Mongodb_datasource": "mongodb://172.20.27.145:10002",
        "Mongodb_database": "datacollector",
        "Mongodb_limit_batch": 100,
        "Mongodb_batch_intervel": "0 */1 * * * ?",
        "Mongodb_exec_onstart": true,
        "Mongodb_tables": [
            "configs"
        ],
        "Mongodb_dateUnit": "day",
        "Mongodb_dateGap": 1,
        "Mongodb_datetimeFormat": "YYYY-MM-DD",
        "Mongodb_startDate": "2021-08-01"
    },
    "parser": {
        "type": "raw",
        "idss_collect_time": true,
        "idss_source_ip": true
    },
    "senders": [
        {
            "type": "kafka",
            "kafka_host": [
                "172.20.27.126:9092",
                "172.20.27.127:9092",
                "172.20.27.128:9092"
            ],
            "kafka_topic": "ETL-IN-CX202108051412338611423165008711389184",
            "kafka_client_id": "logangent",
            "kafka_compression": "gzip",
            "gzip_compression_level": "最快压缩速度"
        }
    ],
    "sign": "54f2cbf8c903a843abc6a7eb9e9010ce"
}
``` 
5. 输出参数
```
{
    "code":0,                                                   //非 0 表示请求错误
    "message":"成功",
    "data":null,
}
```

### 创建采集任务_redis静态表_接口Api
1. 请求地址:/datacollector/addrunner
2. 请求方式:Post
3. 请求参数方式:body-json
4. 请求参数:
```
  "name": "redis_001",
    "ip": ["172.20.27.145"], 
    “maxprocs”:4,                                                                           
    "max_colldata_size": 4194304,
    "reader": {                                                                       
        "type": "redis",
        "encoding":"utf-8",
        "autoClose":false,                                            //是否自动关闭 自动关闭 -> 全量采集
        "Redis_type": 0,                                        //0 单节点  1集群
        "Redis_url": ["172.20.27.145:10001"],                   //地址
        "Redis_db": 1,                                          //数据卷
        "Redis_password": "li13451234",                         //密码
        "Redis_pattern":"*",                                    //匹配key
        "Redis_autoClose":true,                                 //是否自动关闭
        "Redis_intervel": "0 */1 * * * ?"                       //增量查询时间间隔
    },
    "parser": {
        "type": "raw",
        "idss_collect_time": true,
        "idss_source_ip": true
    },
    "senders": [
        {
            "type": "kafka",
            "kafka_host": [
                "172.20.27.126:9092",
                "172.20.27.127:9092",
                "172.20.27.128:9092"
            ],
            "kafka_topic": "ETL-IN-CX202108051412338611423165008711389184",
            "kafka_client_id": "logangent",
            "kafka_compression": "gzip",
            "gzip_compression_level": "最快压缩速度"
        }
    ],
    "sign": "54f2cbf8c903a843abc6a7eb9e9010ce"
}
``` 
5. 输出参数
```
{
    "code":0,                                                   //非 0 表示请求错误
    "message":"成功",
    "data":null,
}
```

### 创建采集任务_snmp_接口Api
1. 请求地址:/datacollector/addrunner
2. 请求方式:Post
3. 请求参数方式:body-json
4. 请求参数:
```
{
    "name": "snmp_001",
    "ip": ["172.20.27.145"], 
    “maxprocs”:4,                                                                           
    "max_colldata_size": 4194304,
    "reader": {
        "type": "snmp",
        "encoding":"utf-8",                          //采集解析编码 必填 string 默认 utf-8
        "autoClose":false,                           //是否自动关闭 自动关闭 -> 全量采集                                                                       
        "snmp_retries": "3",                         //连接失败后重试次数
        "snmp_version": "2",                         //snmp协议版本 ：1,2,3
        "snmp_time_out": "5s", 
        "snmp_table_init_host": "49.235.111.75",     //初始连接地址
        "snmp_community": "public",
        "snmp_agents": "49.235.111.75:161",          //连接被管理设备地址端口，支持多个，逗号间隔
        "snmp_max_repetitions": "50",
        "snmp_sec_level": "noAuthNoPriv",            //协议版为 3 时需要,默认值：noAuthNoPriv
        "snmp_priv_protocol": "NoPriv",              //协议版为 3 时需要,默认值：NoPriv
        "snmp_auth_protocol": "NoAuth",              //协议版为 3 时需要,默认值：NoAuth
        "snmp_interval": "30s",
        "snmp_fields": "[{\"field_name\":\"sysName\",\"field_oid\":\".1.3.6.1.2.1.1.5.0\"}]",  
        "snmp_tables": "[{\"table_name\":\"sysName\",\"table_oid\":\"1.3.6.1.2.1.25.4.2\"}]"
        //snmp_fields和snmp_tables数组至少传一个
    },
    "parser": {
        "type": "raw",
        "idss_collect_time": true,
        "idss_source_ip": true
    },
    "senders": [
        {
            "type": "kafka",
            "kafka_host": [
                "172.20.27.126:9092",
                "172.20.27.127:9092",
                "172.20.27.128:9092"
            ],
            "kafka_topic": "ETL-IN-CX202108051412338611423165008711389184",
            "kafka_client_id": "logangent",
            "kafka_compression": "gzip",
            "gzip_compression_level": "最快压缩速度"
        }
    ],
    "sign": "a55d121914925f7a12f0b1199f43d19e"
}
``` 
5. 输出参数
```
{
    "code":0,                                                   //非 0 表示请求错误
    "message":"成功",
    "data":null,
}
```

### 创建采集任务_snmptrap_接口Api
1. 请求地址:/datacollector/addrunner
2. 请求方式:Post
3. 请求参数方式:body-json
4. 请求参数:
```
{
    "name": "snmptrap_001",
    "ip": ["172.20.27.145"], 
    “maxprocs”:4,                                                                           
    "max_colldata_size": 4194304,
    "reader": {                                                                       
        "type":"snmptrap",
        "encoding":"utf-8",                       //采集解析编码 必填 string 默认 utf-8
        "autoClose":false,                                            //是否自动关闭 自动关闭 -> 全量采集
        "snmptrap_address": "udp://0.0.0.0:9162", //监听地址端口号：udp://0.0.0.0:9162,tcp://0.0.0.0:9162
        "snmptrap_time_out": "5s",                //超时时间 默认 30s
        "snmptrap_interval": "60s"                //超时时间 默认 30s
    },
    "parser": {
        "type": "raw",
        "idss_collect_time": true,
        "idss_source_ip": true
    },
    "senders": [
        {
            "type": "kafka",
            "kafka_host": [
                "172.20.27.126:9092",
                "172.20.27.127:9092",
                "172.20.27.128:9092"
            ],
            "kafka_topic": "ETL-IN-CX202108051412338611423165008711389184",
            "kafka_client_id": "logangent",
            "kafka_compression": "gzip",
            "gzip_compression_level": "最快压缩速度"
        }
    ],
    "sign": "1374f3ea4766fbf3f4ad0aa6acfb1001"
}
``` 
5. 输出参数
```
{
    "code":0,                                                   //非 0 表示请求错误
    "message":"成功",
    "data":null,
}
```


### 创建采集任务_netflow_接口Api
1. 请求地址:/datacollector/addrunner
2. 请求方式:Post
3. 请求参数方式:body-json
4. 请求参数:
```
{
    "name": "netflow_001",
    "ip": ["172.20.27.145"], 
    “maxprocs”:4,                                                                           
    "max_colldata_size": 4194304,
    "reader": {                                                                       
        "type": "netflow",
        "encoding":"utf-8",                                           //采集解析编码 必填 string 默认 utf-8
        "autoClose":false,                                            //是否自动关闭 自动关闭 -> 全量采集
        "netflow_listen_port": 9998,
        "netflow_read_size": 4096,
    },
    "parser": {
        "type": "raw",
        "idss_collect_time": true,
        "idss_source_ip": true
    },
    "senders": [
        {
            "type": "kafka",
            "kafka_host": [
                "172.20.27.126:9092",
                "172.20.27.127:9092",
                "172.20.27.128:9092"
            ],
            "kafka_topic": "ETL-IN-CX202108051412338611423165008711389184",
            "kafka_client_id": "logangent",
            "kafka_compression": "gzip",
            "gzip_compression_level": "最快压缩速度"
        }
    ],
    "sign": "a55d121914925f7a12f0b1199f43d19e"
}
``` 
5. 输出参数
```
{
    "code":0,                                                   //非 0 表示请求错误
    "message":"成功",
    "data":null,
}
```

### 创建采集任务_netcat_接口Api
1. 请求地址:/datacollector/addrunner
2. 请求方式:Post
3. 请求参数方式:body-json
4. 请求参数:
```
{
    "name": "netcat_001",
    "ip": ["172.20.27.145"], 
    “maxprocs”:4,                                                                           
    "max_colldata_size": 4194304,
    "reader": {                                                                       
        "type": "netcat",
        "encoding":"utf-8",                                           //采集解析编码 必填 string 默认 utf-8
        "autoClose":false,                                            //是否自动关闭 自动关闭 -> 全量采集
        "netcat_listen_port": 9998,
    },
    "parser": {
        "type": "raw",
        "idss_collect_time": true,
        "idss_source_ip": true
    },
    "senders": [
        {
            "type": "kafka",
            "kafka_host": [
                "172.20.27.126:9092",
                "172.20.27.127:9092",
                "172.20.27.128:9092"
            ],
            "kafka_topic": "ETL-IN-CX202108051412338611423165008711389184",
            "kafka_client_id": "logangent",
            "kafka_compression": "gzip",
            "gzip_compression_level": "最快压缩速度"
        }
    ],
    "sign": "a55d121914925f7a12f0b1199f43d19e"
}
``` 
5. 输出参数
```
{
    "code":0,                                                   //非 0 表示请求错误
    "message":"成功",
    "data":null,
}
```

### 创建采集任务_file_接口Api
1. 请求地址:/datacollector/addrunner
2. 请求方式:Post
3. 请求参数方式:body-json
4. 请求参数:
```
{
    "name": "netcat_001",
    "ip": ["172.20.27.145"], 
    “maxprocs”:4,                                                                           
    "max_colldata_size": 4194304,
    "reader": {                                                                       
       "type": "file",
        "autoClose":false,                                            //是否自动关闭 自动关闭 -> 全量采集
        "file_path":"/Users/liwei1dao/work/temap/logkit_test/reder/liwei1dao.log",                      //采集目标文件路径
        "file_readbuff_size":4096,                                                                      //采集文件缓存区大小 默认 4096
        "file_scan_initial":“0 */1 * * * ?”,                                                            //采集扫描间隔 默认 一分钟
        “file_exec_onstart”：true                                                                       //启动时执行一次
    },
    "parser": {
        "type": "raw",
        "idss_collect_time": true,
        "idss_source_ip": true
    },
    "senders": [
        {
            "type": "kafka",
            "kafka_host": [
                "172.20.27.126:9092",
                "172.20.27.127:9092",
                "172.20.27.128:9092"
            ],
            "kafka_topic": "ETL-IN-CX202108051412338611423165008711389184",
            "kafka_client_id": "logangent",
            "kafka_compression": "gzip",
            "gzip_compression_level": "最快压缩速度"
        }
    ],
    "sign": "a55d121914925f7a12f0b1199f43d19e"
}
``` 
5. 输出参数
```
{
    "code":0,                                                   //非 0 表示请求错误
    "message":"成功",
    "data":null,
}
```




### 创建采集任务_folder_接口Api
1. 请求地址:/datacollector/addrunner
2. 请求方式:Post
3. 请求参数方式:body-json
4. 请求参数:
```
{
    "name": "netcat_001",
    "ip": ["172.20.27.145"], 
    “maxprocs”:4,                                                                           
    "max_colldata_size": 4194304,
    "reader": {                                                                       
       "type": "folder",
        "autoClose":false,                                            //是否自动关闭 自动关闭 -> 全量采集
        "folder_directory":"/Users/liwei1dao/work/temap/logkit_test/reder/",          //采集目录
        "folder_regular_filter":"\\.(log|txt)$",                                      //采集正则过滤 
        "folder_scan_initial":"0 */1 * * * ?"                                         //采集扫描间隔 默认 60 秒
        “folder_max_collection_num”:1000,                                             //最大文件采集数 默认 1000
        “folder_readerbuf_size”:4096                                                  //采集缓存大小 默认 4k 越大单次采集耗时越久 任务堵塞时间越久
        “folder_exec_onstart”：true                                                   //启动时采集一次
    },
    "parser": {
        "type": "raw",
        "idss_collect_time": true,
        "idss_source_ip": true
    },
    "senders": [
        {
            "type": "kafka",
            "kafka_host": [
                "172.20.27.126:9092",
                "172.20.27.127:9092",
                "172.20.27.128:9092"
            ],
            "kafka_topic": "ETL-IN-CX202108051412338611423165008711389184",
            "kafka_client_id": "logangent",
            "kafka_compression": "gzip",
            "gzip_compression_level": "最快压缩速度"
        }
    ],
    "sign": "a55d121914925f7a12f0b1199f43d19e"
}
``` 
5. 输出参数
```
{
    "code":0,                                                   //非 0 表示请求错误
    "message":"成功",
    "data":null,
}
```

### 创建采集任务_sqlserver_接口Api
1. 请求地址:/datacollector/addrunner
2. 请求方式:Post
3. 请求参数方式:body-json
4. 请求参数:
```
{
    "name": "sqlserver_01",
    "ip": ["172.20.27.999"],
    "maxprocs": 8,
    "max_colldata_size": 4194304,
    "reader": {
        "type": "mssql",                                    //采集任务类型
        "encoding": "utf-8",
        "autoClose":false,                                            //是否自动关闭 自动关闭 -> 全量采集
        "Mssql_collection_type": 0,                         //0 表示静态表，1 表示动态表
        "Mssql_datasource": "server=127.0.0.1;port:1433;database=test;user id=sa;password=123456;encrypt=disable",                                   //数据库连接地址
        "Mssql_database": "test",                           //数据库名
        "Mssql_sql": "select row_number() over (order by (select 0)) as INCREMENTAL,* from $TABLE$",  
                                                            //查询语句
        "Mssql_offset_key": "INCREMENTAL",                  //自增主键字段，offset记录依据，默认INCREMENTAL
        "Mssql_limit_batch": 1000,                          //每次查询多少条数据
        "Mssql_batch_intervel": "0 */1 * * * ?",            //扫描间隔，cron定时任务查询，默认 1m
        "Mssql_exec_onstart": true,                         //是否启动时查询一次
        "Mssql_tables": [
            "zml"
        ],
        "Mssql_dateUnit": "day",                            //year 年表 month 月表 day日表
        "Mssql_dateGap": 1,                                 //动态表 的是假间隔 默认 1 (几年一表 几月一表 几天一表)
        "Mssql_datetimeFormat": "YYYY-MM-DD",               //时间格式
        "Mssql_startDate": "2021-11-01"                     //扫描表的 起始日期
    },
    "parser": {
        "type": "raw",
        "idss_collect_time": true,
        "idss_source_ip": true
    },
    "senders": [
        {
            "type": "file",
            "file_send_path": "/workplace/gopath/src/test/data/lego_mssql.txt"
        }
    ],
    "sign": "4b4172d4c8668637e5034aefa6d27298"
}
``` 
5. 输出参数
```
{
    "code":0,                                                   //非 0 表示请求错误
    "message":"成功",
    "data":null,
}
```

### 创建采集任务_PostgrSQL_接口Api
1. 请求地址:/datacollector/addrunner
2. 请求方式:Post
3. 请求参数方式:body-json
4. 请求参数:
```
{
    "name": "sqlserver_01",
    "ip": ["172.20.27.999"],
    "maxprocs": 8,
    "max_colldata_size": 4194304,
        "reader": {
        "type": "postgre",                      //采集任务类型
        "encoding": "utf-8",                    //编码方式
        "autoClose":false,                                            //是否自动关闭 自动关闭 -> 全量采集
        "Postgre_collection_type": 0,           //0 表示静态表，1 表示动态表
        "Postgre_addr": "127.0.0.1:5432",       //数据库连接地址
        "Postgre_database": "postgres",         //数据库名
        "Postgre_user": "postgres",             //用户名
        "Postgre_password": "postgre",          //密码
        "Postgre_sql": "select row_number() over () as INCREMENTAL,* from $TABLE$",  //查询语句
        "Postgre_offset_key": "INCREMENTAL",    //自增主键字段，offset记录依据，默认
        "Postgre_limit_batch": 1000,            //每次查询多少条数据
        "Postgre_batch_intervel": "0 */1 * * * ?",   //扫描间隔，cron定时任务查询，默认 1m
        "Postgre_exec_onstart": true,                //是否启动时查询一次
        "Postgre_tables": [                         
            "test"
        ],
        "Postgre_dateUnit": "day",                  //year 年表 month 月表 day日表
        "Postgre_dateGap": 1,               //动态表 的是假间隔 默认 1 (几年一表 几月一表 几天一表)
        "Postgre_datetimeFormat": "YYYY-MM-DD",    //时间格式
        "Postgre_startDate": "2021-11-01"          //扫描表的 起始日期
    },
    "parser": {
        "type": "raw",
        "idss_collect_time": true,
        "idss_source_ip": true
    },
    "senders": [
        {
            "type": "file",
            "file_send_path": "/workplace/gopath/src/test/data/lego_mssql.txt"
        }
    ],
    "sign": "4b4172d4c8668637e5034aefa6d27298"
}
``` 
5. 输出参数
```
{
    "code":0,                                                   //非 0 表示请求错误
    "message":"成功",
    "data":null,
}
```

### 创建采集任务_hdfs_接口Api（sender为输出hdfs参数）
1. 请求地址:/datacollector/addrunner
2. 请求方式:Post
3. 请求参数方式:body-json
4. 请求参数:
```
{
    "name": "hdfs_read_01",
    "ip": ["172.20.27.999"],
    "maxprocs": 8,
    "max_colldata_size": 4194304,
    "reader": {
        "type": "hdfs",                                        //采集任务类型
        "encoding": "utf-8",
        "autoClose":false,                                            //是否自动关闭 自动关闭 -> 全量采集
        "hdfs_addr": "172.20.27.128:8020,172.20.27.126:8020",  //hdfs连接地址
        "hdfs_user": "hdfs",                                   //hdfs连接用户名                    
        "hdfs_directory": "/tmp/zmltest/read",                 //hdfs访问目录地址
        "hdfs_regularrules": "\\.(log|txt)$",                  //正则过滤 必填 string 默认 "\\.(log|txt)$"
        "hdfs_max_collection_num": 1000,                        //最大采集文件数 默认 1000
        "hdfs_Kerberos_Enable": true,           //是否开启kerberos认证
        "hdfs_Kerberos_Realm": "HADOOP.COM",    //Kerberos 认证 Realm 字段
        "hdfs_Kerberos_ServiceName": "hdfs/sjzt-wuhan-9",  //Kerberos 认证 ServiceName 字段
        "hdfs_Kerberos_Username": "hdfs/sjzt-wuhan-9",    //Kerberos 认证 Username 字段
        "hdfs_Kerberos_KeyTabPath": "./conf/hdfs.keytab", //Kerberos 认证 KeyTabPath 文件路径
        "hdfs_Kerberos_KerberosConfigPath": "./conf/krb5.conf", //Kerberos 认证 KerberosConfigPath 文件路径
        "hdfs_Kerberos_HadoopConfPath": "./conf"  //Kerberos 认证 core-site.xml和 hdfs-site.xml文件路径
    },
    "parser": {
        "type": "raw",
        "idss_collect_time": true,
        "idss_source_ip": true
    },
    "senders": [
         {
            "type":"hdfs",                                      //输出类型
            "hdfs_addr":"172.20.27.128:8020,172.20.27.126:8020",//输出地址
            "hdfs_user":"hdfs",                                 //输出用户名
            "hdfs_path":"/tmp/zmltest",                         //输出目录
            "hdfs_size":134217728,                              //输出时，文件切割的大小阀值 默认128M
            "hdfs_Kerberos_Enable": true,           //是否开启kerberos认证
            "hdfs_Kerberos_Realm": "HADOOP.COM",    //Kerberos 认证 Realm 字段
            "hdfs_Kerberos_ServiceName": "hdfs/sjzt-wuhan-9",  //Kerberos 认证 ServiceName 字段
            "hdfs_Kerberos_Username": "hdfs/sjzt-wuhan-9",    //Kerberos 认证 Username 字段
            "hdfs_Kerberos_KeyTabPath": "./conf/hdfs.keytab", //Kerberos 认证 KeyTabPath 文件路径
            "hdfs_Kerberos_KerberosConfigPath": "./conf/krb5.conf", //Kerberos 认证 KerberosConfigPath 文件路径
            "hdfs_Kerberos_HadoopConfPath": "./conf"  //Kerberos 认证 core-site.xml和 hdfs-site.xml文件路径
        }
    ],
    "sign": "34744b64bd686ae4cb03f522304b4022"
}
``` 
5. 输出参数
```
{
    "code":0,                                                   //非 0 表示请求错误
    "message":"成功",
    "data":null,
}
```
### 创建采集任务输出hive参数 sender
1. 请求地址:/datacollector/addrunner
2. 请求方式:Post
3. 请求参数方式:body-json
4. 请求参数:
```
{
    "name": "pg-hive02",
    "ip": [
        "172.20.27.999"
    ],
    "maxprocs": 4,
    "max_colldata_size": 4194304,
    "reader": {
        "type": "postgre",
        "encoding": "utf-8",
        "Postgre_collection_type": 0,
        "Postgre_addr": "127.0.0.1:5432",
        "Postgre_database": "postgres",
        "Postgre_user": "postgres",
        "Postgre_password": "postgre",
        "Postgre_sql": "select row_number() over () as INCREMENTAL,* from $TABLE$",
        "Postgre_offset_key": "INCREMENTAL",
        "Postgre_limit_batch": 1000,
        "Postgre_batch_intervel": "0 */1 * * * ?",
        "Postgre_exec_onstart": true,
        "Postgre_tables": [
            "test"
        ],
        "Postgre_dateUnit": "day",
        "Postgre_dateGap": 1,
        "Postgre_datetimeFormat": "YYYYMMDD",
        "Postgre_startDate": "20220405"
    },
    "parser": {
        "type": "raw",
        "idss_collect_time": true,
        "idss_source_ip": true
    },
    "senders": [
        {
            "type": "hive",                          //采集任务类型
            "hdfs_addr": "172.20.27.128:8020,172.20.27.126:8020", // hdfs 连接地址
            "hdfs_user": "hdfs",                     //输出用户名
            "hdfs_path": "/tmp/zmltest",             //输出目录
            "hdfs_size": 134217728,                  //文件切割大小设置 
            "hive_addr": "172.20.27.126:10000",      //hive 连接地址
            "hive_username": "hdfs",                 //hive用户名
            "hive_password": "",                     //hive用户密码
            "hive_auth": "NONE",             //hive鉴权模式 需设置hive-site.xml中hive.server.authentication值
            "hive_sendertype": 0,            // 0 代表不指定字段映射 已默认消息结构体存储  1 指定字段映射 匹配存储
            "hive_dbname": "datacollector",  //hive 库名
            "hive_tableName": "jsontest",    //hive 表名
            "hive_mapping": {                //指定数据源字段和hive表字段映射
                "a": "a",
                "b": "b",
                "c": "c"
            },
            "hive_column_separator":" ",     //数据生成hdfs文件分割符，需与hive建表时数据分隔符一致
            "hive_json":true                 //在hive_sendertype为0时启用，已json结构体存入hive表
        }
    ],
    "sign": "1c25eeaf171f6a13977adbde3b75c006"
}
``` 
5. 输出参数
```
{
    "code":0,                                                   //非 0 表示请求错误
    "message":"成功",
    "data":null,
}
```
### 创建采集任务_elastic_接口Api（sender为输出elastic参数）
1. 请求地址:/datacollector/addrunner
2. 请求方式:Post
3. 请求参数方式:body-json
4. 请求参数:
```
{
    "name": "es_read_01",
    "ip": ["172.20.27.999"],
    "maxprocs": 8,
    "max_colldata_size": 4194304,
    "reader":  {
        "type": "elastic",                                    //采集任务类型
        "autoClose":false,                                            //是否自动关闭 自动关闭 -> 全量采集
        "Elastic_collection_type": 0,                         ////0 表示静态表，1 表示动态表
        "Elastic_taskselfclose":0,                    //采集完成后是否关闭任务 0不关闭（增量）1关闭（全量）
        "Elastic_addr": "172.20.27.144:9200",                 //es连接地址
        "Elastic_limit_batch": 100,                           //每次采集读取的数据条数限制 默认100
        "Elastic_batch_intervel": "0 */1 * * * ?",            //定时任务 corn表达式
        "Elastic_exec_onstart": true,                         //任务创建时 是否立即执行一次采集
        "Elastic_tables": [ 
            "zmltest_local"                                    //采集索引名
        ],
        "Elastic_offset_key": "id",                     //指定自增id字段 int型
        "Elastic_timestamp_key":"time",                 //时间范围查询字段  时间格式："yyyy-MM-dd HH:mm:ss"
        "Elastic_query_starttime":"2021-12-03 00:00:00",//时间范围开始时间
        "Elastic_query_endtime":"2021-12-04 00:00:00",  //时间范围结束时间
        "Elastic_dateUnit": "day",                             //year 年表 month 月表 day日表
        "Elastic_dateGap": 1,                                  //时间间隔对应上面单位（年月日） 默认 1 (几年一表 几月一表 几天一表)
        "Elastic_datetimeFormat": "YYYY-MM-DD",                //时间格式
        "Elastic_startDate": "2020-11-01"                      //起始时间
    },
    "parser": {
        "type": "raw",
        "idss_collect_time": true,
        "idss_collect_ip": true
    },
    "senders": [
        {
            "type":"elasticsearch",                         //输出类型
            "es_host":"http://172.20.27.144:9200",          //输出地址
            "es_index":"zmltest001"                         //输出索引名
        }
    ],
    "sign": "ef430fe1719cf01efc408a120ab3899d"
}
``` 
5. 输出参数
```
{
    "code":0,                                                   //非 0 表示请求错误
    "message":"成功",
    "data":null,
}
```

### 创建采集任务_httpfetch_接口Api
1. 请求地址:/datacollector/addrunner
2. 请求方式:Post
3. 请求参数方式:body-json
4. 请求参数:
```
{
    "name": "httpfetch_read_01",
    "ip": [
        "172.20.27.999"
    ],
    "maxprocs": 8,
    "max_colldata_size": 4194304,
    "reader": {
        "type": "httpfetch",                            //采集任务类型 http请求对方获取数据
        "autoClose":false,                                            //是否自动关闭 自动关闭 -> 全量采集
        "Httpfetch_method": "GET",                      //http请求方式 GET，POST
        "Httpfetch_headers": "{\"Content-Type\": \"text/plain\"}",//http请求头部
        "Httpfetch_address": "127.0.0.1:8000",      //http请求地址 域名或ip+端口
        "Httpfetch_body": "",                       //http请求参数body
        "Httpfetch_exec_onstart": true,             //任务创建时 是否立即执行一次采集
        "Httpfetch_dialTimeout": 30,                //连接超时时间
        "Httpfetch_respTimeout": 30,                //请求超时时间
        "Httpfetch_interval": "0 */1 * * * ?"       //定时任务 corn表达式  
    },
    "parser": {
        "type": "raw",
        "idss_collect_time": true,
        "idss_collect_ip": true
    },
    "senders": [
        {
            "type": "file",
            "file_send_path": "/workplace/gopath/src/test/data/lego_httpfetch.txt"
        }
    ],
    "sign": "3142f980162bcd6d25014d91fe70d38f"
}
``` 
5. 输出参数
```
{
    "code":0,                                                   //非 0 表示请求错误
    "message":"成功",
    "data":null,
}
```

### 创建采集任务_canal_接口Api
1. 请求地址:/datacollector/addrunner
2. 请求方式:Post
3. 请求参数方式:body-json
4. 请求参数:
```
{
    "name": "canal_001",
    "ip": ["172.20.27.145"], 
    “maxprocs”:4,                                                                           
    "max_colldata_size": 4194304,
    "reader": {                                                                       
        "type": "canal",
        "encoding":"utf-8",
        "autoClose":false,                                            //是否自动关闭 自动关闭 -> 全量采集
        "canal_addr": "172.20.27.145",                                          //canal 服务地址
        "canal_port": 11111,                                                    //canal 服务端口
        "canal_user": "canal",                                                  //canal 用户名
        "canal_password": "canal",                                              //canal 密码
        "canal_destination": "example",                                         //配置集
        /*订阅过滤字段
        // https://github.com/alibaba/canal/wiki/AdminGuide
        //mysql 数据解析关注的表，Perl正则表达式.
        //多个正则之间以逗号(,)分隔，转义符需要双斜杠(\\)
        //常见例子：
        //  1.  所有表：.*   or  .*\\..*
        //	2.  canal schema下所有表： canal\\..*
        //	3.  canal下的以canal打头的表：canal\\.canal.*
        //	4.  canal schema下的一张表：canal\\.test1
        //  5.  多个规则组合使用：canal\\..*,mysql.test1,mysql.test2 (逗号分隔)
        */
        "canal_filter": "mysql\\.test9"                                         //订阅过滤字段
    },
    "parser": {
        "type": "raw",
        "idss_collect_time": true,
        "idss_source_ip": true
    },
    "senders": [
        {
            "type": "kafka",
            "kafka_host": [
                "172.20.27.126:9092",
                "172.20.27.127:9092",
                "172.20.27.128:9092"
            ],
            "kafka_topic": "ETL-IN-CX202108051412338611423165008711389184",
            "kafka_client_id": "logangent",
            "kafka_compression": "gzip",
            "gzip_compression_level": "最快压缩速度"
        }
    ],
    "sign": "a55d121914925f7a12f0b1199f43d19e"
}
``` 
5. 输出参数
```
{
    "code":0,                                                   //非 0 表示请求错误
    "message":"成功",
    "data":null,
}
```

### 创建采集任务_xml_接口Api
1. 请求地址:/datacollector/addrunner
2. 请求方式:Post
3. 请求参数方式:body-json
4. 请求参数:
```
{
    "name": "xml_001",
    "ip": ["172.20.27.145"], 
    “maxprocs”:4,                                                                           
    "max_colldata_size": 4194304,
    "reader": {                                                                       
        "type": "xml",
        "encoding":"utf-8",
        "autoClose":false,                                            //是否自动关闭 自动关闭 -> 全量采集
        "xml_collectiontype":0,                             //0 本地采集 1 ftp采集 2 sftp采集
        "xml_max_collection_size":10485760,                 //采集最大文件 10m 默认
        "Xml_max_collection_num":1000,                      //采集最大文件量 1000 默认
        "Xml_server_addr":"12.19.18.78",                    //ftp/sftp 服务地址
        "Xml_server_port":21,                               //ftp/sftp 服务端口
        "Xml_server_user":"",                               //ftp/sftp 服务账号
        "Xml_server_password":"",                           //ftp/sftp 服务密码
        "Xml_node_root":"ROOT",                             //xml数据root节点名称
        "Xml_node_data":"LOG4A",                            //xml数据采集数据节点名称
        "Xml_directory":"/Users/liwei1dao/work/temap",      //采集目录
        "Xml_interval":"0 */1 * * * ?",                     //定时采集
        "Xml_regularrules":"\\.(xml)$",                     //文件正则过滤 默认 \.(xml)$
        "Xml_exec_onstart":true                             //启动时扫描
    },
    "parser": {
        "type": "raw",
        "idss_collect_time": true,
        "idss_source_ip": true
    },
    "senders": [
        {
            "type": "kafka",
            "kafka_host": [
                "172.20.27.126:9092",
                "172.20.27.127:9092",
                "172.20.27.128:9092"
            ],
            "kafka_topic": "ETL-IN-CX202108051412338611423165008711389184",
            "kafka_client_id": "logangent",
            "kafka_compression": "gzip",
            "gzip_compression_level": "最快压缩速度"
        }
    ],
    "sign": "a55d121914925f7a12f0b1199f43d19e"
}
``` 
5. 输出参数
```
{
    "code":0,                                                   //非 0 表示请求错误
    "message":"成功",
    "data":null,
}
```

### 创建采集任务_xls_接口Api
1. 请求地址:/datacollector/addrunner
2. 请求方式:Post
3. 请求参数方式:body-json
4. 请求参数:
```
{
    "name": "xls_001",
    "ip": ["172.20.27.145"], 
    “maxprocs”:4,                                                                           
    "max_colldata_size": 4194304,
    "reader": {                                                                       
        "type": "xls",
        "encoding":"utf-8",
        "autoClose":false,                                            //是否自动关闭 自动关闭 -> 全量采集
        "xls_collectiontype":0,                             //0 本地采集 1 ftp采集 2 sftp采集
        "xls_max_collection_size":10485760,                 //采集最大文件 10m 默认
        "xls_max_collection_num":1000,                      //采集最大文件量 1000 默认
        "xls_server_addr":"",                               //ftp/sftp 服务地址
        "xls_server_port":21,                               //ftp/sftp 服务端口
        "xls_server_user":"",                               //ftp/sftp 服务账号
        "xls_server_password":"",                           //ftp/sftp 服务密码
        "Xls_key_line":0,                                   //key所在行
        "xls_data_line":1,                                  //数据查询起始行
        "xls_directory":"/Users/liwei1dao/work/temap",      //采集目录
        "xls_interval":"0 */1 * * * ?",                     //定时采集
        "xls_regularrules":"\\.(xls)$",                     //正则扫描
        "xls_exec_onstart":true                             //启动时扫描
    },
    "parser": {
        "type": "raw",
        "idss_collect_time": true,
        "idss_source_ip": true
    },
    "senders": [
        {
            "type": "kafka",
            "kafka_host": [
                "172.20.27.126:9092",
                "172.20.27.127:9092",
                "172.20.27.128:9092"
            ],
            "kafka_topic": "ETL-IN-CX202108051412338611423165008711389184",
            "kafka_client_id": "logangent",
            "kafka_compression": "gzip",
            "gzip_compression_level": "最快压缩速度"
        }
    ],
    "sign": "a55d121914925f7a12f0b1199f43d19e"
}
``` 
5. 输出参数
```
{
    "code":0,                                                   //非 0 表示请求错误
    "message":"成功",
    "data":null,
}
```

### 创建采集任务_xlsx_接口Api
1. 请求地址:/datacollector/addrunner
2. 请求方式:Post
3. 请求参数方式:body-json
4. 请求参数:
```
{
    "name": "xlsx_001",
    "ip": ["172.20.27.145"], 
    “maxprocs”:4,                                                                           
    "max_colldata_size": 4194304,
    "reader": {                                                                       
        "type": "xlsx",
        "encoding":"utf-8",
        "autoClose":false,                                            //是否自动关闭 自动关闭 -> 全量采集
        "xlsx_collectiontype":0,                             //0 本地采集 1 ftp采集 2 sftp采集
        "xlsx_max_collection_size":10485760,                 //采集最大文件 10m 默认
        "xlsx_max_collection_num":1000,                      //采集最大文件量 1000 默认
        "xlsx_server_addr":"",                               //ftp/sftp 服务地址
        "xlsx_server_port":21,                               //ftp/sftp 服务端口
        "xlsx_server_user":"",                               //ftp/sftp 服务账号
        "xlsx_server_password":"",                           //ftp/sftp 服务密码
        "Xlsx_key_line":0,                                   //key所在行
        "xlsx_data_line":1,                                  //数据查询起始行
        "xlsx_directory":"/Users/liwei1dao/work/temap",      //采集目录
        "xlsx_interval":"0 */1 * * * ?",                     //定时采集
        "xlsx_regularrules":"\\.(xlsx)$",                //正则扫描
        "xlsx_exec_onstart":true                             //启动时扫描
    },
    "parser": {
        "type": "raw",
        "idss_collect_time": true,
        "idss_source_ip": true
    },
    "senders": [
        {
            "type": "kafka",
            "kafka_host": [
                "172.20.27.126:9092",
                "172.20.27.127:9092",
                "172.20.27.128:9092"
            ],
            "kafka_topic": "ETL-IN-CX202108051412338611423165008711389184",
            "kafka_client_id": "logangent",
            "kafka_compression": "gzip",
            "gzip_compression_level": "最快压缩速度"
        }
    ],
    "sign": "a55d121914925f7a12f0b1199f43d19e"
}
``` 
5. 输出参数
```
{
    "code":0,                                                   //非 0 表示请求错误
    "message":"成功",
    "data":null,
}
```
\

### 创建采集任务_webservice_接口Api
1. 请求地址:/datacollector/addrunner
2. 请求方式:Post
3. 请求参数方式:body-json
4. 请求参数:
```
{
    "name": "xlsx_001",
    "ip": ["172.20.27.145"], 
    “maxprocs”:4,                                                                           
    "max_colldata_size": 4194304,
    "reader": {
		"type": "webservice",                                                       //采集任务类型
        "autoClose":false,                                            //是否自动关闭 自动关闭 -> 全量采集
		"Webservice_address": "http://www.webxml.com.cn/WebServices/WeatherWebService.asmx",//请求地址
        
        //webservice接口的请求体body
		"Webservice_reqbody": "<?xml version=\"1.0\" encoding=\"utf-8\"?> <soap12:Envelope xmlns:xsi=\"http://www.w3.org/2001/XMLSchema-instance\" xmlns:xsd=\"http://www.w3.org/2001/XMLSchema\" xmlns:soap12=\"http://www.w3.org/2003/05/soap-envelope\">   <soap12:Body>     <getSupportCity xmlns=\"http://WebXml.com.cn/\">       <byProvinceName>北京</byProvinceName>     </getSupportCity>   </soap12:Body> </soap12:Envelope>",  
       

		"Webservice_xml_body_node": "getSupportCityResult",   //返回数据解析起始节点，会处理该节点下全部节点数据
		"Webservice_interval": "0 */1 * * * ?",   //扫描间隔，cron定时任务查询，默认
		"Webservice_exec_onstart": true,          //启动时是否扫描执行一次
		"Webservice_respTimeout": 30              //webservice接口请求超时时间
	},
    "parser": {
        "type": "raw",
        "idss_collect_time": true,
        "idss_source_ip": true
    },
    "senders": [
        {
            "type": "kafka",
            "kafka_host": [
                "172.20.27.126:9092",
                "172.20.27.127:9092",
                "172.20.27.128:9092"
            ],
            "kafka_topic": "ETL-IN-CX202108051412338611423165008711389184",
            "kafka_client_id": "logangent",
            "kafka_compression": "gzip",
            "gzip_compression_level": "最快压缩速度"
        }
    ],
    "sign": "a55d121914925f7a12f0b1199f43d19e"
}
``` 
5. 输出参数
```
{
    "code":0,                                                   //非 0 表示请求错误
    "message":"成功",
    "data":null,
}
```

### 创建采集任务_compress(ftp压缩包)_接口Api
1. 请求地址:/datacollector/addrunner
2. 请求方式:Post
3. 请求参数方式:body-json
4. 请求参数:
```
{
    "name": "ftp-gz_read_01",
    "ip": ["172.20.27.145"], 
    "maxprocs":4,                                                                           
    "max_colldata_size": 4194304,
    "reader": {
		"type": "compress",
        "autoClose":false,                                            //是否自动关闭 自动关闭 -> 全量采集
		"Compress_collectiontype": 0,                           //ftp类型
		"Compress_max_collection_size": 1073741824,             //采集压缩包大小限制             
		"Compress_server_addr": "49.235.111.75",                //ftp服务地址
		"Compress_server_port":21,                              //ftp服务端口
		"Compress_server_user":"zmlftp",                        //ftp服务用户名
        "Compress_server_password":"123456",                    //ftp服务密码
        "Compress_directory": "/data/ftp",                      //ftp服务访问目录
        "Compress_interval": "0 */1 * * * ?",                   //定时扫描，采集频率 默认1m
        "Compress_regularrules": "\\.(tar.gz)$",                //正则匹配压缩文件后缀名 默认tar.gz
        "Compress_inter_regularrules": "\\.(bin|txt)$",         //正则匹配压缩文件内部，需采集文件后缀名
        "Compress_xml_regularrules":"\\.(bin)$",                //匹配压缩包内部 为xml文件内容的后缀名
        "Compress_xml_root_node":"getSupportCityResult",        //xml文件解析的根节点
        "Compress_exec_onstart": true                           //是否启动执行
	},
    "parser": {
        "type": "raw",
        "idss_collect_time": true,
        "idss_source_ip": true
    },
    "senders": [
        {
            "type": "kafka",
            "kafka_host": [
                "172.20.27.126:9092",
                "172.20.27.127:9092",
                "172.20.27.128:9092"
            ],
            "kafka_topic": "ETL-IN-CX202108051412338611423165008711389184",
            "kafka_client_id": "logangent",
            "kafka_compression": "gzip",
            "gzip_compression_level": "最快压缩速度"
        }
    ],
    "sign": "a55d121914925f7a12f0b1199f43d19e"
}
``` 
5. 输出参数
```
{
    "code":0,                                                   //非 0 表示请求错误
    "message":"成功",
    "data":null,
}
```
### 创建采集任务_hbase_接口Api
1. 请求地址:/datacollector/addrunner
2. 请求方式:Post
3. 请求参数方式:body-json
4. 请求参数:
```
{
    "name": "Hbase_01",
    "ip": ["172.20.27.999"],
    "maxprocs": 8,
    "max_colldata_size": 4194304,
    "reader": {
        "type": "hbase",
        "encoding": "utf-8",
        "Hbase_collection_type": 1,                                    //0静态表 1动态表
        "Hbase_host": "172.20.27.126,172.20.27.127,172.20.27.128",     //连接地址
        "Hbase_limit_batch": 1000,
        "Hbase_batch_intervel": "0 */1 * * * ?",                       //定时扫描
        "Hbase_exec_onstart": true,
        "Hbase_tables": [
            "student_"                                                 //表名
        ],
        "Hbase_dateUnit": "day",
        "Hbase_dateGap": 1,
        "Hbase_datetimeFormat": "YYYYMMDD",
        "Hbase_startDate": "20220110"
    },
    "parser": {
        "type": "raw",
        "idss_collect_time": true,
        "idss_collect_ip": true,
        "eqpt_ip": true
    },
    "senders": [
        {
            "type": "kafka",
            "kafka_host": [
                "172.20.27.126:9092",
                "172.20.27.127:9092",
                "172.20.27.128:9092"
            ],
            "kafka_topic": "ETL-IN-CX202108051412338611423165008711389184",
            "kafka_client_id": "logangent",
            "kafka_compression": "gzip",
            "gzip_compression_level": "最快压缩速度"
        }
    ],
    "sign": "041c9a5fb1b3f3cf062a0116dd680215"
}
``` 
5. 输出参数
```
{
    "code":0,                                                   //非 0 表示请求错误
    "message":"成功",
    "data":null,
}
```
### 采集任务参数添加transforms匹配对应厂商设备为数据源打标签字段_Api
    "transforms": [{
        "type":"vendor",
        "vendorInfo":[{
            "eqpt_ip": "127.0.0.1",
            "vendor": "1",
            "dev_type": "1",
            "product": "1",
            "eqpt_name": "1",
            "eqpt_device_type": "1"
        },
            {
            "eqpt_ip": "127.0.0.2",
            "vendor": "2",
            "dev_type": "2",
            "product": "2",
            "eqpt_name": "2",
            "eqpt_device_type": "2"
        }]
    }


### 创建采集任务_hbase_接口Api
1. 请求地址:/datacollector/addrunner
2. 请求方式:Post
3. 请求参数方式:body-json
4. 请求参数:
```
{
    "name": "Hbase_01",
    "ip": ["172.20.27.999"],
    "maxprocs": 8,
    "max_colldata_size": 4194304,
    "reader": {
        "type": "hbase",
        "encoding": "utf-8",
        "Hbase_collection_type": 1,                                    //0静态表 1动态表
        "Hbase_host": "172.20.27.126,172.20.27.127,172.20.27.128",     //连接地址
        "Hbase_limit_batch": 1000,
        "Hbase_batch_intervel": "0 */1 * * * ?",                       //定时扫描
        "Hbase_exec_onstart": true,
        "Hbase_tables": [
            "student_"                                                 //表名
        ],
        "Hbase_dateUnit": "day",
        "Hbase_dateGap": 1,
        "Hbase_datetimeFormat": "YYYYMMDD",
        "Hbase_startDate": "20220110"
    },
    "parser": {
        "type": "raw",
        "idss_collect_time": true,
        "idss_collect_ip": true,
        "eqpt_ip": true
    },
    "senders": [
        {
            "type": "kafka",
            "kafka_host": [
                "172.20.27.126:9092",
                "172.20.27.127:9092",
                "172.20.27.128:9092"
            ],
            "kafka_topic": "ETL-IN-CX202108051412338611423165008711389184",
            "kafka_client_id": "logangent",
            "kafka_compression": "gzip",
            "gzip_compression_level": "最快压缩速度"
        }
    ],
    "sign": "041c9a5fb1b3f3cf062a0116dd680215"
}
``` 
5. 输出参数
```
{
    "code":0,                                                   //非 0 表示请求错误
    "message":"成功",
    "data":null,
}
```

### 采集任务参数添加transforms匹配对应厂商设备为数据源打标签字段_Api
    "transforms": [{
        "type":"vendor",
        "vendorInfo":[{
            "eqpt_ip": "127.0.0.1",
            "vendor": "1",
            "dev_type": "1",
            "product": "1",
            "eqpt_name": "1",
            "eqpt_device_type": "1"
        },
            {
            "eqpt_ip": "127.0.0.2",
            "vendor": "2",
            "dev_type": "2",
            "product": "2",
            "eqpt_name": "2",
            "eqpt_device_type": "2"
        }]
    }

### 创建采集任务_doris静态表_接口Api
1. 请求地址:/datacollector/addrunner
2. 请求方式:Post
3. 请求参数方式:body-json
4. 请求参数:
```
{
    "name": "doris_001",
    "ip": ["172.20.27.145"], 
    “maxprocs”:4,                                                                           
    "max_colldata_size": 4194304,
    "reader": {                                                                       
        "type": "doris",
        "encoding":"utf-8",                                             //采集解析编码 必填 string 默认 utf-8
        "autoClose":false,                                            //是否自动关闭 自动关闭 -> 全量采集
        "Doris_collection_type": 0,                                     //0 表示静态表，1 表示动态表
        "Doris_addr": "root:@tcp(172.20.27.148:9030)",
        "Doris_db_name": "doris",
        "Doris_sql": "Select * From `$TABLE$`",    
        "Doris_offset_key": "INCREMENTAL",                              //自增主键字段，offset记录依据，默认INCREMENTAL
        "Doris_limit_batch": 100,                                       //每次查询多少条数据
        "Doris_batch_intervel": "0 */1 * * * ?",                        //扫描间隔，cron定时任务查询，默认 1m 
        "Doris_exec_onstart": true,                                     //是否启动时查询一次
        "Doris_tables": [
            "test"
        ],
        "Doris_dateUnit": "day",                               //year 年表 month 月表 day日表
        "Doris_dateGap": 1,                                    //动态表 的是假间隔 默认 1 (几年一表 几月一表 几天一表)
        "Doris_datetimeFormat": "YYYY-MM-DD",
        "Doris_startDate": "2021-08-01"
    },
    "parser": {
        "type": "raw",
        "idss_collect_time": true,
        "idss_source_ip": true
    },
    "senders": [
        {
            "type": "doris",                                    //输出到doris
            "Doris_sendertype": 0,                              //0 采集模式 1同步模式
            "Doris_ip": "172.20.27.148",                        //doris ip
            "Doris_sql_port": 9030,                             //doris sql 端口
            "Doris_stream_port": 8040,                          //doris stream 上传端口
            "Doris_user": "root",                               //doris 用户名
            "Doris_password":"",                                //doris 密码
            "Doris_dbname":"doris",                             //doris 数据库名
            "Doris_tableName":"collect",                        //doris 数表名
            "Doris_mapping":{"Key1":"Key1","Key2":"Key2"}   //同步映射表
        }
    ],
    "sign": "54f2cbf8c903a843abc6a7eb9e9010ce"
}
``` 
5. 输出参数
```
{
    "code":0,                                                   //非 0 表示请求错误
    "message":"成功",
    "data":null,
}
```

### 创建采集任务_doris动态表_接口Api
1. 请求地址:/datacollector/addrunner
2. 请求方式:Post
3. 请求参数方式:body-json
4. 请求参数:
```
{
    "name": "doris_002",
    "ip": ["172.20.27.145"], 
    “maxprocs”:4,                                                                           
    "max_colldata_size": 4194304,
    "reader": {                                                                       
        "type": "doris",
        "encoding":"utf-8",                                           //采集解析编码 必填 string 默认 utf-8
        "autoClose":false,                                            //是否自动关闭 自动关闭 -> 全量采集
        "Doris_collection_type": 1,                                   //0 表示静态表，1 表示动态表
        "Doris_addr": "root:@tcp(172.20.27.148:9030)",
        "Doris_db_name": "doris",
        "Doris_sql": "Select * From `$TABLE$`",
        "Doris_offset_key": "INCREMENTAL",                            //自增主键字段，offset记录依据，默认INCREMENTAL
        "Doris_limit_batch": 100,                                     //每次查询多少条数据
        "Doris_batch_intervel": "0 */1 * * * ?",                      //扫描间隔，cron定时任务查询，默认 1m
        "Doris_exec_onstart": true,                                   //是否启动时查询一次
        "Doris_tables": [
            "test_"                                                   //动态表前缀字段 匹配test_2021-08-01
        ],
        "Doris_dateUnit": "day",
        "Doris_dateGap": 1,
        "Doris_datetimeFormat": "YYYY-MM-DD",  //时间格式对应动态表名匹配：
                                               //支持：YYYY-MM-DD,YYYYMMDD,YYYY_MM_DD,YYYYMM,YYYY-MM,YYYY_MM,YYYY
        "Doris_startDate": "2021-08-01"
    },
    "parser": {
        "type": "raw",
        "idss_collect_time": true,
        "idss_source_ip": true
    },
    "senders": [
        {
            "type": "kafka",
            "kafka_host": [
                "172.20.27.126:9092",
                "172.20.27.127:9092",
                "172.20.27.128:9092"
            ],
            "kafka_topic": "ETL-IN-CX202108051412338611423165008711389184",
            "kafka_client_id": "logangent",
            "kafka_compression": "gzip",
            "gzip_compression_level": "最快压缩速度"
        }
    ],
    "router": {
        "router_key_name": "",
        "router_match_type": "",
        "router_default_sender": 0,
        "router_routes": null
    },
    "web_folder": true,
    "is_stopped": true,
    "sign":"2bafebb22559ff0cefb4ba40d09ebd55"
    "sign": "0753f3e2517c89310fab7a94663c9634"
}
``` 
5. 输出参数
```
{
    "code":0,                                                   //非 0 表示请求错误
    "message":"成功",
    "data":null,
}
```



### 创建采集任务_gmsm2_接口Api
1. 请求地址:/datacollector/addrunner
2. 请求方式:Post
3. 请求参数方式:body-json
4. 请求参数:
```
{
    "name": "gmsm2_001",
    "ip": ["172.20.27.145"], 
    “maxprocs”:4,                                                                           
    "max_colldata_size": 4194304,
    "reader": {                                                                       
        "type": "gmsm2",
        "encoding":"utf-8",
        "autoClose":false,                                  //是否自动关闭 自动关闭 -> 全量采集
        "xml_collectiontype":0,                              //0 本地采集 1 ftp采集 2 sftp采集
        "gmsm2_prikey":"",                                    //解密私钥
        "gmsm2_prikey_pwd":"",                                //解密私钥 密码                                
        "gmsm2_max_collection_size":10485760,                 //采集最大文件 10m 默认
        "gmsm2_max_collection_num":1000,                      //采集最大文件量 1000 默认
        "gmsm2_server_addr":"12.19.18.78",                    //ftp/sftp 服务地址
        "gmsm2_server_port":21,                               //ftp/sftp 服务端口
        "gmsm2_server_user":"",                               //ftp/sftp 服务账号
        "gmsm2_server_password":"",                           //ftp/sftp 服务密码
        "gmsm2_directory":"/Users/liwei1dao/work/temap",      //采集目录
        "gmsm2_interval":"0 */1 * * * ?",                     //定时采集
        "gmsm2_regularrules":"\\.(txt)$",                     //文件正则过滤 默认 \.(txt)$
        "gmsm2_exec_onstart":true                             //启动时扫描
    },
    "parser": {
        "type": "raw",
        "idss_collect_time": true,
        "idss_source_ip": true
    },
    "senders": [
        {
            "type": "kafka",
            "kafka_host": [
                "172.20.27.126:9092",
                "172.20.27.127:9092",
                "172.20.27.128:9092"
            ],
            "kafka_topic": "ETL-IN-CX202108051412338611423165008711389184",
            "kafka_client_id": "logangent",
            "kafka_compression": "gzip",
            "gzip_compression_level": "最快压缩速度"
        }
    ],
    "sign": "a55d121914925f7a12f0b1199f43d19e"
}
``` 
5. 输出参数
```
{
    "code":0,                                                   //非 0 表示请求错误
    "message":"成功",
    "data":null,
}
```
### 启动采集任务Api
1. 请求地址:/datacollector/startrunners
2. 请求方式:Post
3. 请求参数方式:body-json
4. 请求参数:
```
    "ids":[
      {"rid":"采集任务_001","iid":"采集任务_001_123123123"},
      {"rid":"采集任务_002","iid":"采集任务_002_123123123"}
    ],
    "reqtime":4233123,
    "sign":"16daf080ddc15133ebe243a67db88fde"                                                       //签名 字段 [reqtime]
```
5. 输出参数
```
{
    "code":0,                                                   //非 0 表示请求错误
    "message":"成功",
    "data":{                                                   //只有在 code = 0 时 有数据
            "succ":[“成功任务1"],
            "fail":[“失败任务1"],
            "errmsg":[“错误信息1"]
        }
    },
}
```




### 驱动采集任务Api
1. 请求地址:/datacollector/driverunners
2. 请求方式:Post
3. 请求参数方式:body-json
4. 请求参数:
```
{
    "ids":[
      {"rid":"采集任务_001","iid":"采集任务_001_123123123"},
      {"rid":"采集任务_002","iid":"采集任务_002_123123123"}
    ],
    "reqtime":4233123,
    "sign":"16daf080ddc15133ebe243a67db88fde"                                                       //签名 字段 [reqtime]
}
```
5. 输出参数
```
{
    "code":0,                                                   //非 0 表示请求错误
    "message":"成功",
    "data":{                                                   //只有在 code = 0 时 有数据
            "succ":[“成功任务1"],
            "fail":[“失败任务1"],
            "errmsg":[“错误信息1"]
        }
    },
}
```

### 停止采集任务Api
1. 请求地址:/datacollector/stoprunners
2. 请求方式:Post
3. 请求参数方式:body-json
4. 请求参数:
```
{
    "ids":[
      {"rid":"采集任务_001","iid":"采集任务_001_123123123"},
      {"rid":"采集任务_002","iid":"采集任务_002_123123123"}
    ],
    "reqtime":4233123,
    "sign":"16daf080ddc15133ebe243a67db88fde"                                                       //签名 字段 [reqtime]
}
```
5. 输出参数
```
{
    "code":0,                                                   //非 0 表示请求错误
    "message":"成功",
    "data":{                                                   //只有在 code = 0 时 有数据
            "succ":[“成功任务1"],
            "fail":[“失败任务1"],
            "errmsg":[“错误信息1"]
        }
    },
}
```

### 删除采集任务Api
1. 请求地址:/datacollector/delrunners
2. 请求方式:Post
3. 请求参数方式:body-json
4. 请求参数:
```
{
    "ids":[
      {"rid":"采集任务_001","iid":"采集任务_001_123123123"},
      {"rid":"采集任务_002","iid":"采集任务_002_123123123"}
    ],
    "reqtime":4233123,
    "sign":"16daf080ddc15133ebe243a67db88fde"                                                       //签名 字段 [reqtime]
}
```
5. 输出参数
```
{
    "code":0,                                                   //非 0 表示请求错误
    "message":"成功",
    "data":{                                                   //只有在 code = 0 时 有数据
            "succ":[“成功任务1"],
            "fail":[“失败任务1"],
            "errmsg":[“错误信息1"]
        }
    },
}
```