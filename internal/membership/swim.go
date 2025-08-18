package membership

import (
	"context"
	"math/rand"
	"time"

	"github.com/melihxz/holocompute/internal/hyperbus"
	"github.com/melihxz/holocompute/internal/log"
)

// SWIM implements the SWIM gossip protocol
type SWIM struct {
	*Membership
	bus           *hyperbus.Bus
	gossipPeriod  time.Duration
	suspectPeriod time.Duration
	logger        *log.Logger
	cancel        context.CancelFunc
}

// SWIMConfig contains configuration for SWIM
type SWIMConfig struct {
	GossipPeriod  time.Duration
	SuspectPeriod time.Duration
}

// DefaultSWIMConfig returns the default SWIM configuration
func DefaultSWIMConfig() SWIMConfig {
	return SWIMConfig{
		GossipPeriod:  time.Second,
		SuspectPeriod: 5 * time.Second,
	}
}

// NewSWIM creates a new SWIM instance
func NewSWIM(membership *Membership, bus *hyperbus.Bus, config SWIMConfig, logger *log.Logger) *SWIM {
	return &SWIM{
		Membership:    membership,
		bus:           bus,
		gossipPeriod:  config.GossipPeriod,
		suspectPeriod: config.SuspectPeriod,
		logger:        logger,
	}
}

// Start starts the SWIM protocol
func (s *SWIM) Start(ctx context.Context) {
	ctx, s.cancel = context.WithCancel(ctx)

	// Start gossip loop
	go s.gossipLoop(ctx)

	// Start suspect timeout loop
	go s.suspectLoop(ctx)
}

// Stop stops the SWIM protocol
func (s *SWIM) Stop() {
	if s.cancel != nil {
		s.cancel()
	}
}

// gossipLoop periodically gossips with random members
func (s *SWIM) gossipLoop(ctx context.Context) {
	ticker := time.NewTicker(s.gossipPeriod)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.gossip(ctx)
		}
	}
}

// gossip exchanges membership information with a random member
func (s *SWIM) gossip(ctx context.Context) {
	// Get all alive members except ourselves
	members := make([]*Member, 0, len(s.members))
	for _, member := range s.members {
		if member.ID != s.localMember.ID && member.Status == Alive {
			members = append(members, member)
		}
	}

	if len(members) == 0 {
		return
	}

	// Select a random member to gossip with
	target := members[rand.Intn(len(members))]

	// In a real implementation, we would:
	// 1. Create a gossip message with our membership information
	// 2. Send it to the target member
	// 3. Wait for a response
	// 4. Update our membership based on the response

	s.logger.Debug("gossiping with member", "target_id", target.ID)
}

// suspectLoop handles suspect timeouts
func (s *SWIM) suspectLoop(ctx context.Context) {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.checkSuspects()
		}
	}
}

// checkSuspects checks if any suspects have timed out
func (s *SWIM) checkSuspects() {
	now := time.Now()

	for _, member := range s.members {
		if member.Status == Suspect && now.Sub(member.LastSeen) > s.suspectPeriod {
			// Suspect timeout, mark as dead
			s.UpdateMemberStatus(member.ID, Dead)
		}
	}
}

// OnMemberJoin handles member join events
func (s *SWIM) OnMemberJoin(member *Member) {
	// When a member joins, we might want to do some initialization
	s.logger.Info("member joined", "member_id", member.ID)
}

// OnMemberLeave handles member leave events
func (s *SWIM) OnMemberLeave(member *Member) {
	// When a member leaves, we might want to clean up resources
	s.logger.Info("member left", "member_id", member.ID)
}

// OnMemberStatusChange handles member status change events
func (s *SWIM) OnMemberStatusChange(member *Member, oldStatus, newStatus MemberStatus) {
	s.logger.Debug("member status changed",
		"member_id", member.ID,
		"old_status", oldStatus,
		"new_status", newStatus)
}
