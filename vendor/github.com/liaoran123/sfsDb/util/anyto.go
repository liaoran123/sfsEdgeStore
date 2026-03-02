package util

import (
	"encoding/json"
	"strconv"
	"time"
)

func AnyToStr(i any) string {
	if i == nil {
		return ""
	}
	switch v := i.(type) {
	case int:
		return strconv.Itoa(v)
	case int64:
		return strconv.FormatInt(v, 10)
	case int32:
		return strconv.FormatInt(int64(v), 10)
	case int16:
		return strconv.FormatInt(int64(v), 10)
	case int8:
		return strconv.FormatInt(int64(v), 10)
	case uint:
		return strconv.FormatUint(uint64(v), 10)
	case uint64:
		return strconv.FormatUint(v, 10)
	case uint32:
		return strconv.FormatUint(uint64(v), 10)
	case uint16:
		return strconv.FormatUint(uint64(v), 10)
	case uint8:
		return strconv.FormatUint(uint64(v), 10)
	case time.Time:
		return v.Format(time.DateTime)
	case bool:
		return strconv.FormatBool(v)
	case float64:
		return strconv.FormatFloat(v, 'f', 6, 64)
	case float32:
		return strconv.FormatFloat(float64(v), 'f', 6, 32)
	case complex64:
		return strconv.FormatComplex(complex128(v), 'f', 6, 64)
	case complex128:
		return strconv.FormatComplex(v, 'f', 6, 128)
	case string:
		return v
	default: //默认使用序列化json
		bytes, err := json.Marshal(v)
		if err != nil {
			return err.Error()
		}
		return string(bytes)
	}
}

// StrToAny 将字符串转换为指定类型的值，返回转换后的值和可能的错误
func StrToAny(i string, t any) (any, error) {
	switch t.(type) {
	case int:
		v, err := strconv.Atoi(i)
		if err != nil {
			return 0, err
		}
		return v, nil
	case int64:
		v, err := strconv.ParseInt(i, 10, 64)
		if err != nil {
			return 0, err
		}
		return v, nil
	case int32:
		v, err := strconv.ParseInt(i, 10, 32)
		if err != nil {
			return 0, err
		}
		return int32(v), nil
	case int16:
		v, err := strconv.ParseInt(i, 10, 16)
		if err != nil {
			return 0, err
		}
		return int16(v), nil
	case int8:
		v, err := strconv.ParseInt(i, 10, 8)
		if err != nil {
			return 0, err
		}
		return int8(v), nil
	case uint:
		v, err := strconv.ParseUint(i, 10, intSize)
		if err != nil {
			return 0, err
		}
		return uint(v), nil
	case uint64:
		v, err := strconv.ParseUint(i, 10, 64)
		if err != nil {
			return 0, err
		}
		return v, nil
	case uint32:
		v, err := strconv.ParseUint(i, 10, 32)
		if err != nil {
			return 0, err
		}
		return uint32(v), nil
	case uint16:
		v, err := strconv.ParseUint(i, 10, 16)
		if err != nil {
			return 0, err
		}
		return uint16(v), nil
	case uint8:
		v, err := strconv.ParseUint(i, 10, 8)
		if err != nil {
			return 0, err
		}
		return uint8(v), nil
	case time.Time:
		v, err := time.Parse(time.DateTime, i)
		if err != nil {
			return time.Time{}, err
		}
		// 使用本地时区，与AnyToStr函数保持一致
		return time.Date(v.Year(), v.Month(), v.Day(), v.Hour(), v.Minute(), v.Second(), v.Nanosecond(), time.Local), nil
	case bool:
		v, err := strconv.ParseBool(i)
		if err != nil {
			return false, err
		}
		return v, nil
	case float64:
		v, err := strconv.ParseFloat(i, 64)
		if err != nil {
			return 0, err
		}
		return v, nil
	case float32:
		v, err := strconv.ParseFloat(i, 32)
		if err != nil {
			return 0, err
		}
		return float32(v), nil
	case complex64:
		v, err := strconv.ParseComplex(i, 64)
		if err != nil {
			return 0, err
		}
		return complex64(v), nil
	case complex128:
		v, err := strconv.ParseComplex(i, 128)
		if err != nil {
			return 0, err
		}
		return v, nil
	case string:
		return i, nil
	default: //默认使用反序列化json
		err := json.Unmarshal([]byte(i), t)
		if err != nil {
			return i, err
		}
		return t, nil
	}
}

// 根任意数据类型转换[]byte函数
// 只有字符串和整型可以进行排序，故除整型外，其他类型都返回v.(string)的[]byte
func AnyToBytes(v any) []byte {
	if v == nil {
		return []byte(nil)
	}
	switch val := v.(type) {
	case int:
		return IntToBytes(val, EndianOrder)
	case int8:
		return []byte{byte(val)}
	case uint8:
		return []byte{val}
	case int16:
		return Int16ToBytes(val, EndianOrder)
	case int32:
		return Int32ToBytes(val, EndianOrder)
	case int64:
		return Int64ToBytes(val, EndianOrder)
	case uint:
		if intSize == 32 {
			return Uint32ToBytes(uint32(val), EndianOrder)
		}
		return Uint64ToBytes(uint64(val), EndianOrder)
	case uint16:
		return Uint16ToBytes(val, EndianOrder)
	case uint32:
		return Uint32ToBytes(val, EndianOrder)
	case uint64:
		return Uint64ToBytes(val, EndianOrder)
	case float32:
		return []byte(strconv.FormatFloat(float64(val), 'f', 6, 32))
	case float64:
		return []byte(strconv.FormatFloat(val, 'f', 6, 64))
	case complex64:
		return []byte(strconv.FormatComplex(complex128(val), 'f', 6, 64))
	case complex128:
		return []byte(strconv.FormatComplex(val, 'f', 6, 128))
	case bool:
		if val {
			return []byte{1}
		}
		return []byte{0}
		//return []byte(strconv.FormatBool(val)) //return []byte(strconv.FormatBool(val))
	case string:
		return []byte(val)
	case time.Time:
		return []byte(val.Format(time.DateTime))
	default:
		//默认使用序列化json
		bytes, err := json.Marshal(v)
		if err != nil {
			return []byte(err.Error())
		}
		return bytes
	}
}
