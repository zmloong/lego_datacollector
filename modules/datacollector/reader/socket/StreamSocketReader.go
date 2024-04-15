package socket

import (
	"bufio"
	"lego_datacollector/modules/datacollector/core"
	"net"
	"strings"
	"sync"
	"time"
)

func newStreamSocketReader(rder IReader, l net.Listener) (reader *StreamSocketReader) {
	reader = &StreamSocketReader{
		reader:   rder,
		Listener: l,
		mesgdata: make(chan *SocketMessage, 10),
		wg:       new(sync.WaitGroup),
	}
	for i := 0; i < rder.GetRunner().MaxProcs(); i++ {
		go reader.handlemesgdata()
	}
	return
}

type StreamSocketReader struct {
	reader   IReader
	Listener net.Listener
	mesgdata chan *SocketMessage
	wg       *sync.WaitGroup
}

func (this *StreamSocketReader) listen() {
	var (
		c   net.Conn
		err error
	)
	this.wg.Add(1)
	this.reader.GetRunner().Debugf("Reader socket tcp Accept run")
loop:
	for {
		if c, err = this.Listener.Accept(); err != nil {
			if strings.Contains(err.Error(), "too many open files") { //堵住了 直接抛弃掉
				continue
			} else {
				this.reader.GetRunner().Errorf("Reader socket StreamSocketReader err:%v", err)
				break loop
			}
		}
		this.wg.Add(1)
		go this.read(c)
	}
	this.reader.GetRunner().Debugf("Reader socket tcp Accept exit")
	this.wg.Done()
}

func (this *StreamSocketReader) close() (err error) {
	err = this.Listener.Close()
	this.wg.Wait()
	close(this.mesgdata)
	this.reader.GetRunner().Debugf("Reader socket tcp exit")
	return
}

func (this *StreamSocketReader) read(c net.Conn) {
	defer c.Close()
	defer this.wg.Done()
	scnr := bufio.NewScanner(c)
	var buffdata []byte
	for {
		if this.reader.GetOptions().GetSocket_read_timeout() > 0 {
			c.SetReadDeadline(time.Now().Add(time.Duration(this.reader.GetOptions().GetSocket_read_timeout()) * time.Second))
		}
		if !scnr.Scan() {
			if scnr.Err() != nil {
				this.reader.GetRunner().Errorf("StreamSocketReader [%s] err:%v", c.RemoteAddr().String(), scnr.Err())
			}
			break
		}
		c.Write([]byte{1})
		buffdata = scnr.Bytes()
		this.mesgdata <- &SocketMessage{
			Address: c.RemoteAddr().String(),
			Message: string(buffdata),
		}
	}
	return
}

func (this *StreamSocketReader) handlemesgdata() {
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
