package client

import (
	"fmt"

	"github.com/Circutor/gosem/pkg/axdr"
	"github.com/Circutor/gosem/pkg/dlms"
)

func (c *Client) Get(att *dlms.AttributeDescriptor) (data axdr.DlmsData, err error) {
	req := dlms.CreateGetRequestNormal(unicastInvokeID, *att, nil)

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

func (c *Client) encodeSendReceiveAndDecode(req dlms.CosemPDU) (pdu dlms.CosemPDU, err error) {
	if !c.isAssociated {
		err = fmt.Errorf("client is not associated")
		return
	}

	src, err := req.Encode()
	if err != nil {
		err = fmt.Errorf("error encoding CosemPDU: %w", err)
		return
	}

	out, err := c.transport.Send(src)
	if err != nil {
		err = fmt.Errorf("error sending CosemPDU: %w", err)
		return
	}

	pdu, err = dlms.DecodeCosem(&out)
	if err != nil {
		err = fmt.Errorf("error decoding CosemPDU: %w", err)
		return
	}

	return
}
