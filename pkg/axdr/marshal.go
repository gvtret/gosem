package axdr

import (
	"fmt"
	"reflect"
	"time"
)

func MarshalData(v interface{}) (*DlmsData, error) {
	rv := eindirect(reflect.ValueOf(v))
	return encode(rv)
}

func encode(rv reflect.Value) (data *DlmsData, err error) {
	if !rv.IsValid() {
		return nil, fmt.Errorf("invalid value")
	}

	_, isTime := rv.Interface().(time.Time)
	if isTime {
		data = CreateAxdrOctetString(rv.Interface().(time.Time))
		return
	}

	k := rv.Kind()
	switch k {
	case reflect.Int8:
		data = CreateAxdrInteger(int8(rv.Int()))
	case reflect.Int16:
		data = CreateAxdrLong(int16(rv.Int()))
	case reflect.Int32:
		data = CreateAxdrDoubleLong(int32(rv.Int()))
	case reflect.Int64:
		data = CreateAxdrLong64(rv.Int())
	case reflect.Uint8:
		data = CreateAxdrUnsigned(uint8(rv.Uint()))
	case reflect.Uint16:
		data = CreateAxdrLongUnsigned(uint16(rv.Uint()))
	case reflect.Uint32:
		data = CreateAxdrDoubleLongUnsigned(uint32(rv.Uint()))
	case reflect.Uint64:
		data = CreateAxdrLong64Unsigned(rv.Uint())
	case reflect.Float32:
		data = CreateAxdrFloat32(float32(rv.Float()))
	case reflect.Float64:
		data = CreateAxdrFloat64(rv.Float())
	case reflect.Bool:
		data = CreateAxdrBoolean(rv.Bool())
	case reflect.String:
		data = CreateAxdrOctetString(rv.String())
	case reflect.Array, reflect.Slice:
		axdrArray := make([]*DlmsData, rv.Len())
		for i := 0; i < rv.Len(); i++ {
			axdrArray[i], err = encode(rv.Index(i))
			if err != nil {
				return nil, fmt.Errorf("element[%d]: %w", i, err)
			}
		}
		data = CreateAxdrArray(axdrArray)
	case reflect.Struct:
		axdrStruct := make([]*DlmsData, rv.NumField())
		for i := 0; i < rv.NumField(); i++ {
			axdrStruct[i], err = encode(rv.Field(i))
			if err != nil {
				return nil, fmt.Errorf("field[%s]: %w", rv.Type().Field(i).Name, err)
			}
		}
		data = CreateAxdrStructure(axdrStruct)
	case reflect.Ptr, reflect.Interface:
		data, err = encode(rv.Elem())
	default:
		return nil, fmt.Errorf("unsupported type: %s", k)
	}

	return
}

func eindirect(v reflect.Value) reflect.Value {
	switch v.Kind() {
	case reflect.Ptr, reflect.Interface:
		return eindirect(v.Elem())
	default:
		return v
	}
}
