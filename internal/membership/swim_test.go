package membership

import (
	"context"
	"log/slog"
	"net"
	"testing"
	"time"

	"github.com/melihxz/holocompute/internal/log"
	"github.com/melihxz/holocompute/pkg/proto"
	"github.com/stretchr/testify/assert"
)

func TestSWIM_Gossip(t *testing.T) {
	logger := log.New(slog.LevelDebug)

	// Create local member
	localMember := &Member{
		ID:           "local-node",
		Address:      &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 8443},
		LastSeen:     time.Now(),
		Status:       Alive,
		Capabilities: &proto.NodeCapabilities{CpuCores: 4, MemoryBytes: 1024 * 1024 * 1024},
	}

	// Create membership manager
	membership := NewMembership(localMember, logger)

	// Create SWIM instance
	config := DefaultSWIMConfig()
	swim := NewSWIM(membership, nil, config, logger)

	// Test gossip with no members
	swim.gossip(context.Background())

	// Add a remote member
	remoteMember := &Member{
		ID:           "remote-node",
		Address:      &net.TCPAddr{IP: net.IPv4(127, 0, 0, 2), Port: 8443},
		LastSeen:     time.Now(),
		Status:       Alive,
		Capabilities: &proto.NodeCapabilities{CpuCores: 2, MemoryBytes: 512 * 1024 * 1024},
	}

	membership.Join(context.Background(), remoteMember)

	// Test gossip with members
	swim.gossip(context.Background())
}

func TestSWIM_SuspectHandling(t *testing.T) {
	logger := log.New(slog.LevelDebug)

	// Create local member
	localMember := &Member{
		ID:           "local-node",
		Address:      &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 8443},
		LastSeen:     time.Now(),
		Status:       Alive,
		Capabilities: &proto.NodeCapabilities{CpuCores: 4, MemoryBytes: 1024 * 1024 * 1024},
	}

	// Create membership manager
	membership := NewMembership(localMember, logger)

	// Create SWIM instance
	config := DefaultSWIMConfig()
	config.SuspectPeriod = time.Millisecond * 10 // Short period for testing
	swim := NewSWIM(membership, nil, config, logger)

	// Add a remote member with suspect status
	remoteMember := &Member{
		ID:           "remote-node",
		Address:      &net.TCPAddr{IP: net.IPv4(127, 0, 0, 2), Port: 8443},
		LastSeen:     time.Now().Add(-time.Millisecond * 20), // Old last seen time
		Status:       Suspect,
		Capabilities: &proto.NodeCapabilities{CpuCores: 2, MemoryBytes: 512 * 1024 * 1024},
	}

	membership.Join(context.Background(), remoteMember)

	// Test suspect timeout handling
	time.Sleep(time.Millisecond * 15) // Wait for suspect timeout
	swim.checkSuspects()

	// Verify member is now dead
	member, exists := membership.Members()["remote-node"]
	assert.True(t, exists)
	assert.Equal(t, Dead, member.Status)
}
