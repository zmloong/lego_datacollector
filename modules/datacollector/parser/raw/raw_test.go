package raw_test

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"unicode"
	"unsafe"

	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

func Test_canal(t *testing.T) {
	str := unicode2utf8("\u002f\u0026\u006c\u0074\u003b\u0031\u0036\u0035\u002f\u0026\u0067\u0074\u003b\u0053\u0065\u0070\u0020\u0031\u0038\u0020\u0031\u0036\u003a\u0035\u0038\u003a\u0033\u0030\u0020\u0063\u006e\u0067\u0062\u002d\u006c\u006f\u0067\u0069\u006e\u002d\u0062\u0031\u0033\u002d\u0033\u0020\u0041\u0075\u0064\u0069\u0074\u0064\u0020\u0074\u0079\u0070\u0065\u003d\u0053\u0059\u0053\u0043\u0041\u004c\u004c\u0020\u006d\u0073\u0067\u003d\u0061\u0075\u0064\u0069\u0074\u0028\u0031\u0035\u0030\u0035\u0037\u0032\u0035\u0031\u0030\u0038\u002e\u0038\u0030\u0032\u003a\u0031\u0035\u0033\u0039\u0036\u0036\u0038\u0035\u0029\u003a\u0020\u0061\u0072\u0063\u0068\u003d\u0063\u0030\u0030\u0030\u0030\u0030\u0033\u0065\u0020\u0073\u0079\u0073\u0063\u0061\u006c\u006c\u003d\u0035\u0039\u0020\u0073\u0075\u0063\u0063\u0065\u0073\u0073\u003d\u0079\u0065\u0073\u0020\u0065\u0078\u0069\u0074\u003d\u0030\u0020\u0061\u0030\u003d\u0031\u0066\u0031\u0031\u0062\u0062\u0030\u0020\u0061\u0031\u003d\u0031\u0066\u0030\u0061\u0038\u0030\u0030\u0020\u0061\u0032\u003d\u0031\u0066\u0030\u0036\u0063\u0061\u0030\u0020\u0061\u0033\u003d\u0031\u0030\u0020\u0069\u0074\u0065\u006d\u0073\u003d\u0032\u0020\u0070\u0070\u0069\u0064\u003d\u0032\u0035\u0039\u0033\u0036\u0020\u0070\u0069\u0064\u003d\u0032\u0035\u0039\u0033\u0037\u0020\u0061\u0075\u0069\u0064\u003d\u0033\u0036\u0033\u0037\u0034\u0020\u0075\u0069\u0064\u003d\u0033\u0036\u0033\u0037\u0034\u0020\u0067\u0069\u0064\u003d\u0033\u0030\u0030\u0034\u0020\u0065\u0075\u0069\u0064\u003d\u0033\u0036\u0033\u0037\u0034\u0020\u0073\u0075\u0069\u0064\u003d\u0033\u0036\u0033\u0037\u0034\u0020\u0066\u0073\u0075\u0069\u0064\u003d\u0033\u0036\u0033\u0037\u0034\u0020\u0065\u0067\u0069\u0064\u003d\u0033\u0030\u0030\u0034\u0020\u0073\u0067\u0069\u0064\u003d\u0033\u0030\u0030\u0034\u0020\u0066\u0073\u0067\u0069\u0064\u003d\u0033\u0030\u0030\u0034\u0020\u0074\u0074\u0079\u003d\u0028\u006e\u006f\u006e\u0065\u0029\u0020\u0073\u0065\u0073\u003d\u0036\u0039\u0038\u0031\u0033\u0020\u0063\u006f\u006d\u006d\u003d\u0026\u0071\u0075\u006f\u0074\u003b\u0071\u0073\u0074\u0061\u0074\u0026\u0071\u0075\u006f\u0074\u003b\u0020\u0069\u0064\u0073\u0073\u005f\u0073\u006f\u0075\u0072\u0063\u0065\u005f\u0069\u0070\u003d\u0031\u0030\u0020\u0065\u0078\u0065\u003d\u0026\u0071\u0075\u006f\u0074\u003b\u002f\u006f\u0070\u0074\u002f\u0067\u0072\u0069\u0064\u0065\u006e\u0067\u0069\u006e\u0065\u002f\u0062\u0069\u006e\u002f\u006c\u0069\u006e\u0075\u0078\u002d\u0078\u0036\u0034\u002f\u0071\u0073\u0074\u0061\u0074\u0026\u0071\u0075\u006f\u0074\u003b\u0020\u006b\u0065\u0079\u003d\u0026\u0071\u0075\u006f\u0074\u003b\u0065\u0078\u0065\u0063\u0026\u0071\u0075\u006f\u0074\u003b\u0032\u0030\u0031\u0039\u002d\u0030\u0039\u002d\u0031\u0038\u0020\u0031\u0036\u003a\u0035\u0038\u003a\u0032\u0033\u002e\u0031\u0036\u0038")
	fmt.Printf("str:%s err:%v", str, nil)
	str, err := gbk2utf8("\xce\xe4\xba\xba\xcf\xeb")
	fmt.Printf("str:%s err:%v", str, err)
}

func SpecialLetters(letter rune) (bool, []rune) {
	if unicode.IsPunct(letter) || unicode.IsSymbol(letter) || unicode.Is(unicode.Han, letter) {
		var chars []rune
		chars = append(chars, '\\', letter)
		return true, chars
	}
	return false, nil
}

func TestSpecailLetter(t *testing.T) {
	str := `1234~!@#$%^&*()_+}{":?><"⌘开`
	var chars []rune
	for _, letter := range str {
		ok, letters := SpecialLetters(letter)
		if ok {
			chars = append(chars, letters...)
		} else {
			chars = append(chars, letter)
		}
	}
	fmt.Println(string(chars))
}

// Unicode 转 UTF-8
func unicode2utf8(source string) string {
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
	return strings.Join(res, "")
}
func UnescapeUnicode(raw string) (string, error) {
	str, err := strconv.Unquote(strings.Replace(strconv.Quote(raw), `\\u`, `\u`, -1))
	if err != nil {
		return "", err
	}
	return str, nil
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
