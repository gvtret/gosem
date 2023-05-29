package dlmsclient

import (
	"fmt"

	"gitlab.com/circutor-library/gosem/pkg/axdr"
	"gitlab.com/circutor-library/gosem/pkg/dlms"
)

func (c *client) ActionRequest(mth *dlms.MethodDescriptor, data interface{}) (err error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if mth == nil {
		return dlms.NewError(dlms.ErrorInvalidParameter, "method descriptor must be non-nil")
	}

	dt, ok := data.(*axdr.DlmsData)
	if !ok {
		dt, err = axdr.MarshalData(data)
		if err != nil {
			return dlms.NewError(dlms.ErrorInvalidParameter, fmt.Sprintf("error marshaling %s data: %v", mth.String(), err))
		}
	}

	req := dlms.CreateActionRequestNormal(unicastInvokeID, *mth, dt)

	pdu, err := c.encodeSendReceiveAndDecode(req)
	if err != nil {
		return
	}

	resp, ok := pdu.(dlms.ActionResponseNormal)
	if !ok {
		return dlms.NewError(dlms.ErrorInvalidResponse, fmt.Sprintf("in %s unexpected PDU response type: %T", mth.String(), pdu))
	}

	if resp.Response.Result != dlms.TagActSuccess {
		return dlms.NewError(dlms.ErrorActionRejected, fmt.Sprintf("action %s rejected: %s", mth.String(), resp.Response.Result.String()))
	}

	return
}
