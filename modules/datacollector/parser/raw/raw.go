package raw

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"lego_datacollector/modules/datacollector/core"
	"lego_datacollector/modules/datacollector/parser"
	"reflect"
	"strconv"
	"strings"
	"unsafe"

	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

type Parser struct {
	parser.Parser
	options      IOptions
	keyRaw       string
	keyTimestamp string
}

func (this *Parser) Init(rer core.IRunner, per core.IParser, options core.IParserOptions) (err error) {
	this.Parser.Init(rer, per, options)
	this.options = options.(IOptions)
	this.keyRaw = core.KeyRaw
	this.keyTimestamp = core.KeyTimestamp
	if this.options.GetInternalKeyPrefix() != "" {
		this.keyRaw = this.options.GetInternalKeyPrefix() + core.KeyRaw
		this.keyTimestamp = this.options.GetInternalKeyPrefix() + core.KeyTimestamp
	}
	return
}

func (this *Parser) Parse(bucket core.ICollDataBucket) {
	var (
		valstr string
		valmap map[string]interface{}
		ok     bool
		value  interface{}
	)
	for _, v := range bucket.Items() {
		value = v.GetValue()
		if valstr, ok = value.(string); ok {
			if len(strings.TrimSpace(valstr)) <= 0 {
				v.SetError(errors.New("Data is Space"))
				continue
			}
		}
		///处理编码统一
		if this.Runner.Reader().GetEncoding() != core.Encoding_utf8 {
			if valstr, ok = value.(string); ok {
				value = this.TransEncoding(valstr, this.Runner.Reader().GetEncoding())
			} else if valmap, ok = value.(map[string]interface{}); ok {
				for k1, v1 := range valmap {
					if valstr, ok = v1.(string); ok {
						valmap[k1] = this.TransEncoding(valstr, this.Runner.Reader().GetEncoding())
					}
				}
				value = valmap
			}
		}
		v.GetData()[this.keyRaw] = value
		if this.options.GetKeyTimestamp() {
			v.GetData()[this.keyTimestamp] = v.GetTime()
		}
		if this.options.GetIdss_collect_ip() {
			v.GetData()[core.KeyIdsCollecip] = this.Runner.ServiceIP()
		}
		if this.options.GetEqpt_ip() {
			v.GetData()[core.KeyEqptip] = v.GetSource()
		}
		v.GetData()[core.KeyIdsCollecexpmatch] = true
	}
	this.Parser.Parse(bucket)
	return
}

//编码转换
func (this *Parser) TransEncoding(val string, encoding core.Encoding) string {
	switch encoding {
	case core.Encoding_utf8:
		return val
	case core.Encoding_unicode:
		return unicode2utf8(val)
	case core.Encoding_gbk:
		if result, err := gbk2utf8(val); err == nil {
			return result
		} else {
			this.Runner.Errorf("Parser TransEncoding val:%s err:%v", val, err)
			return val
		}
	default:
		return val
	}
}

// Unicode 转 UTF-8
func unicode2utf8(source string) (result string) {
	var res = []string{""}
	sUnicode := strings.Split(source, "\\u")
	var context = ""
	for _, v := range sUnicode {
		var additional = ""
		if len(v) < 1 {
			continue
		}
		if len(v) > 4 {
			rs := []rune(v)
			v = string(rs[:4])
			additional = string(rs[4:])
		}
		temp, err := strconv.ParseInt(v, 16, 32)
		if err != nil {
			context += v
		}
		context += fmt.Sprintf("%c", temp)
		context += additional
	}
	res = append(res, context)
	result = strings.Join(res, "")
	result = strings.Replace(result, "&lt;", "<", -1)
	result = strings.Replace(result, "&gt;", ">", -1)
	return
}

// GBK 转 UTF-8
func gbk2utf8(source string) (string, error) {
	reader := transform.NewReader(bytes.NewReader(string2bytes(source)), simplifiedchinese.GBK.NewDecoder())
	d, e := ioutil.ReadAll(reader)
	if e != nil {
		return source, e
	}
	return bytes2string(d), nil
}

func string2bytes(s string) []byte {
	sh := (*reflect.StringHeader)(unsafe.Pointer(&s))
	bh := reflect.SliceHeader{
		Data: sh.Data,
		Len:  sh.Len,
		Cap:  sh.Len,
	}
	return *(*[]byte)(unsafe.Pointer(&bh))
}

func bytes2string(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}
