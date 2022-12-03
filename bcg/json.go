package bcg

import (
	"bytes"
	"encoding/json"
	"strconv"
)

/* Golang 的数组、结构等变量，实际上都是引用。
理论上数字字串也是引用，但是它们本身是不可改变的，所以是引用还是复制是没有区别的。
*/

type JsonObject map[string]interface{}

func (js *JsonObject) ToString() string {
	b, err := json.Marshal(js)
	if err != nil {
		LogRed(err.Error())
		return ""
	}
	return string(b)
}
func (js *JsonObject) ToBytes() []byte {
	b, err := json.Marshal(js)
	if err != nil {
		LogRed(err.Error())
	}
	return b
}

//如果 str 不是用大括号括起来的 Object 字串，ParseString 就会失败，
//ParseAnyString 会把字串作为一个 key 来分析，这样只要是合法的 Json
//字串，就能够返回正确的对象，比如布尔、字串，数字等等。
func (js *JsonObject) ParseAnyString(str string) (interface{}, bool) {
	s := `{"a":` + str + `}`
	if nil != js.ParseString(s) {
		return nil, false
	}
	j := js.GetInterface("a")
	return j, true
}
func (js *JsonObject) ParseString(str string) error {
	return json.Unmarshal([]byte(str), js)
}
func (js *JsonObject) ToStruct(st interface{}) error {
	data := js.ToBytes()
	return json.Unmarshal(data, st)
}
func (js *JsonObject) ParseBytes(b []byte) error {
	return json.Unmarshal(b, js)
}
func (js *JsonObject) GetInterface(key string) interface{} {
	return (*js)[key]
}

//GetStrInt64 这个函数只是用于取回字串是数字的值，在json字串里面对应的属性仍然是字串，这是因为float64转int64是存在精度误差的
func (js *JsonObject) GetStrInt64(key string) (int64, bool) {
	sv, rs := js.GetString(key)
	if !rs {
		return 0, false
	}
	v, err := strconv.ParseInt(sv, 10, 64)
	return v, err == nil
}
func (js *JsonObject) GetString(key string) (string, bool) {
	val := (*js)[key]
	switch val.(type) {
	case string:
		return val.(string), true
	default:
		return "", false
	}
}
func (js *JsonObject) GetFloat64(key string) (float64, bool) {
	val := (*js)[key]
	switch val.(type) {
	case float64:
		return val.(float64), true
	default:
		return 0, false
	}
}
func (js *JsonObject) GetArray(key string) ([]interface{}, bool) {
	val := (*js)[key]
	switch val.(type) {
	case []interface{}:
		return val.([]interface{}), true
	default:
		return []interface{}{}, false
	}
}
func (js *JsonObject) GetJson(key string) (JsonObject, bool) {
	v := (*js)[key]
	switch v.(type) {
	case map[string]interface{}:
		jn := v.(map[string]interface{})
		return jn, true
	default:
		return nil, false
	}
}
func (js *JsonObject) GetBool(key string) (bool, bool) {
	val := (*js)[key]
	switch val.(type) {
	case bool:
		return val.(bool), true
	default:
		return false, false
	}
}
func (js *JsonObject) SetValueTo(key string, v interface{}) (r bool) {
	var val float64
	switch v.(type) {
	case *bool:
		*v.(*bool), r = js.GetBool(key)
	case *string:
		*v.(*string), r = js.GetString(key)
	case *float64:
		*v.(*float64), r = js.GetFloat64(key)
	case *int64:
		val, r = js.GetFloat64(key)
		if r {
			*v.(*int64) = int64(val)
		}
	case *uint64:
		val, r = js.GetFloat64(key)
		if r {
			*v.(*uint64) = uint64(val)
		}
	case *int:
		val, r = js.GetFloat64(key)
		if r {
			*v.(*int) = int(val)
		}
	case *uint:
		val, r = js.GetFloat64(key)
		if r {
			*v.(*uint) = uint(val)
		}
	case *int32:
		val, r = js.GetFloat64(key)
		if r {
			*v.(*int32) = int32(val)
		}
	case *uint32:
		val, r = js.GetFloat64(key)
		if r {
			*v.(*uint32) = uint32(val)
		}
	case *int16:
		val, r = js.GetFloat64(key)
		if r {
			*v.(*int16) = int16(val)
		}
	case *uint16:
		val, r = js.GetFloat64(key)
		if r {
			*v.(*uint16) = uint16(val)
		}
	case *int8:
		val, r = js.GetFloat64(key)
		if r {
			*v.(*int8) = int8(val)
		}
	case *uint8:
		val, r = js.GetFloat64(key)
		if r {
			*v.(*uint8) = uint8(val)
		}
	case *JsonObject:
		*v.(*JsonObject), r = js.GetJson(key)
	case *[]interface{}:
		*v.(*[]interface{}), r = js.GetArray(key)
	}
	return r
}

//SetValue 这个函数会设置map对应key 值，但是可能和读取时是不一样的，因为反编译json字串的对象都是map，而设置的不一定，
//你可以设置一个其它对象，它会被json序列化，而反序列化并不能还原这个对象，要反序列化为特定的对象使用 SetValueTo 函数
func (js *JsonObject) SetValue(key string, v interface{}) {
	(*js)[key] = v
}
func (js *JsonObject) SetStrInt64(key string, v int64) {
	vs := strconv.FormatInt(v, 10)
	(*js)[key] = vs
}
func (js *JsonObject) Delete(key string) {
	delete(*js, key)
}
func (js *JsonObject) Clear() {
	for key, _ := range *js {
		delete(*js, key)
	}
}
func (js *JsonObject) Marshal(v interface{}) {
	b, err := json.Marshal(v)
	if err != nil {
		LogRed(err.Error())
		return
	}
	js.ParseBytes(b)
}
func NewJsonObject() JsonObject {
	return JsonObject{}
}

type JsonArray []interface{}

func (ja *JsonArray) ParseString(str string) error {
	return json.Unmarshal([]byte(str), ja)
}

func (ja *JsonArray) ParseBytes(b []byte) error {
	return json.Unmarshal(b, ja)
}
func (ja *JsonArray) UnMarshal(st interface{}) error {
	data := ja.ToBytes()
	return json.Unmarshal(data, st)
}
func (ja *JsonArray) Marshal(v interface{}) {
	b, err := json.Marshal(v)
	if err != nil {
		LogRed(err.Error())
		return
	}
	ja.ParseBytes(b)
}
func (ja *JsonArray) ToString() string {
	b, err := json.Marshal(ja)
	if err != nil {
		LogRed(err.Error())
		return ""
	}
	return string(b)
}
func (ja *JsonArray) ToBytes() []byte {
	b, err := json.Marshal(ja)
	if err != nil {
		LogRed(err.Error())
	}
	return b
}

func NewJsonArray() JsonArray {
	return JsonArray{}
}

func JsonParseString(src string, pobj interface{}) bool {
	err := json.Unmarshal([]byte(src), pobj)
	return err == nil
}
func JsonParseBytes(src []byte, pobj interface{}) bool {
	err := json.Unmarshal(src, pobj)
	return err == nil
}
func JsonToString(obj interface{}, format bool) string {
	data, _ := json.Marshal(obj)
	if format {
		var str bytes.Buffer
		err := json.Indent(&str, data, "", "\t")
		if err != nil {
			return ""
		}
		return str.String()
	} else {
		return string(data)
	}
}
func JsonToBytes(obj interface{}, format bool) []byte {
	data, _ := json.Marshal(obj)
	if format {
		var str bytes.Buffer
		err := json.Indent(&str, data, "", "\t")
		if err != nil {
			return nil
		}
		return str.Bytes()
	} else {
		return data
	}
}
func JsonLoadConf(fn string, conf interface{}) bool {
	data := ReadFile(fn)
	return JsonParseBytes(data, conf)
}
func JsonSaveConf(fn string, conf interface{}) bool {
	data := JsonToBytes(conf, true)
	return nil == SaveFile(fn, data)
}
