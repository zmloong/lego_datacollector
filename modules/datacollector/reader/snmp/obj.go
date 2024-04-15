package snmp

import (
	"os/exec"
	"sync"

	"github.com/gosnmp/gosnmp"
)

type snmpConnection interface {
	Host() string
	Walk(string, gosnmp.WalkFunc) error
	Get(oids []string) (*gosnmp.SnmpPacket, error)
	Close()
}
type gosnmpWrapper struct {
	*gosnmp.GoSNMP
}
type Table struct {
	Name        string   `json:"table_name"`
	InheritTags []string `json:"table_inherit_tags"`
	IndexAsTag  bool     `json:"table_index_tag"`
	Fields      []Field  `json:"table_fields"`
	Oid         string   `json:"table_oid"`
}
type RTable struct {
	Name string
	Time string
	Rows []RTableRow
}

type RTableRow struct {
	Tags   map[string]string
	Fields map[string]interface{}
}

type Field struct {
	Name           string `json:"field_name"`
	Oid            string `json:"field_oid"`
	OidIndexSuffix string `json:"field_oid_index_suffix"`
	IsTag          bool   `json:"field_is_tag"`
	Conversion     string `json:"field_conversion"`
}

type snmpTableCache struct {
	mibName string
	oidNum  string
	oidText string
	fields  []Field
	err     error
}

type snmpTranslateCache struct {
	mibName    string
	oidNum     string
	oidText    string
	conversion string
	err        error
}

var snmpTableCaches map[string]snmpTableCache
var snmpTableCachesLock sync.Mutex
var snmpTranslateCachesLock sync.Mutex
var snmpTranslateCaches map[string]snmpTranslateCache
var execCommand = exec.Command
