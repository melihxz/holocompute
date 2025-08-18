package hyperbus

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"fmt"
	"math/big"
	"time"

	"github.com/melihxz/holocompute/internal/log"
	"github.com/melihxz/holocompute/pkg/proto"
	"github.com/quic-go/quic-go"
)

// QUICConnection implements the Connection interface using QUIC
type QUICConnection struct {
	nodeID  NodeID
	conn    *quic.Conn
	logger  *log.Logger
	streams map[quic.StreamID]*quic.Stream
}

// NodeID returns the ID of the remote node
func (c *QUICConnection) NodeID() NodeID {
	return c.nodeID
}

// OpenStream opens a new stream of the specified type
func (c *QUICConnection) OpenStream(ctx context.Context, streamType StreamType) (Stream, error) {
	qstream, err := c.conn.OpenStreamSync(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to open QUIC stream: %w", err)
	}

	// Send stream type as first byte
	streamTypeByte := byte(streamType)
	if _, err := qstream.Write([]byte{streamTypeByte}); err != nil {
		qstream.Close()
		return nil, fmt.Errorf("failed to write stream type: %w", err)
	}

	stream := &QUICStream{
		stream: qstream,
		logger: c.logger.With("stream_id", qstream.StreamID()),
	}

	c.streams[qstream.StreamID()] = qstream
	return stream, nil
}

// Close closes the connection
func (c *QUICConnection) Close() error {
	c.logger.Info("closing connection", "node_id", c.nodeID)
	return c.conn.CloseWithError(0, "connection closed")
}

// QUICStream implements the Stream interface using QUIC streams
type QUICStream struct {
	stream *quic.Stream
	logger *log.Logger
}

// ReadMessage reads a message from the stream
func (s *QUICStream) ReadMessage(ctx context.Context) ([]byte, error) {
	// Read the header (6 bytes: 2 for type + 4 for size)
	headerBuf := make([]byte, 6)
	n, err := s.stream.Read(headerBuf)
	if err != nil {
		return nil, err
	}
	if n != 6 {
		return nil, fmt.Errorf("incomplete header read: expected 6 bytes, got %d", n)
	}

	// Decode header to get message size
	header, err := DecodeHeader(headerBuf)
	if err != nil {
		return nil, fmt.Errorf("failed to decode header: %w", err)
	}

	// Read the message body
	bodyBuf := make([]byte, header.Size)
	n, err = s.stream.Read(bodyBuf)
	if err != nil {
		return nil, err
	}
	if uint32(n) != header.Size {
		return nil, fmt.Errorf("incomplete message body read: expected %d bytes, got %d", header.Size, n)
	}

	// Combine header and body
	result := make([]byte, 6+len(bodyBuf))
	copy(result[:6], headerBuf)
	copy(result[6:], bodyBuf)

	return result, nil
}

// WriteMessage writes a message to the stream
func (s *QUICStream) WriteMessage(ctx context.Context, data []byte) error {
	_, err := s.stream.Write(data)
	return err
}

// Close closes the stream
func (s *QUICStream) Close() error {
	s.logger.Debug("closing stream")
	return s.stream.Close()
}

// QUICBus implements the Bus interface using QUIC
type QUICBus struct {
	*Bus
	listener *quic.Listener
}

// NewQUICBus creates a new QUIC-based hyperbus
func NewQUICBus(localNode NodeInfo, handler MessageHandler, logger *log.Logger) (*QUICBus, error) {
	// Generate TLS certificate for QUIC
	tlsConfig, err := generateTLSConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to generate TLS config: %w", err)
	}

	// Create QUIC listener
	addr := localNode.Address.String()
	listener, err := quic.ListenAddr(addr, tlsConfig, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create QUIC listener: %w", err)
	}

	bus := &QUICBus{
		Bus:      New(localNode, handler, logger),
		listener: listener,
	}

	// Start accepting connections
	go bus.acceptLoop()

	return bus, nil
}

// acceptLoop accepts incoming connections
func (b *QUICBus) acceptLoop() {
	for {
		conn, err := b.listener.Accept(context.Background())
		if err != nil {
			b.logger.Error("failed to accept connection", "error", err)
			return
		}

		go b.handleConnection(conn)
	}
}

// handleConnection handles an incoming connection
func (b *QUICBus) handleConnection(conn *quic.Conn) {
	b.logger.Info("handling new connection", "remote_addr", conn.RemoteAddr())

	// Accept the first stream which should be the control stream
	stream, err := conn.AcceptStream(context.Background())
	if err != nil {
		b.logger.Error("failed to accept control stream", "error", err)
		return
	}
	defer stream.Close()

	// Read the stream type
	streamTypeBuf := make([]byte, 1)
	if _, err := stream.Read(streamTypeBuf); err != nil {
		b.logger.Error("failed to read stream type", "error", err)
		return
	}

	streamType := StreamType(streamTypeBuf[0])
	if streamType != ControlStream {
		b.logger.Error("expected control stream", "received_type", streamType)
		return
	}

	// Read the ControlHello message
	// First read the header
	headerBuf := make([]byte, 6) // 2 bytes for type + 4 bytes for size
	if _, err := stream.Read(headerBuf); err != nil {
		b.logger.Error("failed to read message header", "error", err)
		return
	}

	header, err := DecodeHeader(headerBuf)
	if err != nil {
		b.logger.Error("failed to decode message header", "error", err)
		return
	}

	if header.Type != MsgControlHello {
		b.logger.Error("expected ControlHello message", "received_type", header.Type)
		return
	}

	// Read the message body
	bodyBuf := make([]byte, header.Size)
	if _, err := stream.Read(bodyBuf); err != nil {
		b.logger.Error("failed to read message body", "error", err)
		return
	}

	// Decode the ControlHello message
	var hello proto.ControlHello
	if err := DecodeMessage(bodyBuf, &hello); err != nil {
		b.logger.Error("failed to decode ControlHello", "error", err)
		return
	}

	// Create connection wrapper
	qconn := &QUICConnection{
		nodeID:  NodeID(hello.NodeId),
		conn:    conn,
		logger:  b.logger.With("remote_node", hello.NodeId),
		streams: make(map[quic.StreamID]*quic.Stream),
	}

	// Store connection
	b.connections[NodeID(hello.NodeId)] = qconn

	b.logger.Info("established connection with node", "node_id", hello.NodeId)
}

// Connect establishes a connection to a remote node using QUIC
func (b *QUICBus) Connect(ctx context.Context, node NodeInfo) error {
	// Generate TLS config
	tlsConfig, err := generateTLSConfig()
	if err != nil {
		return fmt.Errorf("failed to generate TLS config: %w", err)
	}

	// Connect to remote node
	conn, err := quic.DialAddr(ctx, node.Address.String(), tlsConfig, &quic.Config{})
	if err != nil {
		return fmt.Errorf("failed to dial remote node: %w", err)
	}

	// Create connection wrapper
	qconn := &QUICConnection{
		nodeID:  node.ID,
		conn:    conn,
		logger:  b.logger.With("remote_node", node.ID),
		streams: make(map[quic.StreamID]*quic.Stream),
	}

	// Store connection
	b.connections[node.ID] = qconn

	// Send ControlHello message
	if err := b.sendControlHello(ctx, qconn); err != nil {
		qconn.Close()
		return fmt.Errorf("failed to send ControlHello: %w", err)
	}

	return nil
}

// sendControlHello sends a ControlHello message to establish the connection
func (b *QUICBus) sendControlHello(ctx context.Context, conn *QUICConnection) error {
	// Open control stream
	stream, err := conn.OpenStream(ctx, ControlStream)
	if err != nil {
		return fmt.Errorf("failed to open control stream: %w", err)
	}
	defer stream.Close()

	// Create ControlHello message
	hello := &proto.ControlHello{
		NodeId: string(b.localNode.ID),
		Caps:   b.localNode.Capabilities,
		Pubkey: b.localNode.PublicKey,
	}

	// Encode and send the message
	data, err := EncodeMessage(MsgControlHello, hello)
	if err != nil {
		return fmt.Errorf("failed to encode ControlHello: %w", err)
	}

	if err := stream.WriteMessage(ctx, data); err != nil {
		return fmt.Errorf("failed to send ControlHello: %w", err)
	}

	b.logger.Debug("sent ControlHello", "remote_node", conn.NodeID())
	return nil
}

// generateTLSConfig generates a self-signed TLS certificate for QUIC
func generateTLSConfig() (*tls.Config, error) {
	// Generate key pair
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, err
	}

	// Create certificate template
	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization: []string{"HoloCompute"},
		},
		NotBefore: time.Now(),
		NotAfter:  time.Now().Add(time.Hour * 24 * 365),
		KeyUsage:  x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage: []x509.ExtKeyUsage{
			x509.ExtKeyUsageServerAuth,
			x509.ExtKeyUsageClientAuth,
		},
		BasicConstraintsValid: true,
	}

	// Create self-signed certificate
	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &key.PublicKey, key)
	if err != nil {
		return nil, err
	}

	// Create TLS certificate
	cert := tls.Certificate{
		Certificate: [][]byte{derBytes},
		PrivateKey:  key,
	}

	return &tls.Config{
		Certificates: []tls.Certificate{cert},
		NextProtos:   []string{"holocompute"},
	}, nil
}
