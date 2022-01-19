package client

import (
	"fmt"

	"github.com/Circutor/gosem/pkg/axdr"
	"github.com/Circutor/gosem/pkg/dlms"
)

func (c *Client) ActionRequest(mth *dlms.MethodDescriptor, data interface{}) (err error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if mth == nil {
		err = fmt.Errorf("method descriptor is nil")
		return
	}

	dt, ok := data.(*axdr.DlmsData)
	if !ok {
		dt, err = axdr.MarshalData(data)
		if err != nil {
			err = fmt.Errorf("failed to marshal data: %w", err)
			return
		}
	}

	req := dlms.CreateActionRequestNormal(unicastInvokeID, *mth, dt)

	pdu, err := c.encodeSendReceiveAndDecode(req)
	if err != nil {
		return
	}

	resp, ok := pdu.(dlms.ActionResponseNormal)
	if !ok {
		err = fmt.Errorf("expected action response, got %T", pdu)
		return
	}

	if resp.Response.Result != dlms.TagActSuccess {
		err = fmt.Errorf("action failed with result: %s", resp.Response.Result.String())
		return
	}

	return
}
