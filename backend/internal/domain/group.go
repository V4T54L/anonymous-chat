package domain

import (
	"context"
	"errors"
	"math/rand"
	"sync"
	"time"
)

var ErrMemberNotFound = errors.New("member not found in group")

// Group represents a chat group.
type Group struct {
	ID                string
	Name              string
	JoinTag           string // Unique, user-friendly tag to join a group
	ProfilePictureURL string
	OwnerID           string
	Members           map[string]*User // Map of UserID to User
	CreatedAt         time.Time
	mu                sync.RWMutex
}

// NewGroup creates a new group.
func NewGroup(id, name, joinTag, ownerID string) *Group {
	return &Group{
		ID:        id,
		Name:      name,
		JoinTag:   joinTag,
		OwnerID:   ownerID,
		Members:   make(map[string]*User),
		CreatedAt: time.Now().UTC(),
	}
}

// AddMember adds a user to the group.
func (g *Group) AddMember(user *User) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.Members[user.ID] = user
}

// RemoveMember removes a user from the group.
// It also handles re-assigning ownership if the owner leaves.
func (g *Group) RemoveMember(userID string) (newOwnerID string, err error) {
	g.mu.Lock()
	defer g.mu.Unlock()

	if _, ok := g.Members[userID]; !ok {
		return "", ErrMemberNotFound
	}

	delete(g.Members, userID)

	// If the owner left and there are still members, assign a new owner.
	if g.OwnerID == userID && len(g.Members) > 0 {
		var memberIDs []string
		for id := range g.Members {
			memberIDs = append(memberIDs, id)
		}
		// Pick a random new owner
		g.OwnerID = memberIDs[rand.Intn(len(memberIDs))]
		return g.OwnerID, nil
	}

	return "", nil
}

// GetMemberIDs returns a slice of all member IDs.
func (g *Group) GetMemberIDs() []string {
	g.mu.RLock()
	defer g.mu.RUnlock()
	ids := make([]string, 0, len(g.Members))
	for id := range g.Members {
		ids = append(ids, id)
	}
	return ids
}

// IsEmpty checks if the group has any members.
func (g *Group) IsEmpty() bool {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return len(g.Members) == 0
}

// UpdateDetails updates the group's name and profile picture.
func (g *Group) UpdateDetails(name, profilePicURL string) {
	g.mu.Lock()
	defer g.mu.Unlock()
	if name != "" {
		g.Name = name
	}
	if profilePicURL != "" {
		g.ProfilePictureURL = profilePicURL
	}
}

// GroupRepository defines the interface for group persistence.
type GroupRepository interface {
	Add(ctx context.Context, group *Group) error
	GetByID(ctx context.Context, id string) (*Group, error)
	GetByTag(ctx context.Context, tag string) (*Group, error)
	Remove(ctx context.Context, id string) error
	GetAll(ctx context.Context) ([]*Group, error)
	Save(ctx context.Context, group *Group) error // For updating members, owner, etc.
}
