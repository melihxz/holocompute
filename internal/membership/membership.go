package membership

import (
	"context"
	"net"
	"time"

	"github.com/melihxz/holocompute/internal/hyperbus"
	"github.com/melihxz/holocompute/internal/log"
	"github.com/melihxz/holocompute/pkg/proto"
)

// Member represents a cluster member
type Member struct {
	ID           hyperbus.NodeID
	Address      net.Addr
	LastSeen     time.Time
	Status       MemberStatus
	Capabilities *proto.NodeCapabilities
}

// MemberStatus represents the status of a member
type MemberStatus int

const (
	// Alive means the member is alive and healthy
	Alive MemberStatus = iota
	// Suspect means the member is suspected to be dead
	Suspect
	// Dead means the member is confirmed dead
	Dead
)

// Membership manages cluster membership using SWIM protocol
type Membership struct {
	localMember   *Member
	members       map[hyperbus.NodeID]*Member
	eventHandlers []EventHandler
	logger        *log.Logger
}

// EventHandler handles membership events
type EventHandler interface {
	// OnMemberJoin is called when a member joins the cluster
	OnMemberJoin(member *Member)

	// OnMemberLeave is called when a member leaves the cluster
	OnMemberLeave(member *Member)

	// OnMemberStatusChange is called when a member's status changes
	OnMemberStatusChange(member *Member, oldStatus, newStatus MemberStatus)
}

// NewMembership creates a new membership manager
func NewMembership(localMember *Member, logger *log.Logger) *Membership {
	return &Membership{
		localMember: localMember,
		members:     make(map[hyperbus.NodeID]*Member),
		logger:      logger,
	}
}

// LocalMember returns the local member
func (m *Membership) LocalMember() *Member {
	return m.localMember
}

// Members returns all known members
func (m *Membership) Members() map[hyperbus.NodeID]*Member {
	return m.members
}

// AddEventHandler adds an event handler
func (m *Membership) AddEventHandler(handler EventHandler) {
	m.eventHandlers = append(m.eventHandlers, handler)
}

// Join adds a member to the cluster
func (m *Membership) Join(ctx context.Context, member *Member) {
	m.logger.Info("member joining", "member_id", member.ID)

	oldMember, exists := m.members[member.ID]
	m.members[member.ID] = member

	if !exists {
		// New member
		for _, handler := range m.eventHandlers {
			handler.OnMemberJoin(member)
		}
	} else {
		// Existing member status update
		if oldMember.Status != member.Status {
			for _, handler := range m.eventHandlers {
				handler.OnMemberStatusChange(member, oldMember.Status, member.Status)
			}
		}
	}
}

// Leave removes a member from the cluster
func (m *Membership) Leave(ctx context.Context, memberID hyperbus.NodeID) {
	member, exists := m.members[memberID]
	if !exists {
		return
	}

	m.logger.Info("member leaving", "member_id", memberID)
	delete(m.members, memberID)

	for _, handler := range m.eventHandlers {
		handler.OnMemberLeave(member)
	}
}

// UpdateMemberStatus updates the status of a member
func (m *Membership) UpdateMemberStatus(memberID hyperbus.NodeID, status MemberStatus) {
	member, exists := m.members[memberID]
	if !exists {
		return
	}

	oldStatus := member.Status
	if oldStatus == status {
		return
	}

	member.Status = status
	member.LastSeen = time.Now()

	m.logger.Debug("member status updated",
		"member_id", memberID,
		"old_status", oldStatus,
		"new_status", status)

	for _, handler := range m.eventHandlers {
		handler.OnMemberStatusChange(member, oldStatus, status)
	}
}
