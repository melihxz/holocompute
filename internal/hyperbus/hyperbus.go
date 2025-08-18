package hyperbus

import (
	"context"
	"crypto/ed25519"
	"fmt"
	"net"

	"github.com/melihxz/holocompute/internal/log"
	"github.com/melihxz/holocompute/pkg/proto"
)

// NodeID represents a unique identifier for a node
type NodeID string

// NodeInfo contains information about a node
type NodeInfo struct {
	ID           NodeID
	Address      net.Addr
	PublicKey    ed25519.PublicKey
	PQPublicKey  []byte
	Capabilities *proto.NodeCapabilities
}

// StreamType represents the type of stream
type StreamType int

const (
	// ControlStream is used for control plane messages
	ControlStream StreamType = iota
	// DataStream is used for data plane messages
	DataStream
)

// Connection represents a connection to a remote node
type Connection interface {
	// NodeID returns the ID of the remote node
	NodeID() NodeID

	// OpenStream opens a new stream of the specified type
	OpenStream(ctx context.Context, streamType StreamType) (Stream, error)

	// Close closes the connection
	Close() error
}

// Stream represents a bidirectional stream
type Stream interface {
	// ReadMessage reads a message from the stream
	ReadMessage(ctx context.Context) ([]byte, error)

	// WriteMessage writes a message to the stream
	WriteMessage(ctx context.Context, data []byte) error

	// Close closes the stream
	Close() error
}

// MessageHandler handles incoming messages
type MessageHandler interface {
	// HandleMessage handles an incoming message
	HandleMessage(ctx context.Context, conn Connection, stream Stream, data []byte) error
}

// Bus represents the hyperbus network layer
type Bus struct {
	localNode   NodeInfo
	connections map[NodeID]Connection
	handler     MessageHandler
	logger      *log.Logger
}

// New creates a new hyperbus
func New(localNode NodeInfo, handler MessageHandler, logger *log.Logger) *Bus {
	return &Bus{
		localNode:   localNode,
		connections: make(map[NodeID]Connection),
		handler:     handler,
		logger:      logger,
	}
}

// LocalNode returns information about the local node
func (b *Bus) LocalNode() NodeInfo {
	return b.localNode
}

// Connect establishes a connection to a remote node
func (b *Bus) Connect(ctx context.Context, node NodeInfo) error {
	// TODO: Implement connection logic
	b.logger.Info("connecting to node", "node_id", node.ID, "address", node.Address)
	return nil
}

// SendControlMessage sends a control message to a specific node
func (b *Bus) SendControlMessage(ctx context.Context, nodeID NodeID, msg []byte) error {
	// Get the connection
	conn, exists := b.connections[nodeID]
	if !exists {
		return fmt.Errorf("no connection to node %s", nodeID)
	}

	// Open a control stream
	stream, err := conn.OpenStream(ctx, ControlStream)
	if err != nil {
		return fmt.Errorf("failed to open control stream: %w", err)
	}
	defer stream.Close()

	// Send the message
	if err := stream.WriteMessage(ctx, msg); err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}

	b.logger.Debug("sent control message", "node_id", nodeID)
	return nil
}

// BroadcastControlMessage sends a control message to all connected nodes
func (b *Bus) BroadcastControlMessage(ctx context.Context, msg []byte) error {
	// TODO: Implement broadcasting control messages
	b.logger.Debug("broadcasting control message")
	return nil
}

// Close closes the hyperbus and all connections
func (b *Bus) Close() error {
	// TODO: Implement closing logic
	b.logger.Info("closing hyperbus")
	return nil
}
