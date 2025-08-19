package hyperbus

import (
	"testing"

	"github.com/melihxz/holocompute/pkg/proto"
	"github.com/stretchr/testify/assert"
)

func TestQUICBusCommunication(t *testing.T) {
	// This test requires actual network communication, so we'll skip it for now
	// A complete implementation would:
	// 1. Start a server node
	// 2. Start a client node
	// 3. Have them connect and exchange messages
	// 4. Verify the communication works

	t.Skip("Skipping network communication test - requires actual QUIC setup")
}

func TestMessageEncoding(t *testing.T) {
	// Create a test ControlHello message
	hello := &proto.ControlHello{
		NodeId: "test-node",
		Caps: &proto.NodeCapabilities{
			CpuCores:    4,
			MemoryBytes: 1024 * 1024 * 1024,
			HasGpu:      true,
		},
		Pubkey: []byte("test-public-key"),
	}

	// Encode the message
	data, err := EncodeMessage(MsgControlHello, hello)
	assert.NoError(t, err)
	assert.NotNil(t, data)
	assert.Greater(t, len(data), 6) // At least header size

	// Decode the header
	header, err := DecodeHeader(data[:6])
	assert.NoError(t, err)
	assert.Equal(t, MsgControlHello, header.Type)
	assert.Equal(t, uint32(len(data)-6), header.Size)

	// Decode the message
	var decoded proto.ControlHello
	err = DecodeMessage(data[6:], &decoded)
	assert.NoError(t, err)
	assert.Equal(t, hello.NodeId, decoded.NodeId)
	assert.Equal(t, hello.Caps.CpuCores, decoded.Caps.CpuCores)
	assert.Equal(t, hello.Pubkey, decoded.Pubkey)
}
