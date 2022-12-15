package dlmsclient

import (
	"fmt"
	"reflect"

	"github.com/Circutor/gosem/pkg/axdr"
	"github.com/Circutor/gosem/pkg/dlms"
)

func (c *client) SetRequest(att *dlms.AttributeDescriptor, data interface{}) (err error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	return c.setRequest(att, data)
}

func (c *client) SetRequestWithStructOfElements(data interface{}) (err error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	v := eindirect(reflect.ValueOf(data))

	if v.Kind() != reflect.Struct {
		return dlms.NewError(dlms.ErrorInvalidParameter, "data must be a struct")
	}

	isSomethingDone := false

	for i := 0; i < v.NumField(); i++ {
		ad, err := c.getAttributeDescriptor(v.Type().Field(i))
		if err != nil {
			return err
		}

		if ad == nil {
			continue
		}

		if v.Field(i).Kind() == reflect.Pointer {
			// All fields need to have been set beforehand: nil fields will be ignored
			if v.Field(i).IsNil() {
				continue
			}
		}

		err = c.setRequest(ad, v.Field(i).Interface())
		if err != nil {
			if isSomethingDone {
				err = dlms.NewError(dlms.ErrorSetPartial, fmt.Sprintf("partial set: %v", err))
			}

			return err
		}

		isSomethingDone = true
	}

	return nil
}

func (c *client) setRequest(att *dlms.AttributeDescriptor, data interface{}) (err error) {
	if att == nil {
		return dlms.NewError(dlms.ErrorInvalidParameter, "attribute descriptor must be non-nil")
	}

	dt, ok := data.(*axdr.DlmsData)
	if !ok {
		dt, err = axdr.MarshalData(data)
		if err != nil {
			return dlms.NewError(dlms.ErrorInvalidParameter, fmt.Sprintf("error marshaling %s data: %v", att.String(), err))
		}
	}

	req := dlms.CreateSetRequestNormal(unicastInvokeID, *att, nil, *dt)

	pdu, err := c.encodeSendReceiveAndDecode(req)
	if err != nil {
		return
	}

	resp, ok := pdu.(dlms.SetResponseNormal)
	if !ok {
		return dlms.NewError(dlms.ErrorInvalidResponse, fmt.Sprintf("in %s unexpected PDU response type: %T", att.String(), pdu))
	}

	if resp.Result != dlms.TagAccSuccess {
		return dlms.NewError(dlms.ErrorSetRejected, fmt.Sprintf("set %s rejected: %s", att.String(), resp.Result.String()))
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
