package client

import (
	"fmt"

	"github.com/Circutor/gosem/pkg/axdr"
	"github.com/Circutor/gosem/pkg/dlms"
)

func (c *Client) SetRequest(att *dlms.AttributeDescriptor, data interface{}) (err error) {
	if att == nil {
		err = fmt.Errorf("attribute descriptor is nil")
		return
	}

	dt, ok := data.(*axdr.DlmsData)
	if !ok {
		err = fmt.Errorf("data is not a valid *axdr.DlmsData")
		return
	}

	req := dlms.CreateSetRequestNormal(unicastInvokeID, *att, nil, *dt)

	pdu, err := c.encodeSendReceiveAndDecode(req)
	if err != nil {
		return
	}

	resp, ok := pdu.(dlms.SetResponseNormal)
	if !ok {
		err = fmt.Errorf("expected SetResponseNormal, got %T", pdu)
		return
	}

	if resp.Result != dlms.TagAccSuccess {
		err = fmt.Errorf("set failed with result: %s", resp.Result.String())
		return
	}

	return
}
