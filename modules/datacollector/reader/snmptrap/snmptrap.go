package snmptrap

import (
	"fmt"
	"lego_datacollector/modules/datacollector/core"
	"lego_datacollector/modules/datacollector/reader"
	"net"
	"strings"
	"sync"

	"github.com/gosnmp/gosnmp"
)

type Reader struct {
	reader.Reader
	options    IOptions
	lock       sync.RWMutex
	TrapServer *gosnmp.TrapListener
	wg         *sync.WaitGroup
}

func (this *Reader) Init(runner core.IRunner, reader core.IReader, meta core.IMetaerData, options core.IReaderOptions) (err error) {
	if err = this.Reader.Init(runner, reader, meta, options); err != nil {
		return
	}
	this.options = options.(IOptions)
	this.wg = new(sync.WaitGroup)
	return
}
func (this *Reader) Start() (err error) {
	err = this.Reader.Start()
	spl := strings.SplitN(this.options.GetSnmptrap_address(), "://", 2)
	if len(spl) != 2 {
		return fmt.Errorf("SNMPTRAP invalid service address: %s", this.options.GetSnmptrap_address())
	}
	tl := gosnmp.NewTrapListener()
	tl.OnNewTrap = this.readTrap
	tl.Params = gosnmp.Default
	this.TrapServer = tl
	go this.runlisten()
	//err = this.Start()
	return
}
func (this *Reader) Close() (err error) {
	if this.TrapServer != nil {
		this.TrapServer.Close()
	}
	this.wg.Wait()
	err = this.Reader.Close()
	return
}
func (this *Reader) runlisten() {
	err := this.TrapServer.Listen(this.options.GetSnmptrap_address())
	if err != nil {
		this.GetRunner().Errorf("ServiceAddress：%s Error iTn RAP listen: %s\n", this.options.GetSnmptrap_address(), err)
		return
	}
}
func (this *Reader) readTrap(packet *gosnmp.SnmpPacket, addr *net.UDPAddr) {

	// log.Printf("SNMP trap received from: %s:%d. Community:%s, SnmpVersion:%s\n",
	// 	addr.IP, addr.Port, packet.Community, packet.Version)
	data := make(map[string]interface{})
	data["Address"] = fmt.Sprintf("%s:%d", addr.IP, addr.Port)
	data["Community"] = packet.Community
	data["SnmpVersion"] = packet.Version
	for _, variable := range packet.Variables {
		var val, tp string
		switch variable.Type {
		case gosnmp.OctetString:
			val = string(variable.Value.([]byte))
			tp = fmt.Sprintf("%s", variable.Type)

		case gosnmp.ObjectIdentifier:
			val = fmt.Sprintf("%s", variable.Value)
			tp = fmt.Sprintf("%s", variable.Type)

		case gosnmp.TimeTicks:
			a := gosnmp.ToBigInt(variable.Value)
			val = fmt.Sprintf("%d", (*a).Int64())
			tp = fmt.Sprintf("%s", variable.Type)
		case gosnmp.Integer:
			a := gosnmp.ToBigInt(variable.Value)
			val = fmt.Sprintf("%d", (*a).Int64())
			tp = fmt.Sprintf("%s", variable.Type)
		case gosnmp.Null:
			val = ""

		default:
			// ... or often you're just interested in numeric values.
			// ToBigInt() will return the Value as a BigInt, for plugging
			// into your calculations.
			a := gosnmp.ToBigInt(variable.Value)
			val = fmt.Sprintf("%d", (*a).Int64())
			tp = "UnknownType"
		}

		key := fmt.Sprintf("oid:%s(%s)", variable.Name, tp)
		data[key] = val
	}
	address := fmt.Sprintf("%s", addr.IP)
	this.Reader.Input() <- core.NewCollData(address, data)
}
