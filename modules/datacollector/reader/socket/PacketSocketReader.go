package socket

import (
	"lego_datacollector/modules/datacollector/core"
	"net"
	"strings"
	"sync"
)

func newPacketSocketReader(rder IReader, c net.PacketConn) (reader *PacketSocketReader) {
	reader = &PacketSocketReader{
		reader:     rder,
		PacketConn: c,
		mesgdata:   make(chan *SocketMessage, 10),
		wg:         new(sync.WaitGroup),
	}
	for i := 0; i < rder.GetRunner().MaxProcs(); i++ {
		go reader.handlemesgdata()
	}
	return
}

type PacketSocketReader struct {
	reader     IReader
	PacketConn net.PacketConn
	mesgdata   chan *SocketMessage
	wg         *sync.WaitGroup
}

func (this *PacketSocketReader) listen() {
	var (
		n    int
		addr net.Addr
		err  error
		buf  = make([]byte, this.reader.GetOptions().GetSocket_read_buffer_size())
	)
	this.wg.Add(1)
	this.reader.GetRunner().Debugf("Reader socket udp Accept run")
locp:
	for {
		if n, addr, err = this.PacketConn.ReadFrom(buf); err != nil {
			this.reader.GetRunner().Debugf("Reader socket PacketSocketReader err:%v", err)
			this.reader.GetRunner().Close(core.Runner_Runing, err.Error())
			break locp
		}
		this.mesgdata <- &SocketMessage{
			Address: addr.String(),
			Message: string(buf[:n]),
		}
	}
	this.reader.GetRunner().Debugf("Reader socket udp Accept exit")
	this.wg.Done()
}

func (this *PacketSocketReader) close() (err error) {
	err = this.PacketConn.Close()
	this.wg.Wait()
	return
}

func (this *PacketSocketReader) handlemesgdata() {
	for v := range this.mesgdata {
		if this.reader.GetOptions().GetSocket_split_by_line() || this.reader.GetOptions().GetSocket_rule() == SocketRuleLine {
			vals := strings.Split(v.Message, "\n")
			for _, value := range vals {
				if value = strings.TrimSpace(value); value != "" {
					this.reader.Input() <- core.NewCollData(strings.Split(v.Address, ":")[0], value)
				}
			}
		} else {
			this.reader.Input() <- core.NewCollData(strings.Split(v.Address, ":")[0], v.Message)
		}
	}
}
