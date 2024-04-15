package netflow9

import (
	"encoding/hex"
	"fmt"
	"io"
	"lego_datacollector/modules/datacollector/reader/netflow/session"
)

func errInvalidVersion(v uint16) error {
	return fmt.Errorf("version %d is not a valid NetFlow packet version", v)
}

func errProtocol(f string, v ...interface{}) error {
	return fmt.Errorf("protocol error: "+f, v...)
}

func errTemplateNotFound(t uint16) error {
	return fmt.Errorf("template with id=%d not found", t)
}

// Read a single Netflow packet from the provided reader and decode all the sets.
func Read(r io.Reader, s session.Session, t *Translate) (*Packet, error) {
	p := new(Packet)

	if t == nil && s != nil {
		t = NewTranslate(s)
	}

	if err := p.Header.Unmarshal(r); err != nil {
		return nil, err
	}
	if p.Header.Version != Version {
		return nil, errInvalidVersion(p.Header.Version)
	}
	if p.Header.Len() < 4 {
		return nil, io.ErrShortBuffer
	}
	if p.Header.Count == 0 {
		return p, nil
	}
	return p, p.UnmarshalFlowSets(r, s, t)
}

func Dump(p *Packet) {
	fmt.Println("NetFlow version 9 packet")
	for _, ds := range p.DataFlowSets {
		fmt.Printf("  data set template %d, length: %d\n", ds.Header.ID, ds.Header.Length)
		if ds.Records == nil {
			fmt.Printf("    %d raw bytes:\n", len(ds.Bytes))
			fmt.Println(hex.Dump(ds.Bytes))
			continue
		}
		fmt.Printf("    %d records:\n", len(ds.Records))
		for i, dr := range ds.Records {
			fmt.Printf("      record %d:\n", i)
			for _, f := range dr.Fields {
				if f.Translated != nil {
					if f.Translated.Name != "" {
						fmt.Printf("        %s: %v\n", f.Translated.Name, f.Translated.Value)
					} else {
						fmt.Printf("        %d: %v\n", f.Translated.Type, f.Bytes)
					}
				} else {
					fmt.Printf("        %d: %v (raw)\n", f.Type, f.Bytes)
				}
			}
		}
	}
}
