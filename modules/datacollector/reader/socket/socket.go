package socket

import (
	"fmt"
	"lego_datacollector/modules/datacollector/core"
	"lego_datacollector/modules/datacollector/reader"
	"net"
	"os"
	"strings"
)

type Reader struct {
	reader.Reader
	options  IOptions //以接口对象传递参数 方便后期继承扩展
	netproto string
	sredaer  ISocketReader
}

func (this *Reader) GetOptions() IOptions {
	return this.options
}
func (this *Reader) GetNetproto() string {
	return this.netproto
}

func (this *Reader) Init(runner core.IRunner, reader core.IReader, meta core.IMetaerData, options core.IReaderOptions) (err error) {
	var (
		spl []string
		l   net.Listener
		pc  net.PacketConn
	)
	if err = this.Reader.Init(runner, reader, meta, options); err != nil {
		return
	}
	this.options = options.(IOptions)
	if spl = strings.SplitN(this.options.GetSocket_service_address(), "://", 2); len(spl) != 2 {
		err = fmt.Errorf("Reader socket invalid Socket_service_address: %s", this.options.GetSocket_service_address())
		return
	}

	if spl[0] == "unix" || spl[0] == "unixpacket" || spl[0] == "unixgram" {
		// 通过remove来检测套接字文件是否存在
		os.Remove(spl[1])
	}
	this.netproto = spl[0]
	switch spl[0] {
	case "tcp", "tcp4", "tcp6", "unix", "unixpacket":
		if l, err = net.Listen(spl[0], spl[1]); err != nil {
			return
		}
		this.sredaer = newStreamSocketReader(this, l)
	case "udp", "udp4", "udp6", "ip", "ip4", "ip6", "unixgram":
		if pc, err = net.ListenPacket(spl[0], spl[1]); err != nil {
			return
		}
		this.sredaer = newPacketSocketReader(this, pc)
	default:
		err = fmt.Errorf("unknown protocol '%s' in '%s'", spl[0], this.options.GetSocket_service_address())
		return
	}
	return
}
func (this *Reader) Start() (err error) {
	err = this.Reader.Start()
	go this.sredaer.listen()
	return
}

func (this *Reader) Close() (err error) {
	err = this.sredaer.close()
	err = this.Reader.Close()
	return
}
