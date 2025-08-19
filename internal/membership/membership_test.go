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
	"github.com/stretchr/testify/mock"
)

// MockEventHandler is a mock implementation of EventHandler
type MockEventHandler struct {
	mock.Mock
}

func (m *MockEventHandler) OnMemberJoin(member *Member) {
	m.Called(member)
}

func (m *MockEventHandler) OnMemberLeave(member *Member) {
	m.Called(member)
}

func (m *MockEventHandler) OnMemberStatusChange(member *Member, oldStatus, newStatus MemberStatus) {
	m.Called(member, oldStatus, newStatus)
}

func TestMembership_Join(t *testing.T) {
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

	// Create mock event handler
	mockHandler := &MockEventHandler{}
	membership.AddEventHandler(mockHandler)

	// Create remote member
	remoteMember := &Member{
		ID:           "remote-node",
		Address:      &net.TCPAddr{IP: net.IPv4(127, 0, 0, 2), Port: 8443},
		LastSeen:     time.Now(),
		Status:       Alive,
		Capabilities: &proto.NodeCapabilities{CpuCores: 2, MemoryBytes: 512 * 1024 * 1024},
	}

	// Set up expectations
	mockHandler.On("OnMemberJoin", remoteMember).Return()

	// Join the remote member
	membership.Join(context.TODO(), remoteMember)

	// Verify the member was added
	member, exists := membership.members["remote-node"]
	assert.True(t, exists)
	assert.Equal(t, remoteMember, member)

	// Verify the event handler was called
	mockHandler.AssertExpectations(t)
}

func TestMembership_Leave(t *testing.T) {
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

	// Create mock event handler
	mockHandler := &MockEventHandler{}
	membership.AddEventHandler(mockHandler)

	// Create remote member
	remoteMember := &Member{
		ID:           "remote-node",
		Address:      &net.TCPAddr{IP: net.IPv4(127, 0, 0, 2), Port: 8443},
		LastSeen:     time.Now(),
		Status:       Alive,
		Capabilities: &proto.NodeCapabilities{CpuCores: 2, MemoryBytes: 512 * 1024 * 1024},
	}

	// Add the remote member
	membership.members["remote-node"] = remoteMember

	// Set up expectations
	mockHandler.On("OnMemberLeave", remoteMember).Return()

	// Leave the remote member
	membership.Leave(context.TODO(), "remote-node")

	// Verify the member was removed
	_, exists := membership.members["remote-node"]
	assert.False(t, exists)

	// Verify the event handler was called
	mockHandler.AssertExpectations(t)
}

func TestMembership_UpdateMemberStatus(t *testing.T) {
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

	// Create mock event handler
	mockHandler := &MockEventHandler{}
	membership.AddEventHandler(mockHandler)

	// Create remote member
	remoteMember := &Member{
		ID:           "remote-node",
		Address:      &net.TCPAddr{IP: net.IPv4(127, 0, 0, 2), Port: 8443},
		LastSeen:     time.Now(),
		Status:       Alive,
		Capabilities: &proto.NodeCapabilities{CpuCores: 2, MemoryBytes: 512 * 1024 * 1024},
	}

	// Add the remote member
	membership.members["remote-node"] = remoteMember

	// Set up expectations
	mockHandler.On("OnMemberStatusChange", remoteMember, Alive, Suspect).Return()

	// Update the member status
	membership.UpdateMemberStatus("remote-node", Suspect)

	// Verify the status was updated
	member, exists := membership.members["remote-node"]
	assert.True(t, exists)
	assert.Equal(t, Suspect, member.Status)

	// Verify the event handler was called
	mockHandler.AssertExpectations(t)
}
