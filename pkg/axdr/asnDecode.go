// Copyright (c) 2023 Circutor S.A. All rights reserved.

package axdr

import (
	"fmt"
	"strconv"
	"time"
)

func AsnDecode(value *DlmsData) (data string, err error) {
	if value == nil {
		return "", fmt.Errorf("value to encode cannot be nil")
	}

	switch value.Tag {
	case TagNull:
		data = strNull + "{}"
	case TagArray:
		data = strArray + "{"
		strArray := value.Value.([]*DlmsData)
		for _, str := range strArray {
			tmp, err := AsnDecode(str)
			if err != nil {
				return "", fmt.Errorf(nonEncodableError+"%w", err)
			}

			data += tmp
		}
		data += "}"
	case TagStructure:
		data = strStructure + "{"
		strStructure := value.Value.([]*DlmsData)
		for _, str := range strStructure {
			tmp, err := AsnDecode(str)
			if err != nil {
				return "", fmt.Errorf(nonEncodableError+"%w", err)
			}

			data += tmp
		}
		data += "}"
	case TagBoolean:
		data = fmt.Sprintf(strBoolean+"{%s}", strconv.FormatBool(value.Value.(bool)))
	case TagBitString:
		data = fmt.Sprintf(strBitString+"{%s}", value.Value.(string))
	case TagDoubleLong:
		data = fmt.Sprintf(strDoubleLong+"{%d}", (value.Value.(int32)))
	case TagDoubleLongUnsigned:
		data = fmt.Sprintf(strDoubleLongUnsigned+"{%d}", (value.Value.(uint32)))
	case TagFloatingPoint:
		data = fmt.Sprintf(strFloatingPoint+"{%g}", (value.Value.(float32)))
	case TagOctetString:
		data = fmt.Sprintf(strOctetString+"{%s}", value.Value.(string))
	case TagVisibleString:
		data = fmt.Sprintf(strVisibleString+"{%s}", value.Value.(string))
	case TagBCD:
		data = fmt.Sprintf(strBCD+"{%d}", value.Value.(int8))
	case TagInteger:
		data = fmt.Sprintf(strInteger+"{%d}", value.Value.(int8))
	case TagLong:
		data = fmt.Sprintf(strLong+"{%d}", value.Value.(int16))
	case TagUnsigned:
		data = fmt.Sprintf(strUnsigned+"{%d}", value.Value.(uint8))
	case TagCompactArray:
		data = strCompactArray + "{"
		strStructure := value.Value.([]*DlmsData)
		for _, str := range strStructure {
			tmp, err := AsnDecode(str)
			if err != nil {
				return "", fmt.Errorf(nonEncodableError+"%w", err)
			}

			data += tmp
		}
		data += "}"
	case TagLongUnsigned:
		data = fmt.Sprintf(strLongUnsigned+"{%d}", value.Value.(uint16))
	case TagLong64:
		data = fmt.Sprintf(strLong64+"{%d}", value.Value.(int64))
	case TagLong64Unsigned:
		data = fmt.Sprintf(strLong64Unsigned+"{%d}", value.Value.(uint64))
	case TagEnum:
		data = fmt.Sprintf(strEnum+"{%d}", value.Value.(uint8))
	case TagFloat32:
		data = fmt.Sprintf(strFloat32+"{%g}", value.Value.(float32))
	case TagFloat64:
		data = fmt.Sprintf(strFloat64+"{%g}", value.Value.(float64))
	case TagDateTime:
		data = fmt.Sprintf(strDateTime+"{%s}", value.Value.(time.Time).Format(dateTimeLayout))
	case TagDate:
		data = fmt.Sprintf(strDate+"{%s}", value.Value.(time.Time).Format(dateLayout))
	case TagTime:
		data = fmt.Sprintf(strTime+"{%s}", value.Value.(time.Time).Format(timeLayout))
	}

	return data, nil
}
