package dlmsclient

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/Circutor/gosem/pkg/axdr"
	"github.com/Circutor/gosem/pkg/dlms"
)

func (c *client) GetRequest(att *dlms.AttributeDescriptor, data interface{}) (err error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	return c.getRequestWithUnmarshal(att, nil, data)
}

func (c *client) GetRequestWithSelectiveAccessByDate(att *dlms.AttributeDescriptor, start time.Time, end time.Time, data interface{}) (err error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	acc := dlms.CreateSelectiveAccessDescriptor(dlms.AccessSelectorRange, []time.Time{start, end})
	return c.getRequestWithUnmarshal(att, acc, data)
}

func (c *client) GetRequestWithStructOfElements(data interface{}) (err error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	return c.getRequestWithStructOfElements(data)
}

func (c *client) CheckRequestWithStructOfElements(data interface{}) (err error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	return c.checkRequestWithStructOfElements(data)
}

func (c *client) getAttributeDescriptor(field reflect.StructField) (*dlms.AttributeDescriptor, error) {
	tag := field.Tag.Get("obis")
	if tag == "" {
		return nil, nil
	}

	values := strings.Split(tag, ",")
	if len(values) != 3 {
		return nil, dlms.NewError(dlms.ErrorInvalidParameter, fmt.Sprintf("invalid obis tag: %s", tag))
	}

	class, err := strconv.ParseUint(values[0], 0, 16)
	if err != nil {
		return nil, dlms.NewError(dlms.ErrorInvalidParameter, fmt.Sprintf("invalid class: %s", tag))
	}
	obis := values[1]
	att, err := strconv.ParseUint(values[2], 0, 8)
	if err != nil {
		return nil, dlms.NewError(dlms.ErrorInvalidParameter, fmt.Sprintf("invalid attribute: %s", tag))
	}

	attribute := dlms.CreateAttributeDescriptor(uint16(class), obis, int8(att))

	return attribute, nil
}

func (c *client) getRequestWithUnmarshal(att *dlms.AttributeDescriptor, acc *dlms.SelectiveAccessDescriptor, data interface{}) (err error) {
	axdrData, err := c.getRequest(att, acc)
	if err != nil {
		return
	}

	if data != nil {
		err = axdr.UnmarshalData(axdrData, data)
		if err != nil {
			return dlms.NewError(dlms.ErrorInvalidResponse, fmt.Sprintf("error unmarshaling %s data: %v", att.String(), err))
		}
	}

	return
}

func (c *client) getRequest(att *dlms.AttributeDescriptor, acc *dlms.SelectiveAccessDescriptor) (data axdr.DlmsData, err error) {
	if att == nil {
		err = dlms.NewError(dlms.ErrorInvalidParameter, "attribute descriptor cannot be nil")
		return
	}

	req := dlms.CreateGetRequestNormal(unicastInvokeID, *att, acc)

	pdu, err := c.encodeSendReceiveAndDecode(req)
	if err != nil {
		return
	}

	switch resp := pdu.(type) {
	case dlms.GetResponseNormal:
		data, err = resp.Result.ValueAsData()
		if err != nil {
			access, _ := resp.Result.ValueAsAccess()
			err = dlms.NewError(dlms.ErrorGetRejected, fmt.Sprintf("get %s rejected: %s", att.String(), access.String()))
		}
	case dlms.GetResponseWithDataBlock:
		blockNumber := 1
		out := make([]byte, 0)
		for {
			if resp.Result.IsResult {
				access, _ := resp.Result.ResultAsAccess()
				err = dlms.NewError(dlms.ErrorGetRejected, fmt.Sprintf("get %s rejected: %s", att.String(), access.String()))
				return
			}

			if blockNumber != int(resp.Result.BlockNumber) {
				err = dlms.NewError(dlms.ErrorInvalidResponse, fmt.Sprintf("block number mismatch in %s: expected %d, got %d", att.String(), blockNumber, resp.Result.BlockNumber))
				return
			}

			res, _ := resp.Result.ResultAsBytes()
			out = append(out, res...)

			if resp.Result.LastBlock {
				break
			}

			req := dlms.CreateGetRequestNext(unicastInvokeID, uint32(blockNumber))
			blockNumber++

			pdu, err = c.encodeSendReceiveAndDecode(req)
			if err != nil {
				return
			}

			var ok bool
			resp, ok = pdu.(dlms.GetResponseWithDataBlock)
			if !ok {
				err = dlms.NewError(dlms.ErrorInvalidResponse, fmt.Sprintf("in %s expected GetResponseWithDataBlock response, got %T", att.String(), pdu))
				return
			}
		}

		decoder := axdr.NewDataDecoder(&out)
		data, err = decoder.Decode(&out)
		if err != nil {
			err = dlms.NewError(dlms.ErrorInvalidResponse, fmt.Sprintf("error decoding %s data: %v", att.String(), err))
			return
		}
	default:
		err = dlms.NewError(dlms.ErrorInvalidResponse, fmt.Sprintf("in %s unexpected PDU response type: %T", att.String(), pdu))
	}

	return
}

func (c *client) getRequestWithStructOfElements(data interface{}) (err error) {
	rv := reflect.ValueOf(data)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return dlms.NewError(dlms.ErrorInvalidParameter, "data must be a non-nil pointer")
	}

	v := reflect.Indirect(rv)
	if v.Kind() != reflect.Struct {
		return dlms.NewError(dlms.ErrorInvalidParameter, "data must be a pointer to a struct")
	}

	for i := 0; i < v.NumField(); i++ {
		ad, err := c.getAttributeDescriptor(v.Type().Field(i))
		if err != nil {
			return err
		}

		field := v.Field(i)

		if ad != nil {
			err = c.getRequestWithUnmarshal(ad, nil, field.Addr().Interface())
			if err != nil {
				// If a get is rejected in a field which is a pointer, then we will continue without any error
				var dlmsError *dlms.Error
				if errors.As(err, &dlmsError) && dlmsError.Code() == dlms.ErrorGetRejected && field.Kind() == reflect.Ptr {
					field.Set(reflect.Zero(field.Type()))
				} else {
					return err
				}
			}
		} else if field.Kind() == reflect.Struct {
			err = c.getRequestWithStructOfElements(field.Addr().Interface())
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (c *client) checkRequestWithStructOfElements(data interface{}) (err error) {
	rv := reflect.ValueOf(data)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return dlms.NewError(dlms.ErrorInvalidParameter, "data must be a non-nil pointer")
	}

	v := reflect.Indirect(rv)
	if v.Kind() != reflect.Struct {
		return dlms.NewError(dlms.ErrorInvalidParameter, "data must be a pointer to a struct")
	}

	for i := 0; i < v.NumField(); i++ {
		ad, err := c.getAttributeDescriptor(v.Type().Field(i))
		if err != nil {
			return err
		}

		field := v.Field(i)

		if ad != nil {
			if v.Field(i).Kind() == reflect.Pointer {
				// All nil fields will be ignored
				if v.Field(i).IsNil() {
					continue
				}
			}

			// Get expected value
			expected := reflect.Indirect(field).Interface()

			// Copy the expected value
			copy := reflect.New(reflect.Indirect(field).Type())

			err = c.getRequestWithUnmarshal(ad, nil, copy.Interface())
			if err != nil {
				return err
			}

			if str, ok := expected.(string); ok {
				expected = strings.ToLower(str)
			}

			// Get got value
			got := reflect.Indirect(copy).Interface()

			// Compare values
			if !reflect.DeepEqual(expected, got) {
				return dlms.NewError(dlms.ErrorCheckDoesNotMatch, fmt.Sprintf("values are not equal. Expected %v, got %v", expected, got))
			}
		} else if field.Kind() == reflect.Struct {
			err = c.checkRequestWithStructOfElements(field.Addr().Interface())
			if err != nil {
				return err
			}
		}
	}

	return nil
}
