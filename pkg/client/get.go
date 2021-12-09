package client

import (
	"fmt"

	"github.com/Circutor/gosem/pkg/axdr"
	"github.com/Circutor/gosem/pkg/dlms"
)

func (c *Client) Get(att *dlms.AttributeDescriptor) (data axdr.DlmsData, err error) {
	if !c.isAssociated {
		err = fmt.Errorf("client is not associated")
		return
	}

	req := dlms.CreateGetRequestNormal(unicastInvokeID, *att, nil)
	src, err := req.Encode()
	if err != nil {
		err = fmt.Errorf("error encoding get request: %w", err)
		return
	}

	out, err := c.transport.Send(src)
	if err != nil {
		err = fmt.Errorf("error sending GetRequestNormal: %w", err)
		return
	}

	pdu, err := dlms.DecodeCosem(&out)
	if err != nil {
		err = fmt.Errorf("error decoding CosemPDU: %w", err)
		return
	}

	res, ok := pdu.(dlms.GetResponseNormal)
	if !ok {
		err = fmt.Errorf("response isn't a GetResponseNormal")
		return
	}

	data, err = res.Result.ValueAsData()
	if err != nil {
		access, _ := res.Result.ValueAsAccess()
		err = fmt.Errorf("get failed with result: %s", access.String())
		return
	}

	return
}

func (c *Client) GetWithUnmarshal(att *dlms.AttributeDescriptor, data interface{}) (err error) {
	axdrData, err := c.Get(att)
	if err != nil {
		return
	}

	err = axdr.UnmarshalData(axdrData, data)
	if err != nil {
		err = fmt.Errorf("error unmarshaling data: %w", err)
		return
	}

	return
}
