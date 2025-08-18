package hyperbus

import (
	"bytes"
	"encoding/binary"
	"fmt"
	
	"google.golang.org/protobuf/proto"
)

// MessageType identifies the type of message
type MessageType uint16

const (
	// Control messages
	MsgControlHello MessageType = iota
	MsgClusterState
	MsgLeaseRequest
	MsgLeaseGrant
	
	// Data messages
	MsgPageRequest
	MsgPageResponse
	MsgTaskSubmit
	MsgTaskResult
)

// MessageHeader is the header for all messages
type MessageHeader struct {
	Type MessageType
	Size uint32
}

// EncodeMessage encodes a protobuf message with header
func EncodeMessage(msgType MessageType, pb proto.Message) ([]byte, error) {
	// Serialize the protobuf message
	data, err := proto.Marshal(pb)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal protobuf: %w", err)
	}
	
	// Create header
	header := MessageHeader{
		Type: msgType,
		Size: uint32(len(data)),
	}
	
	// Encode header and message
	buf := new(bytes.Buffer)
	if err := binary.Write(buf, binary.BigEndian, header); err != nil {
		return nil, fmt.Errorf("failed to write header: %w", err)
	}
	
	if _, err := buf.Write(data); err != nil {
		return nil, fmt.Errorf("failed to write message: %w", err)
	}
	
	return buf.Bytes(), nil
}

// DecodeHeader decodes a message header
func DecodeHeader(data []byte) (MessageHeader, error) {
	var header MessageHeader
	buf := bytes.NewReader(data)
	if err := binary.Read(buf, binary.BigEndian, &header); err != nil {
		return header, fmt.Errorf("failed to read header: %w", err)
	}
	return header, nil
}

// DecodeMessage decodes a protobuf message
func DecodeMessage(data []byte, pb proto.Message) error {
	if err := proto.Unmarshal(data, pb); err != nil {
		return fmt.Errorf("failed to unmarshal protobuf: %w", err)
	}
	return nil
}