package snmp

import (
	"bufio"
	"bytes"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/gosnmp/gosnmp"
)

func (gsw gosnmpWrapper) Host() string {
	return gsw.Target
}

func (gsw gosnmpWrapper) Walk(oid string, fn gosnmp.WalkFunc) error {
	var err error
	// On error, retry once.
	// Unfortunately we can't distinguish between an error returned by gosnmp, and one returned by the walk function.
	for i := 0; i < 2; i++ {
		if gsw.Version == gosnmp.Version1 {
			err = gsw.GoSNMP.Walk(oid, fn)
		} else {
			err = gsw.GoSNMP.BulkWalk(oid, fn)
		}
		if err == nil {
			return nil
		}
		if err := gsw.GoSNMP.Connect(); err != nil {
			return fmt.Errorf("Reader gosnmpWrapper Walk Reconnecting err:%s", err)
		}
	}
	return err
}

func (gsw gosnmpWrapper) Get(oids []string) (*gosnmp.SnmpPacket, error) {
	var err error
	var pkt *gosnmp.SnmpPacket
	for i := 0; i < 2; i++ {
		pkt, err = gsw.GoSNMP.Get(oids)
		if err == nil {
			return pkt, nil
		}
		if err := gsw.GoSNMP.Connect(); err != nil {
			return nil, fmt.Errorf("Reader gosnmpWrapper Get Reconnecting err:%s", err)
		}
	}
	return nil, err
}
func (gsw gosnmpWrapper) Close() {
	gsw.GoSNMP.Conn.Close()
}
func (f *Field) init() error {
	_, oidNum, oidText, conversion, err := snmpTranslate(f.Oid)
	if err != nil {
		return fmt.Errorf("Field init translating err:%s", err)
	}
	f.Oid = oidNum
	if f.Name == "" {
		f.Name = oidText
	}
	if f.Conversion == "" {
		f.Conversion = conversion
	}

	//TODO use textual convention conversion from the MIB

	return nil
}
func (t *Table) init(host string) error {
	if err := t.initBuild(host); err != nil {
		return err
	}
	for i := range t.Fields {
		if err := t.Fields[i].init(); err != nil {
			return fmt.Errorf("Table init err:%s,initializing field %s", err, t.Fields[i].Name)
		}
	}
	return nil
}
func (t *Table) initBuild(host string) error {
	if t.Oid == "" {
		return nil
	}

	_, _, oidText, fields, err := snmpTable(t.Oid, host)
	if err != nil {
		return err
	}
	if t.Name == "" {
		t.Name = oidText
	}
	t.Fields = append(t.Fields, fields...)

	return nil
}
func (t Table) Build(gs snmpConnection, walk bool) (*RTable, error) {
	rows := map[string]RTableRow{}

	tagCount := 0
	for _, f := range t.Fields {
		if f.IsTag {
			tagCount++
		}

		if len(f.Oid) == 0 {
			return nil, fmt.Errorf("cannot have empty OID on field %s", f.Name)
		}
		var oid string
		if f.Oid[0] == '.' {
			oid = f.Oid
		} else {
			oid = "." + f.Oid
		}

		ifv := map[string]interface{}{}

		if !walk {
			if pkt, err := gs.Get([]string{oid}); err != nil {
				return nil, fmt.Errorf("err:%s,SNMP performing get on field %s", err, f.Name)
			} else if pkt != nil && len(pkt.Variables) > 0 && pkt.Variables[0].Type != gosnmp.NoSuchObject && pkt.Variables[0].Type != gosnmp.NoSuchInstance {
				ent := pkt.Variables[0]
				fv, err := fieldConvert(f.Conversion, ent.Value)
				if err != nil {
					return nil, fmt.Errorf("err:%s,SNMP converting %q (OID %s) for field %s", err, ent.Value, ent.Name, f.Name)
				}
				ifv[""] = fv
			}
		} else {
			err := gs.Walk(oid, func(ent gosnmp.SnmpPDU) error {
				if len(ent.Name) <= len(oid) || ent.Name[:len(oid)+1] != oid+"." {
					return NestedError{} // break the walk
				}
				idx := ent.Name[len(oid):]
				if f.OidIndexSuffix != "" {
					if !strings.HasSuffix(idx, f.OidIndexSuffix) {
						return nil
					}
					idx = idx[:len(idx)-len(f.OidIndexSuffix)]
				}
				fv, err := fieldConvert(f.Conversion, ent.Value)
				if err != nil {
					return fmt.Errorf("err:%s,SNMP converting %q (OID %s) for field %s", err, ent.Value, ent.Name, f.Name)
				}
				ifv[idx] = fv
				return nil
			})
			if err != nil {
				if _, ok := err.(NestedError); !ok {
					return nil, fmt.Errorf("err:%s,SNMP performing bulk walk for field %s", err, f.Name)
				}
			}
		}

		for idx, v := range ifv {
			rtr, ok := rows[idx]
			if !ok {
				rtr = RTableRow{}
				rtr.Tags = map[string]string{}
				rtr.Fields = map[string]interface{}{}
				rows[idx] = rtr
			}
			if t.IndexAsTag && idx != "" {
				if idx[0] == '.' {
					idx = idx[1:]
				}
				rtr.Tags["index"] = idx
			}
			if vs, ok := v.(string); !ok || vs != "" {
				if f.IsTag {
					if ok {
						rtr.Tags[f.Name] = vs
					} else {
						rtr.Tags[f.Name] = fmt.Sprintf("%v", v)
					}
				} else {
					rtr.Fields[f.Name] = v
				}
			}
		}
	}

	rt := RTable{
		Name: t.Name,
		Rows: make([]RTableRow, 0, len(rows)),
		Time: time.Now().Format(time.RFC3339Nano),
	}
	for _, r := range rows {
		rt.Rows = append(rt.Rows, r)
	}
	return &rt, nil
}

func snmpTable(oid, host string) (mibName string, oidNum string, oidText string, fields []Field, err error) {
	snmpTableCachesLock.Lock()
	if snmpTableCaches == nil {
		snmpTableCaches = map[string]snmpTableCache{}
	}

	var stc snmpTableCache
	var ok bool
	if stc, ok = snmpTableCaches[oid]; !ok {
		stc.mibName, stc.oidNum, stc.oidText, stc.fields, stc.err = snmpTableCall(oid, host)
		snmpTableCaches[oid] = stc
	}

	snmpTableCachesLock.Unlock()
	return stc.mibName, stc.oidNum, stc.oidText, stc.fields, stc.err
}

func snmpTableCall(oid, host string) (mibName string, oidNum string, oidText string, fields []Field, err error) {
	mibName, oidNum, oidText, _, err = snmpTranslate(oid)
	if err != nil {
		return "", "", "", nil, fmt.Errorf("snmpTableCall translating err：%s", err)
	}

	mibPrefix := mibName + "::"
	oidFullName := mibPrefix + oidText

	tagOids := map[string]struct{}{}
	if out, err := execCmd("snmptranslate", "-Td", oidFullName+".1"); err == nil {
		scanner := bufio.NewScanner(bytes.NewBuffer(out))
		for scanner.Scan() {
			line := scanner.Text()

			if !strings.HasPrefix(line, "  INDEX") {
				continue
			}

			i := strings.Index(line, "{ ")
			if i == -1 { // parse error
				continue
			}
			line = line[i+2:]
			i = strings.Index(line, " }")
			if i == -1 { // parse error
				continue
			}
			line = line[:i]
			for _, col := range strings.Split(line, ", ") {
				tagOids[mibPrefix+col] = struct{}{}
			}
		}
	}

	out, err := execCmd("snmptable", "-Ch", "-Cl", "-c", "public", host, oidFullName)
	if err != nil {
		return "", "", "", nil, fmt.Errorf("snmpTableCall execCmd：snmptable... getting table columns err：%s", err)
	}
	scanner := bufio.NewScanner(bytes.NewBuffer(out))
	scanner.Scan()
	cols := scanner.Text()
	if len(cols) == 0 {
		return "", "", "", nil, fmt.Errorf("could not find any columns in table")
	}
	for _, col := range strings.Split(cols, " ") {
		if len(col) == 0 {
			continue
		}
		_, isTag := tagOids[mibPrefix+col]
		fields = append(fields, Field{Name: col, Oid: mibPrefix + col, IsTag: isTag})
	}

	return mibName, oidNum, oidText, fields, err
}

func snmpTranslate(oid string) (mibName string, oidNum string, oidText string, conversion string, err error) {
	snmpTranslateCachesLock.Lock()
	if snmpTranslateCaches == nil {
		snmpTranslateCaches = map[string]snmpTranslateCache{}
	}

	var stc snmpTranslateCache
	var ok bool
	if stc, ok = snmpTranslateCaches[oid]; !ok {

		stc.mibName, stc.oidNum, stc.oidText, stc.conversion, stc.err = snmpTranslateCall(oid)
		snmpTranslateCaches[oid] = stc
	}

	snmpTranslateCachesLock.Unlock()

	return stc.mibName, stc.oidNum, stc.oidText, stc.conversion, stc.err
}

func snmpTranslateCall(oid string) (mibName string, oidNum string, oidText string, conversion string, err error) {
	var out []byte
	if strings.ContainsAny(oid, ":abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ") {
		out, err = execCmd("snmptranslate", "-Td", "-Ob", oid)
	} else {
		out, err = execCmd("snmptranslate", "-Td", "-Ob", "-m", "all", oid)
		if err, ok := err.(*exec.Error); ok && err.Err == exec.ErrNotFound {
			return "", oid, oid, "", nil
		}
	}
	if err != nil {
		return "", "", "", "", err
	}

	scanner := bufio.NewScanner(bytes.NewBuffer(out))
	ok := scanner.Scan()
	if !ok && scanner.Err() != nil {
		return "", "", "", "", fmt.Errorf("snmpTranslateCall execCmd:snmptranslate... getting OID text err:%s", scanner.Err())
	}

	oidText = scanner.Text()

	i := strings.Index(oidText, "::")
	if i == -1 {
		if bytes.Contains(out, []byte("[TRUNCATED]")) {
			return "", oid, oid, "", nil
		}
		oidText = oid
	} else {
		mibName = oidText[:i]
		oidText = oidText[i+2:]
	}

	for scanner.Scan() {
		line := scanner.Text()

		if strings.HasPrefix(line, "  -- TEXTUAL CONVENTION ") {
			tc := strings.TrimPrefix(line, "  -- TEXTUAL CONVENTION ")
			switch tc {
			case "MacAddress", "PhysAddress":
				conversion = "hwaddr"
			case "InetAddressIPv4", "InetAddressIPv6", "InetAddress":
				conversion = "ipaddr"
			}
		} else if strings.HasPrefix(line, "::= { ") {
			objs := strings.TrimPrefix(line, "::= { ")
			objs = strings.TrimSuffix(objs, " }")

			for _, obj := range strings.Split(objs, " ") {
				if len(obj) == 0 {
					continue
				}
				if i := strings.Index(obj, "("); i != -1 {
					obj = obj[i+1:]
					oidNum += "." + obj[:strings.Index(obj, ")")]
				} else {
					oidNum += "." + obj
				}
			}
			break
		}
	}
	return mibName, oidNum, oidText, conversion, nil
}

func execCmd(arg0 string, args ...string) ([]byte, error) {
	out, err := execCommand(arg0, args...).Output()
	if err != nil {
		if err, ok := err.(*exec.ExitError); ok {
			return nil, NestedError{
				Err:       err,
				NestedErr: fmt.Errorf("%s", bytes.TrimRight(err.Stderr, "\r\n")),
			}
		}
		return nil, err
	}
	return out, nil
}
