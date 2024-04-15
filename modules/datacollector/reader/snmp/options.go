package snmp

import (
	"fmt"
	"lego_datacollector/modules/datacollector/core"
	"strconv"
	"strings"
	"time"

	jsoniter "github.com/json-iterator/go"
	"github.com/mitchellh/mapstructure"
)

type (
	IOptions interface {
		core.IReaderOptions //继承基础配置
		GetSnmp_reader_name() string
		GetSnmp_agents() string
		GetSnmp_table_init_host() string
		GetSnmp_community() string
		GetSnmp_auth_protocol() string
		GetSnmp_auth_password() string
		GetSnmp_priv_protocol() string
		GetSnmp_priv_password() string
		GetSnmp_sec_level() string
		GetSnmp_sec_name() string
		GetSnmp_context_name() string
		GetVersion() int
		GetRetries() int
		GetMax_repetitions() int
		GetSnmp_engine_id() string
		GetEngine_boots() int
		GetEngine_time() int
		GetSnmp_fields() string
		GetSnmp_tables() string
		GetSnmp_time_out() string
		GetSnmp_interval() string
		GetFields() []Field
		GetTables() []Table
		GetAgents() []string
		GetTime_out() time.Duration
	}
	Options struct {
		core.ReaderOptions   //继承基础配置
		Snmp_reader_name     string
		Snmp_agents          string
		Snmp_table_init_host string
		Snmp_community       string
		Snmp_auth_protocol   string
		Snmp_auth_password   string
		Snmp_priv_protocol   string
		Snmp_priv_password   string
		Snmp_context_name    string
		Snmp_sec_level       string
		Snmp_sec_name        string
		Snmp_version         string
		Snmp_retries         string
		Snmp_max_repetitions string
		Snmp_fields          string
		Snmp_engine_id       string
		Snmp_engine_boots    string
		Snmp_engine_time     string
		Snmp_tables          string
		Snmp_time_out        string
		Snmp_interval        string
		Fields               []Field
		Tables               []Table
		Agents               []string
		Version              int
		Retries              int
		Max_repetitions      int
		Engine_boots         int
		Engine_time          int
		Time_out             time.Duration
		//Interval             time.Duration
	}
)

func newOptions(config map[string]interface{}) (opt IOptions, err error) {
	time_out, err := time.ParseDuration("5s")
	if err != nil {
		return nil, err
	}
	options := &Options{
		Snmp_reader_name:     "distributed_default_snmp_name",
		Agents:               []string{"127.0.0.1:161"},
		Snmp_table_init_host: "127.0.0.1",
		Snmp_community:       "public",
		Snmp_sec_level:       "noAuthNoPriv",
		Snmp_sec_name:        "user",
		Snmp_auth_protocol:   "NoAuth",
		Snmp_auth_password:   "",
		Snmp_priv_protocol:   "NoPriv",
		Snmp_priv_password:   "mypass",
		Snmp_fields:          "[]",
		Snmp_tables:          "[]",
		Max_repetitions:      50,
		Retries:              3,
		Version:              2,
		Time_out:             time_out,
	}
	if config != nil {
		if err = mapstructure.Decode(config, options); err == nil {
			err = mapstructure.Decode(config, &options.ReaderOptions)
		}
	}
	if len(options.Snmp_version) != 0 {
		if options.Version, err = strconv.Atoi(options.Snmp_version); err != nil {
			err = fmt.Errorf("Reader snmp newOptions err:Snmp_version%s", options.Snmp_version)
			return
		}
	}
	if len(options.Snmp_retries) != 0 {
		if options.Retries, err = strconv.Atoi(options.Snmp_retries); err != nil {
			err = fmt.Errorf("Reader snmp newOptions err:Snmp_retries%s", options.Snmp_retries)
			return
		}
	}
	if len(options.Snmp_max_repetitions) != 0 {
		if options.Max_repetitions, err = strconv.Atoi(options.Snmp_max_repetitions); err != nil {
			err = fmt.Errorf("Reader snmp newOptions err:Snmp_max_repetitions%s", options.Snmp_max_repetitions)
			return
		}
	}
	if len(options.Snmp_engine_boots) != 0 {
		if options.Engine_boots, err = strconv.Atoi(options.Snmp_engine_boots); err != nil {
			err = fmt.Errorf("Reader snmp newOptions err:Snmp_engine_boots%s", options.Snmp_engine_boots)
			return
		}
	}
	if len(options.Snmp_engine_time) != 0 {
		if options.Engine_time, err = strconv.Atoi(options.Snmp_engine_time); err != nil {
			err = fmt.Errorf("Reader snmp newOptions err:Snmp_engine_time%s", options.Snmp_engine_time)
			return
		}
	}
	if len(options.Snmp_time_out) != 0 {
		if options.Time_out, err = time.ParseDuration(options.Snmp_time_out); err != nil {
			err = fmt.Errorf("Reader snmp newOptions err:Snmp_time_out%s", options.Snmp_time_out)
			return
		}
	}
	if len(options.Snmp_agents) != 0 {
		options.Agents = GetStringList(options.Snmp_agents)
	}

	if err = jsoniter.Unmarshal([]byte(options.Snmp_tables), &options.Tables); err != nil {
		return nil, err
	}
	if err = jsoniter.Unmarshal([]byte(options.Snmp_fields), &options.Fields); err != nil {
		return nil, err
	}

	opt = options
	return
}
func (this *Options) GetSnmp_reader_name() string {
	return this.Snmp_reader_name
}
func (this *Options) GetSnmp_agents() string {
	return this.Snmp_agents
}
func (this *Options) GetAgents() []string {
	return this.Agents
}
func (this *Options) GetSnmp_table_init_host() string {
	return this.Snmp_table_init_host
}
func (this *Options) GetSnmp_community() string {
	return this.Snmp_community
}
func (this *Options) GetSnmp_auth_protocol() string {
	return this.Snmp_auth_protocol
}
func (this *Options) GetSnmp_auth_password() string {
	return this.Snmp_auth_password
}
func (this *Options) GetSnmp_priv_protocol() string {
	return this.Snmp_priv_protocol
}
func (this *Options) GetSnmp_priv_password() string {
	return this.Snmp_priv_password
}
func (this *Options) GetSnmp_context_name() string {
	return this.Snmp_context_name
}
func (this *Options) GetSnmp_sec_level() string {
	return this.Snmp_sec_level
}
func (this *Options) GetSnmp_sec_name() string {
	return this.Snmp_sec_name
}
func (this *Options) GetVersion() int {
	return this.Version
}
func (this *Options) GetRetries() int {
	return this.Retries
}
func (this *Options) GetMax_repetitions() int {
	return this.Max_repetitions
}
func (this *Options) GetSnmp_engine_id() string {
	return this.Snmp_engine_id
}
func (this *Options) GetEngine_boots() int {
	return this.Engine_boots
}
func (this *Options) GetEngine_time() int {
	return this.Engine_time
}
func (this *Options) GetSnmp_fields() string {
	return this.Snmp_fields
}
func (this *Options) GetSnmp_tables() string {
	return this.Snmp_tables
}
func (this *Options) GetSnmp_time_out() string {
	return this.Snmp_time_out
}
func (this *Options) GetSnmp_interval() string {
	return this.Snmp_interval
}
func (this *Options) GetFields() []Field {
	return this.Fields
}
func (this *Options) GetTables() []Table {
	return this.Tables
}
func (this *Options) GetTime_out() time.Duration {
	return this.Time_out
}

func GetStringList(value string) []string {
	v := strings.Split(value, ",")
	var newV []string
	for _, i := range v {
		trimI := strings.TrimSpace(i)
		if len(trimI) > 0 {
			newV = append(newV, trimI)
		}
	}
	return newV
}
