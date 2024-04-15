package canal

import (
	"encoding/json"
	"lego_datacollector/modules/datacollector/core"
	"lego_datacollector/modules/datacollector/reader"
	"unsafe"

	"github.com/withlin/canal-go/protocol/entry"
)

func init() {
	reader.RegisterReader(ReaderType, NewReader)
}

const (
	ReaderType = "canal"
)

var EventTypeCode = map[entry.EventType]string{
	entry.EventType_EVENTTYPECOMPATIBLEPROTO2: "EVENTTYPECOMPATIBLEPROTO2",
	entry.EventType_INSERT:                    "INSERT",
	entry.EventType_UPDATE:                    "UPDATE",
	entry.EventType_DELETE:                    "DELETE",
	entry.EventType_CREATE:                    "CREATE",
	entry.EventType_ALTER:                     "ALTER",
	entry.EventType_ERASE:                     "ERASE",
	entry.EventType_QUERY:                     "QUERY",
	entry.EventType_TRUNCATE:                  "TRUNCATE",
	entry.EventType_RENAME:                    "RENAME",
	entry.EventType_CINDEX:                    "CINDEX",
	entry.EventType_DINDEX:                    "DINDEX",
	entry.EventType_GTID:                      "GTID",
	entry.EventType_XACOMMIT:                  "XACOMMIT",
	entry.EventType_XAROLLBACK:                "XAROLLBACK",
	entry.EventType_MHEARTBEAT:                "MHEARTBEAT",
}

type (
	ReaderData struct {
		LogfileName   string           `json:"logfilename"`
		LogfileOffset int64            `json:"logfileoffset"`
		SchemaName    string           `json:"schemaname"`
		TableName     string           `json:"tablename"`
		EventType     string           `json:"eventtype"`
		RowDatas      []*entry.RowData `json:"rowdatas"`
	}
)

func NewReader(runner core.IRunner, conf map[string]interface{}) (rder core.IReader, err error) {
	var (
		opt IOptions
		r   *Reader
	)
	if opt, err = newOptions(conf); err != nil {
		return
	}
	r = &Reader{}
	if err = r.Init(runner, r, nil, opt); err != nil {
		return
	}
	rder = r
	return
}

//编码到string
func (this *ReaderData) ToString() (value string, err error) {
	var buff []byte
	if buff, err = json.Marshal(this); err != nil {
		return
	}
	value = *(*string)(unsafe.Pointer(&buff)) //直接指针转换 避免不必要的内存拷贝
	return
}

func getEventType(etype entry.EventType) (code string) {
	if e, ok := EventTypeCode[etype]; ok {
		return e
	} else {
		return "未知事件"
	}
}
