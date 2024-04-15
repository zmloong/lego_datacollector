package ipfix

import (
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"lego_datacollector/modules/datacollector/reader/netflow/session"
)

func errInvalidVersion(v uint16) error {
	return fmt.Errorf("version %d is not a valid IPFIX message version", v)
}

func errProtocol(f string) error {
	return errors.New("protocol error: " + f)
}

func errTemplateNotFound(t uint16) error {
	return fmt.Errorf("template with id=%d not found", t)
}

// Read a single IPFIX message from the provided reader and decode all the sets.
func Read(r io.Reader, s session.Session, t *Translate) (*Message, error) {
	m := new(Message)

	if t == nil && s != nil {
		t = NewTranslate(s)
	}

	if err := m.Header.Unmarshal(r); err != nil {
		return nil, err
	}
	if int(m.Header.Length) < m.Header.Len() {
		return nil, io.ErrShortBuffer
	}
	if m.Header.Version != Version {
		return nil, errInvalidVersion(m.Header.Version)
	}

	return m, m.UnmarshalSets(r, s, t)
}

func Dump(m *Message) {
	fmt.Println("IPFIX message")
	for _, ds := range m.DataSets {
		fmt.Println("  data set")
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
						fmt.Printf("        %d.%d: %v\n", f.Translated.EnterpriseNumber, f.Translated.InformationElementID, f.Bytes)
					}
				} else {
					fmt.Printf("        %v\n", f.Bytes)
				}
			}
		}
	}
}
