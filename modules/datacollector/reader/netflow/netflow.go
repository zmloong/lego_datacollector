package netflow

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"lego_datacollector/modules/datacollector/core"
	"lego_datacollector/modules/datacollector/reader"
	"lego_datacollector/modules/datacollector/reader/netflow/session"

	"net"
)

type Reader struct {
	reader.Reader
	options  IOptions //以接口对象传递参数 方便后期继承扩展
	decoders map[string]*Decoder
	conn     *net.UDPConn
}

func (this *Reader) Init(runner core.IRunner, reader core.IReader, meta core.IMetaerData, options core.IReaderOptions) (err error) {
	if err = this.Reader.Init(runner, reader, meta, options); err != nil {
		return
	}
	this.options = options.(IOptions)
	this.decoders = make(map[string]*Decoder)
	var addr *net.UDPAddr
	if addr, err = net.ResolveUDPAddr("udp", fmt.Sprintf("0.0.0.0:%d", this.options.GetNetflow_listen_port())); err != nil {
		return
	}
	if this.conn, err = net.ListenUDP("udp", addr); err != nil {
		return
	}
	return
}
func (this *Reader) Start() (err error) {
	err = this.Reader.Start()
	go this.run()
	return
}

func (this *Reader) Close() (err error) {
	this.conn.Close()
	err = this.Reader.Close()
	return
}

func (this *Reader) run() {
	var (
		buff      = make([]byte, this.options.GetNetflow_read_size())
		size      int
		raddr     *net.UDPAddr
		m         Message
		err       error
		databytes []byte
	)
loop:
	for {
		if size, raddr, err = this.conn.ReadFromUDP(buff[:]); err == nil || err == io.EOF {
			this.Runner.Debugf("reader netflow addr：%s message%v", raddr.String(), buff[:size])
			d, found := this.decoders[raddr.String()]
			if !found {
				s := session.New()
				d = NewDecoder(s)
				this.decoders[raddr.String()] = d
			}
			m, err = d.Read(bytes.NewBuffer(buff[:size]))
			if err != nil {
				this.Runner.Errorf("reader netflow decoder error:", err)
				continue
			}
			if databytes, err = json.Marshal(m); err != nil {
				this.Runner.Errorf("reader netflow jsonMarshal error:", err)
				continue
			}
			this.Runner.Debugf("reader netflow output:%s", string(databytes))
			this.Input() <- core.NewCollData(raddr.IP.String(), string(databytes))
		} else {
			break loop
		}
	}
	this.Runner.Errorf("Reader run exit err:%v", err)
}
