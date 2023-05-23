package dlms

import (
	"bytes"
	"time"

	"github.com/Circutor/gosem/pkg/axdr"
)

type EventNotificationRequest struct {
	Time           *time.Time
	AttributeInfo  AttributeDescriptor
	AttributeValue axdr.DlmsData
}

func CreateEventNotificationRequest(tm *time.Time, attInfo AttributeDescriptor, attValue axdr.DlmsData) *EventNotificationRequest {
	return &EventNotificationRequest{
		Time:           tm,
		AttributeInfo:  attInfo,
		AttributeValue: attValue,
	}
}

func (ev EventNotificationRequest) Encode() (out []byte, err error) {
	var buf bytes.Buffer
	buf.WriteByte(byte(TagEventNotificationRequest))
	if ev.Time == nil {
		buf.WriteByte(0)
	} else {
		buf.WriteByte(1)
		tm, e := axdr.EncodeDateTime(*ev.Time)
		if e != nil {
			err = e
			return
		}
		buf.WriteByte(uint8(len(tm)))
		buf.Write(tm)
	}
	attInfo, eInfo := ev.AttributeInfo.Encode()
	if eInfo != nil {
		err = eInfo
		return
	}
	buf.Write(attInfo)
	attValue, eValue := ev.AttributeValue.Encode()
	if eValue != nil {
		err = eValue
		return
	}
	buf.Write(attValue)

	out = buf.Bytes()
	return
}

func DecodeEventNotificationRequest(ori *[]byte) (out EventNotificationRequest, err error) {
	src := *ori

	_, tag, _ := axdr.DecodeUnsigned(&src)
	if tag != TagEventNotificationRequest.Value() {
		err = ErrWrongTag(0, tag, byte(TagDataNotification))
		return
	}

	_, haveTime, err := axdr.DecodeUnsigned(&src)
	if err != nil {
		return
	}

	if haveTime == 0x0 {
		out.Time = nil
	} else {
		axdr.DecodeUnsigned(&src) // length of time
		_, time, e := axdr.DecodeDateTime(&src)
		if e != nil {
			err = e
			return
		}
		out.Time = &time
	}

	out.AttributeInfo, err = DecodeAttributeDescriptor(&src)
	if err != nil {
		return
	}

	decoder := axdr.NewDataDecoder(&src)
	out.AttributeValue, err = decoder.Decode(&src)

	(*ori) = (*ori)[len(*ori)-len(src):]
	return
}
