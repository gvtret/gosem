package client

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/Circutor/gosem/pkg/axdr"
	"github.com/Circutor/gosem/pkg/dlms"
)

func (c *Client) GetRequest(att *dlms.AttributeDescriptor, data interface{}) (err error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	return c.getRequestWithUnmarshal(att, nil, data)
}

func (c *Client) GetRequestWithSelectiveAccessByDate(att *dlms.AttributeDescriptor, start time.Time, end time.Time, data interface{}) (err error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	acc := dlms.CreateSelectiveAccessDescriptor(dlms.AccessSelectorRange, []time.Time{start, end})
	return c.getRequestWithUnmarshal(att, acc, data)
}

func (c *Client) GetRequestWithStructOfElements(data interface{}) (err error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	rv := reflect.ValueOf(data)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return fmt.Errorf("data must be a non-nil pointer")
	}

	v := reflect.Indirect(rv)
	if v.Kind() != reflect.Struct {
		return fmt.Errorf("data must be a pointer to a struct")
	}

	for i := 0; i < v.NumField(); i++ {
		ad, err := getAttributeDescriptor(v.Type().Field(i))
		if err != nil {
			return fmt.Errorf("error getting attribute descriptor: %w", err)
		}
		if ad == nil {
			continue
		}

		field := v.Field(i)
		err = c.getRequestWithUnmarshal(ad, nil, field.Addr().Interface())
		if err != nil {
			return fmt.Errorf("error getting attribute %s: %w", ad.InstanceID, err)
		}
	}

	return nil
}

func getAttributeDescriptor(field reflect.StructField) (*dlms.AttributeDescriptor, error) {
	tag := field.Tag.Get("obis")
	if tag == "" {
		return nil, nil
	}

	values := strings.Split(tag, ",")
	if len(values) != 3 {
		return nil, fmt.Errorf("invalid obis tag: %s", tag)
	}

	class, err := strconv.ParseUint(values[0], 0, 16)
	if err != nil {
		return nil, fmt.Errorf("invalid class: %s", tag)
	}
	obis := values[1]
	att, err := strconv.ParseUint(values[2], 0, 8)
	if err != nil {
		return nil, fmt.Errorf("invalid attribute: %s", tag)
	}

	attribute := dlms.CreateAttributeDescriptor(uint16(class), obis, int8(att))

	return attribute, nil
}

func (c *Client) getRequestWithUnmarshal(att *dlms.AttributeDescriptor, acc *dlms.SelectiveAccessDescriptor, data interface{}) (err error) {
	axdrData, err := c.getRequest(att, acc)
	if err != nil {
		return
	}

	if data != nil {
		err = axdr.UnmarshalData(axdrData, data)
		if err != nil {
			err = fmt.Errorf("error unmarshaling data: %w", err)
			return
		}
	}

	return
}

func (c *Client) getRequest(att *dlms.AttributeDescriptor, acc *dlms.SelectiveAccessDescriptor) (data axdr.DlmsData, err error) {
	if att == nil {
		err = fmt.Errorf("attribute descriptor is nil")
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
			err = fmt.Errorf("get failed with result: %s", access.String())
		}
	case dlms.GetResponseWithDataBlock:
		blockNumber := 1
		out := make([]byte, 0)
		for {
			if resp.Result.IsResult {
				access, _ := resp.Result.ResultAsAccess()
				err = fmt.Errorf("get failed with result: %s", access.String())
				return
			}

			if blockNumber != int(resp.Result.BlockNumber) {
				err = fmt.Errorf("block number mismatch: expected %d, got %d", blockNumber, resp.Result.BlockNumber)
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
				err = fmt.Errorf("expected GetResponseWithDataBlock, got %T", pdu)
				return
			}
		}

		decoder := axdr.NewDataDecoder(&out)
		data, err = decoder.Decode(&out)
		if err != nil {
			err = fmt.Errorf("error decoding data: %w", err)
			return
		}
	default:
		err = fmt.Errorf("unexpected CosemPDU type: %T", pdu)
	}

	return
}
