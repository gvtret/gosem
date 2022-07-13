package dlmsclient

import (
	"fmt"

	"github.com/Circutor/gosem/pkg/axdr"
	"github.com/Circutor/gosem/pkg/dlms"
)

func (c *Client) SetRequest(att *dlms.AttributeDescriptor, data interface{}) (err error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if att == nil {
		return NewError(ErrorInvalidParameter, "attribute descriptor must be non-nil")
	}

	dt, ok := data.(*axdr.DlmsData)
	if !ok {
		dt, err = axdr.MarshalData(data)
		if err != nil {
			return NewError(ErrorInvalidParameter, fmt.Sprintf("error marshaling data: %v", err))
		}
	}

	req := dlms.CreateSetRequestNormal(unicastInvokeID, *att, nil, *dt)

	pdu, err := c.encodeSendReceiveAndDecode(req)
	if err != nil {
		return
	}

	resp, ok := pdu.(dlms.SetResponseNormal)
	if !ok {
		return NewError(ErrorInvalidResponse, fmt.Sprintf("unexpected PDU type: %T", pdu))
	}

	if resp.Result != dlms.TagAccSuccess {
		return NewError(ErrorSetRejected, fmt.Sprintf("set rejected: %s", resp.Result.String()))
	}

	return
}
