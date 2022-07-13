package dlmsclient

import (
	"fmt"

	"github.com/Circutor/gosem/pkg/axdr"
	"github.com/Circutor/gosem/pkg/dlms"
)

func (c *client) ActionRequest(mth *dlms.MethodDescriptor, data interface{}) (err error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if mth == nil {
		return NewError(ErrorInvalidParameter, "method descriptor must be non-nil")
	}

	dt, ok := data.(*axdr.DlmsData)
	if !ok {
		dt, err = axdr.MarshalData(data)
		if err != nil {
			return NewError(ErrorInvalidParameter, fmt.Sprintf("error marshaling data: %v", err))
		}
	}

	req := dlms.CreateActionRequestNormal(unicastInvokeID, *mth, dt)

	pdu, err := c.encodeSendReceiveAndDecode(req)
	if err != nil {
		return
	}

	resp, ok := pdu.(dlms.ActionResponseNormal)
	if !ok {
		return NewError(ErrorInvalidResponse, fmt.Sprintf("unexpected PDU type: %T", pdu))
	}

	if resp.Response.Result != dlms.TagActSuccess {
		return NewError(ErrorActionRejected, fmt.Sprintf("action rejected: %s", resp.Response.Result.String()))
	}

	return
}
