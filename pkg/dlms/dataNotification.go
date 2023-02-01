package dlms

import (
	"bytes"
	"time"

	"github.com/Circutor/gosem/pkg/axdr"
)

type DataNotification struct {
	InvokeIDAndPriority uint32
	DateTime            *time.Time
	DataValue           axdr.DlmsData
}

func CreateDataNotification(invokeIDAndPriority uint32, tm *time.Time, dataValue axdr.DlmsData) *DataNotification {
	return &DataNotification{
		InvokeIDAndPriority: invokeIDAndPriority,
		DateTime:            tm,
		DataValue:           dataValue,
	}
}

func (dn DataNotification) Encode() (out []byte, err error) {
	var buf bytes.Buffer
	buf.WriteByte(byte(TagDataNotification))

	invokeIDAndPriority, _ := axdr.EncodeDoubleLongUnsigned(dn.InvokeIDAndPriority)
	buf.Write(invokeIDAndPriority)

	if dn.DateTime == nil {
		buf.WriteByte(0)
	} else {
		buf.WriteByte(1)
		tm, _ := axdr.EncodeDateTime(*dn.DateTime)
		buf.WriteByte(uint8(len(tm)))
		buf.Write(tm)
	}

	dataValue, eValue := dn.DataValue.Encode()
	if eValue != nil {
		err = eValue
		return
	}
	buf.Write(dataValue)

	out = buf.Bytes()
	return
}

func DecodeDataNotification(ori *[]byte) (out DataNotification, err error) {
	src := append([]byte(nil), (*ori)...)

	_, tag, _ := axdr.DecodeUnsigned(&src)
	if tag != TagDataNotification.Value() {
		err = ErrWrongTag(0, tag, byte(TagDataNotification))
		return
	}

	_, invokeIDAndPriority, err := axdr.DecodeDoubleLongUnsigned(&src)
	if err != nil {
		return
	}
	out.InvokeIDAndPriority = invokeIDAndPriority

	_, haveTime, err := axdr.DecodeUnsigned(&src)
	if err != nil {
		return
	}

	if haveTime == 0x0 {
		out.DateTime = nil
	} else {
		axdr.DecodeUnsigned(&src) // length of time
		_, time, e := axdr.DecodeDateTime(&src)
		if e != nil {
			err = e
			return
		}
		out.DateTime = &time
	}

	decoder := axdr.NewDataDecoder(&src)
	out.DataValue, err = decoder.Decode(&src)

	(*ori) = (*ori)[len(*ori)-len(src):]
	return
}
