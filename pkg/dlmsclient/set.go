package dlmsclient

import (
	"fmt"

	"github.com/Circutor/gosem/pkg/axdr"
	"github.com/Circutor/gosem/pkg/dlms"
)

func (c *client) SetRequest(att *dlms.AttributeDescriptor, data interface{}) (err error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

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
