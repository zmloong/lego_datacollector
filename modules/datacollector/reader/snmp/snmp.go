package snmp

import (
	"encoding/json"
	"fmt"
	"lego_datacollector/modules/datacollector/core"
	"lego_datacollector/modules/datacollector/reader"
	"net"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"

	lgcron "github.com/liwei1dao/lego/sys/cron"
	"github.com/robfig/cron/v3"

	"github.com/gosnmp/gosnmp"
)

type Reader struct {
	reader.Reader
	options         IOptions
	ConnectionCache []snmpConnection
	runId           cron.EntryID
	rungoroutine    int32
	wg              sync.WaitGroup //用于等待采集器完全关闭
}

func (this *Reader) Init(runner core.IRunner, reader core.IReader, meta core.IMetaerData, options core.IReaderOptions) (err error) {
	if err = this.Reader.Init(runner, reader, meta, options); err != nil {
		return
	}
	this.options = options.(IOptions)
	this.ConnectionCache = make([]snmpConnection, len(this.options.GetAgents()))
	var tables []Table
	var fields []Field
	tables = this.options.GetTables()
	fields = this.options.GetFields()
	if len(tables) == 0 && len(fields) == 0 {
		return fmt.Errorf("Reader Snmp Init 'snmp_tables' and 'snmp_fields' are both empty, must have one of them")
	}

	// for i := range tables {
	// 	if subErr := tables[i].init(this.options.GetSnmp_table_init_host()); subErr != nil {
	// 		return fmt.Errorf("Reader Snmp Init initializing table %s,err:%s", tables[i].Name, subErr)
	// 	}
	// }
	for i := range fields {
		if subErr := fields[i].init(); subErr != nil {
			return fmt.Errorf("Reader Snmp Init initializing field %s,err:%s", fields[i].Name, subErr)
		}
	}
	return
}

func (this *Reader) Start() (err error) {
	err = this.Reader.Start()
	if this.runId, err = lgcron.AddFunc(this.options.GetSnmp_interval(), this.Gather); err != nil {
		err = fmt.Errorf("lgcron.AddFunc corn:%s err:%v", this.options.GetSnmp_interval(), err)
		return
	}
	return nil
}

///外部调度器 驱动执行  此接口 不可阻塞
func (this *Reader) Drive() (err error) {
	err = this.Reader.Drive()
	go this.Gather()
	return
}
func (this *Reader) Close() (err error) {
	lgcron.Remove(this.runId)
	this.wg.Wait()
	err = this.Reader.Close()
	return nil
}

func (this *Reader) getConnection(idx int) (snmpConnection, error) {
	if gs := this.ConnectionCache[idx]; gs != nil {
		return gs, nil
	}

	agent := this.options.GetAgents()[idx]

	gs := gosnmpWrapper{&gosnmp.GoSNMP{}}
	this.ConnectionCache[idx] = gs

	host, portStr, err := net.SplitHostPort(agent)
	if err != nil {
		if err, ok := err.(*net.AddrError); !ok || err.Err != "missing port in address" {
			return nil, fmt.Errorf("Reader Snmp getConnection parsing host err:%s", err)
		}
		host = agent
		portStr = "161"
	}
	gs.Target = host

	port, err := strconv.ParseUint(portStr, 10, 16)
	if err != nil {
		return nil, fmt.Errorf("Reader Snmp getConnection parsing port err:%s", err)
	}
	gs.Port = uint16(port)

	gs.Timeout = this.options.GetTime_out()

	gs.Retries = this.options.GetRetries()

	switch this.options.GetVersion() {
	case 3:
		gs.Version = gosnmp.Version3
	case 2, 0:
		gs.Version = gosnmp.Version2c
	case 1:
		gs.Version = gosnmp.Version1
	default:
		return nil, fmt.Errorf("Reader Snmp getConnection invalid version")
	}

	if this.options.GetVersion() < 3 {
		if this.options.GetSnmp_community() == "" {
			gs.Community = "public"
		} else {
			gs.Community = this.options.GetSnmp_community()
		}
	}

	gs.MaxRepetitions = uint32(this.options.GetMax_repetitions())

	if this.options.GetVersion() == 3 {
		gs.ContextName = this.options.GetSnmp_context_name()

		sp := &gosnmp.UsmSecurityParameters{}
		gs.SecurityParameters = sp
		gs.SecurityModel = gosnmp.UserSecurityModel

		switch strings.ToLower(this.options.GetSnmp_sec_level()) {
		case "noauthnopriv", "":
			gs.MsgFlags = gosnmp.NoAuthNoPriv
		case "authnopriv":
			gs.MsgFlags = gosnmp.AuthNoPriv
		case "authpriv":
			gs.MsgFlags = gosnmp.AuthPriv
		default:
			return nil, fmt.Errorf("Reader Snmp getConnection invalid secLevel")
		}

		sp.UserName = this.options.GetSnmp_sec_name()

		switch strings.ToLower(this.options.GetSnmp_auth_protocol()) {
		case "md5":
			sp.AuthenticationProtocol = gosnmp.MD5
		case "sha":
			sp.AuthenticationProtocol = gosnmp.SHA
		case "noauth", "":
			sp.AuthenticationProtocol = gosnmp.NoAuth
		default:
			return nil, fmt.Errorf("Reader Snmp getConnection invalid authProtocol")
		}

		sp.AuthenticationPassphrase = this.options.GetSnmp_auth_password()

		switch strings.ToLower(this.options.GetSnmp_priv_protocol()) {
		case "des":
			sp.PrivacyProtocol = gosnmp.DES
		case "aes":
			sp.PrivacyProtocol = gosnmp.AES
		case "nopriv", "":
			sp.PrivacyProtocol = gosnmp.NoPriv
		default:
			return nil, fmt.Errorf("Reader Snmp getConnection invalid privProtocol")
		}

		sp.PrivacyPassphrase = this.options.GetSnmp_priv_password()

		sp.AuthoritativeEngineID = this.options.GetSnmp_engine_id()

		sp.AuthoritativeEngineBoots = uint32(this.options.GetEngine_boots())

		sp.AuthoritativeEngineTime = uint32(this.options.GetEngine_time())
	}

	if err := gs.Connect(); err != nil {
		return nil, fmt.Errorf("Reader Snmp getConnection setting up connection err:%s", err)
	}

	return gs, nil
}
func (this *Reader) Gather() {
	var (
		rungoroutine = atomic.LoadInt32(&this.rungoroutine)
	)
	if rungoroutine != 0 {
		this.Runner.Debugf("Reader is collectioning rungoroutine:%d", rungoroutine)
		return
	}
	//var wg sync.WaitGroup

	for i, agent := range this.options.GetAgents() {
		this.wg.Add(1)
		atomic.AddInt32(&this.rungoroutine, int32(1))
		addr := strings.Split(agent, ":")[0]
		go this.AgentRun(i, addr)
	}
	return
}
func (this *Reader) AgentRun(i int, agent string) {
	this.GetRunner().Debugf("Reader %q is reading from agent: %s", this.GetType(), agent)
	defer this.wg.Done()
	defer atomic.AddInt32(&this.rungoroutine, int32(-1))
	datas := make([]map[string]interface{}, 0)
	conn, err := this.getConnection(i)
	defer conn.Close()
	if err != nil {
		return
	}
	t := Table{
		Name:   this.options.GetSnmp_reader_name(),
		Fields: this.options.GetFields(),
	}
	topTags := map[string]string{}
	if datas, err = this.gatherTable(conn, t, topTags, false); err != nil {
		this.GetRunner().Debugf("Reader AgentRun gatherTable err: %s", err)
		return
	}
	if err = this.StoreData(agent, datas); err != nil {
		this.GetRunner().Debugf("Reader AgentRun StoreData err: %s", err)
		return
	}
	for _, t := range this.options.GetTables() {
		if datas, err = this.gatherTable(conn, t, topTags, true); err != nil {
			this.GetRunner().Debugf("Reader gatherTable err: %s", err)
			return
		}
		if err = this.StoreData(agent, datas); err != nil {
			this.GetRunner().Debugf("Reader AgentRun StoreData err: %s", err)
			return
		}
	}
}
func (this *Reader) gatherTable(gs snmpConnection, t Table, topTags map[string]string, walk bool) ([]map[string]interface{}, error) {
	rt, err := t.Build(gs, walk)
	if err != nil {
		return nil, err
	}
	datas := make([]map[string]interface{}, 0, len(rt.Rows))
	for _, tr := range rt.Rows {
		if !walk {
			for k, v := range tr.Tags {
				topTags[k] = v
			}
		} else {
			for _, k := range t.InheritTags {
				if v, ok := topTags[k]; ok {
					tr.Tags[k] = v
				}
			}
		}
		if _, ok := tr.Tags["agent_host"]; !ok {
			tr.Tags["agent_host"] = gs.Host()
		}

		data := make(map[string]interface{})
		for k, v := range tr.Fields {
			data[k] = v
		}
		for k, v := range tr.Tags {
			data[k] = v
		}
		data[KeyTimestamp] = rt.Time
		data[KeySnmpTableName] = t.Name
		datas = append(datas, data)
	}
	return datas, nil
}
func (this *Reader) StoreData(address string, datas []map[string]interface{}) (err error) {
	this.GetRunner().Debugf("Reader is StoreData: %s", datas)
	if datas == nil || len(datas) == 0 {
		return
	}
	for _, data := range datas {
		if data == nil || len(data) == 0 {
			continue
		}
		databytes, err1 := json.Marshal(data)
		if err1 != nil {
			continue
		}
		this.Reader.Input() <- core.NewCollData(address, string(databytes))
	}
	return
}
