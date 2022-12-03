package bcg

import (
	"crypto/rand"
	"encoding/binary"
	"encoding/hex"
	"math/big"
	"runtime"
	"strings"
	"time"
)

func CheckError(err error) bool {
	if err != nil {
		LogTrace(TextRed, 1, err.Error())
		return true
	}
	return false
}

//CheckErrTrace
//trace = 1 记录使用这个函数的位置，trace = 2 记录上一级
func CheckErrTrace(err error, trace uint) bool {
	if err != nil {
		LogTrace(TextRed, trace, err.Error())
		return true
	}
	return false
}

const (
	FormatDateTime = "2006-01-02 15:04:05"
	FormatDate     = "2006-01-02"
)

func GetNowDate() string {
	return time.Now().Format(FormatDateTime)
}

// HexParse 解析 Hex 字串
func HexParse(h string) []byte {
	data, _ := hex.DecodeString(h)
	//CheckError(err)
	return data
}

//HexString v 可以是某种整数类型或者字串，或者字节数组
func HexString(v interface{}) (str string) {
	switch v.(type) {
	case []byte:
		str = hex.EncodeToString(v.([]byte))
	case string:
		str = hex.EncodeToString([]byte(v.(string)))
	default:
		buf := IntToBytes(v)
		if buf == nil {
			return "type not exact"
		}
		str = hex.EncodeToString(buf)
	}
	return strings.ToUpper(str)
}

//IntToBytes v 必须是某种 int 类型
func IntToBytes(v interface{}) (buf []byte) {
	switch v.(type) {
	case byte:
		buf = make([]byte, 1)
		buf[0] = v.(byte)
	case uint16:
		buf = make([]byte, 2)
		binary.BigEndian.PutUint16(buf, v.(uint16))
	case uint32:
		buf = make([]byte, 4)
		binary.BigEndian.PutUint32(buf, v.(uint32))
	case uint64:
		buf = make([]byte, 8)
		binary.BigEndian.PutUint64(buf, v.(uint64))
	case int8:
		vi := v.(int8)
		buf = make([]byte, 1)
		if vi >= 0 {
			buf[0] = byte(vi)
		} else {
			buf[0] = ^byte(^vi)
		}
	case int16:
		vi := v.(int16)
		buf = make([]byte, 2)
		if vi >= 0 {
			binary.BigEndian.PutUint16(buf, uint16(vi))
		} else {
			binary.BigEndian.PutUint16(buf, ^uint16(^vi))
		}
	case int32:
		vi := v.(int32)
		buf = make([]byte, 4)
		if vi >= 0 {
			binary.BigEndian.PutUint32(buf, uint32(vi))
		} else {
			binary.BigEndian.PutUint32(buf, ^uint32(^vi))
		}
	case int64:
		vi := v.(int64)
		buf = make([]byte, 8)
		if vi >= 0 {
			binary.BigEndian.PutUint64(buf, uint64(vi))
		} else {
			binary.BigEndian.PutUint64(buf, ^uint64(^vi))
		}
	case int:
		vi := v.(int)
		if (-1 << 32) == 0 {
			return IntToBytes(int32(vi))
		} else {
			return IntToBytes(int64(vi))
		}
	default:
		LogTrace(TextRed, 1, true, "mlib.IntToBytes() error type:", v)
	}
	return buf
}

func OsIsWin() bool {
	return runtime.GOOS == "windows"
}

//Rand58String 返回一个随机字串，使用不容易混淆的58个字符
func Rand58String(n int) string {
	allowedChars := "123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz"
	max := len(allowedChars)
	b := ""
	for i := 0; i < n; i++ {
		r, _ := rand.Int(rand.Reader, big.NewInt(int64(max)))
		pos := int(r.Int64())
		b += allowedChars[pos : pos+1]
	}
	return b
}
