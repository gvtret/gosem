package hdlc

import (
	"encoding/binary"
	"fmt"
	"log"
	"sync"
	"time"

	"gitlab.com/circutor-library/gosem/pkg/dlms"
)

const (
	maxInfoFieldLength = 512
	minInfoFieldLength = 32

	startAndEndFlag                = 0x7E
	frameFormatWithoutSegmentation = 0xA000
	frameFormatWithSegmentation    = 0xA800

	controlSNRM = 0x93
	controlUA   = 0x73
)

type ReceivedFrame struct {
	UpperAddress  int
	LowerAddress  int
	ClientAddress int
	Control       uint8
	IsSegmented   bool
	Data          []byte
	HCS           uint16
	FCS           uint16
}

type hdlc struct {
	maxInfoFieldLengthSend int
	upperAddress           int
	lowerAddress           int
	clientAddress          int
	replyTimeout           time.Duration
	rrr                    int
	sss                    int
	fcsTable               [256]uint16
	transport              dlms.Transport
	dc                     dlms.DataChannel
	tc                     dlms.DataChannel
	fc                     chan *ReceivedFrame
	logger                 *log.Logger
	mutex                  sync.Mutex
}

func New(transport dlms.Transport, address int, client int, server int) dlms.Transport {
	h := &hdlc{
		maxInfoFieldLengthSend: maxInfoFieldLength,
		upperAddress:           server,
		lowerAddress:           address,
		clientAddress:          client,
		replyTimeout:           2 * time.Second,
		rrr:                    0,
		sss:                    0,
		fcsTable:               generateFCSTable(),
		transport:              transport,
		dc:                     nil,
		tc:                     make(dlms.DataChannel, 10),
		fc:                     nil,
		logger:                 nil,
		mutex:                  sync.Mutex{},
	}

	transport.SetReception(h.tc)

	go h.manager()

	return h
}

func (h *hdlc) Close() {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	h.transport.Close()
	if h.dc != nil {
		close(h.dc)
		h.dc = nil
	}
}

func (h *hdlc) Connect() error {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	if err := h.transport.Connect(); err != nil {
		return err
	}

	h.maxInfoFieldLengthSend = maxInfoFieldLength
	h.rrr = 0
	h.sss = 0

	frameToSend := h.createFrame(controlSNRM, false, nil)

	rf, err := h.sendReceive(frameToSend)
	if err != nil {
		return fmt.Errorf("send error: %w", err)
	}

	err = h.handleConnectReply(rf)
	if err != nil {
		h.transport.Disconnect()
		return fmt.Errorf("connect error: %w", err)
	}

	return nil
}

func (h *hdlc) handleConnectReply(rf *ReceivedFrame) error {
	if rf.Control != controlUA {
		return fmt.Errorf("invalid control byte, have %02X, expected %02X", rf.Control, controlUA)
	}

	data := rf.Data

	if len(data) < 3 {
		return fmt.Errorf("invalid UA data, have %d", len(data))
	}

	if data[0] != 0x81 || data[1] != 0x80 || data[2] != byte(len(data)-3) {
		return fmt.Errorf("invalid UA data, have %02X:%02X:%02X", data[0], data[1], data[2])
	}

	data = data[3:]

	for {
		if len(data) < 2 {
			break
		}

		code := data[0]
		length := int(data[1])

		if len(data) < length+2 {
			return fmt.Errorf("invalid UA data, have %d", len(data))
		}

		if code == 0x06 {
			var maxInfoFieldLengthSend int

			switch length {
			case 0x01:
				maxInfoFieldLengthSend = int(data[2])
			case 0x02:
				maxInfoFieldLengthSend = int(binary.BigEndian.Uint16(data[2:]))
			default:
				return fmt.Errorf("invalid UA data, have %d", length)
			}

			if maxInfoFieldLengthSend > maxInfoFieldLength {
				maxInfoFieldLengthSend = maxInfoFieldLength
			}
			if maxInfoFieldLengthSend < minInfoFieldLength {
				maxInfoFieldLengthSend = minInfoFieldLength
			}
		}

		data = data[length+2:]
	}

	return nil
}

func (h *hdlc) Disconnect() error {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	return h.transport.Disconnect()
}

func (h *hdlc) IsConnected() bool {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	return h.transport.IsConnected()
}

func (h *hdlc) SetAddress(client int, server int) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	h.clientAddress = client
	h.upperAddress = server
}

func (h *hdlc) SetReception(dc dlms.DataChannel) {
	h.dc = dc
}

func (h *hdlc) Send(_ []byte) error {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	if !h.transport.IsConnected() {
		return fmt.Errorf("not connected")
	}

	return nil
}

func (h *hdlc) SetLogger(logger *log.Logger) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	h.logger = logger
	h.transport.SetLogger(logger)
}

func (h *hdlc) manager() {
	frame := make([]byte, 0)

	for {
		data, ok := <-h.tc
		if !ok {
			return
		}

		frame = append(frame, data...)

		for {
			if len(frame) == 0 {
				break
			}

			rf := h.searchFrame(&frame)

			if rf != nil && h.fc != nil {
				h.fc <- rf
			}
		}
	}
}

func (h *hdlc) createFrame(control uint8, segmented bool, data []byte) []byte {
	frame := make([]byte, 0, 8+len(data))

	// Starting flag
	frame = append(frame, startAndEndFlag)

	// Frame length and segmentation
	lenAndSeg := frameFormatWithoutSegmentation

	if segmented {
		lenAndSeg = frameFormatWithSegmentation
	}

	if data != nil {
		lenAndSeg |= 10 + len(data)
	} else {
		lenAndSeg |= 8
	}

	frame = append(frame, byte(lenAndSeg>>8))
	frame = append(frame, byte(lenAndSeg))

	// Destination address
	frame = append(frame, byte(h.upperAddress<<1))
	frame = append(frame, byte(h.lowerAddress<<1)|0x01)

	// Source address
	frame = append(frame, byte(h.clientAddress<<1)|0x01)

	// Control byte
	frame = append(frame, control)

	// HCS
	checksum := h.chksum(frame[1:])
	frame = append(frame, byte(checksum))
	frame = append(frame, byte(checksum>>8))

	// Data
	if data != nil {
		frame = append(frame, data...)

		// FCS
		checksum = h.chksum(frame[1:])
		frame = append(frame, byte(checksum))
		frame = append(frame, byte(checksum>>8))
	}

	// Closing flag
	frame = append(frame, startAndEndFlag)

	return frame
}

func generateFCSTable() [256]uint16 {
	var table [256]uint16
	for i := 0; i < 256; i++ {
		crc := uint16(i)
		for j := 0; j < 8; j++ {
			if crc&1 != 0 {
				crc = (crc >> 1) ^ 0x8408
			} else {
				crc >>= 1
			}
		}
		table[i] = crc
	}

	return table
}

func (h *hdlc) chksum(data []byte) uint16 {
	fcs := uint16(0xFFFF)

	for _, b := range data {
		fcs = (fcs >> 8) ^ h.fcsTable[(fcs^uint16(b))&0xFF]
	}

	return fcs ^ 0xFFFF
}

func (h *hdlc) searchFrame(frame *[]byte) *ReceivedFrame {
	// Search for the start flag
	isStartIndexFound := false
	for i, b := range *frame {
		if b == startAndEndFlag {
			*frame = (*frame)[i:]
			isStartIndexFound = true
			break
		}
	}

	// If no start flag is found, return nil and flush the buffer
	if !isStartIndexFound {
		*frame = nil
		return nil
	}

	// Check minimum frame length
	if len(*frame) < 9 {
		return nil
	}

	// Check if the frame is long enough
	length := int(binary.BigEndian.Uint16((*frame)[1:]) & 0x07FF)
	if len(*frame) < length+2 {
		return nil
	}

	// Check end flag, if it is not found, return nil and remove the start flag
	if (*frame)[length+1] != startAndEndFlag {
		*frame = (*frame)[1:]
		return nil
	}

	// Remove the frame from the buffer
	src := (*frame)[:length+2]
	*frame = (*frame)[length+2:]

	// Parse received frame
	receivedFrame, err := h.parseFrame(src)
	if err != nil {
		if h.logger != nil {
			h.logger.Printf("Invalid received data: %e", err)
		}

		return nil
	}

	return receivedFrame
}

func (h *hdlc) parseFrame(src []byte) (*ReceivedFrame, error) {
	if len(src) < 9 {
		return nil, fmt.Errorf("frame too short, have %d", len(src))
	}

	if (src[1] & 0xF0) != 0xA0 {
		return nil, fmt.Errorf("invalid frame format, have %02X", src[1])
	}

	isSegmented := (src[1] & 0x08) == 0x80

	clientAddress := int(src[3]) >> 1
	if clientAddress != h.clientAddress {
		return nil, fmt.Errorf("invalid client address, have %d, expected %d", clientAddress, h.clientAddress)
	}

	upperAddress := int(src[4]) >> 1
	lowerAddress := int(src[5]) >> 1

	if upperAddress != h.upperAddress || lowerAddress != h.lowerAddress {
		return nil, fmt.Errorf("invalid source address, have %d:%d", upperAddress, lowerAddress)
	}

	control := src[6]
	hcs := binary.LittleEndian.Uint16(src[7:])
	calculatedHCS := h.chksum(src[1:7])
	if hcs != calculatedHCS {
		return nil, fmt.Errorf("HCS error, have %04X, expected %04X", hcs, calculatedHCS)
	}

	var data []byte
	var fcs uint16

	if len(src) > 9 {
		data = src[9 : len(src)-3]
		fcs = binary.LittleEndian.Uint16(src[len(src)-3:])
		calculatedFCS := h.chksum(src[1 : len(src)-3])
		if fcs != calculatedFCS {
			return nil, fmt.Errorf("FCS error, have %04X, expected %04X", fcs, calculatedFCS)
		}
	} else {
		data = nil
		fcs = hcs
	}

	receivedFrame := &ReceivedFrame{
		UpperAddress:  upperAddress,
		LowerAddress:  lowerAddress,
		ClientAddress: clientAddress,
		Control:       control,
		IsSegmented:   isSegmented,
		Data:          data,
		HCS:           hcs,
		FCS:           fcs,
	}

	return receivedFrame, nil
}

func (h *hdlc) sendReceive(src []byte) (*ReceivedFrame, error) {
	h.fc = make(chan *ReceivedFrame, 1)
	defer func() { h.fc = nil }()

	err := h.transport.Send(src)
	if err != nil {
		return nil, fmt.Errorf("send error: %w", err)
	}

	// Wait for the device response
	timeout := time.NewTimer(h.replyTimeout)
	defer timeout.Stop()

	select {
	case data := <-h.fc:
		return data, nil
	case <-timeout.C:
		return nil, fmt.Errorf("timeout waiting for response")
	}
}
