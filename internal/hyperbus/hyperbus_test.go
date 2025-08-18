package hyperbus

import (
	"context"
	"crypto/ed25519"
	"log/slog"
	"net"
	"testing"

	"github.com/melihxz/holocompute/internal/log"
	"github.com/melihxz/holocompute/pkg/proto"
	"github.com/stretchr/testify/assert"
)

type mockHandler struct{}

func (m *mockHandler) HandleMessage(ctx context.Context, conn Connection, stream Stream, data []byte) error {
	return nil
}

func TestBus_LocalNode(t *testing.T) {
	logger := log.New(slog.LevelDebug)

	// Create local node info
	publicKey := ed25519.PublicKey("test-public-key")
	capabilities := &proto.NodeCapabilities{
		CpuCores:    4,
		MemoryBytes: 1024 * 1024 * 1024,
		HasGpu:      true,
		Tags:        []string{"test"},
	}

	localNode := NodeInfo{
		ID:           "test-node",
		Address:      &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 8443},
		PublicKey:    publicKey,
		PQPublicKey:  []byte("test-pq-key"),
		Capabilities: capabilities,
	}

	// Create bus
	handler := &mockHandler{}
	bus := New(localNode, handler, logger)

	// Get local node
	returnedNode := bus.LocalNode()

	// Verify
	assert.Equal(t, localNode, returnedNode)
}

func TestBus_Connect(t *testing.T) {
	logger := log.New(slog.LevelDebug)

	// Create local node info
	localNode := NodeInfo{
		ID:        "local-node",
		Address:   &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 8443},
		PublicKey: ed25519.PublicKey("local-public-key"),
	}

	// Create bus
	handler := &mockHandler{}
	bus := New(localNode, handler, logger)

	// Create remote node info
	remoteNode := NodeInfo{
		ID:        "remote-node",
		Address:   &net.TCPAddr{IP: net.IPv4(127, 0, 0, 2), Port: 8443},
		PublicKey: ed25519.PublicKey("remote-public-key"),
	}

	// Connect to remote node (this is a mock, so it should not error)
	err := bus.Connect(nil, remoteNode)
	assert.NoError(t, err)
}

func TestNodeID_String(t *testing.T) {
	nodeID := NodeID("test-node")
	assert.Equal(t, "test-node", string(nodeID))
}
