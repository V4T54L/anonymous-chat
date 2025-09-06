package domain

import (
	"context"
	"sync"
	"time"
)

// User represents a user in the system.
type User struct {
	ID               string
	DisplayName      string
	ProfilePictureURL string
	PublicKey        string    // User's public identity key for E2EE
	LastSeen         time.Time
	mu               sync.RWMutex
}

// NewUser creates a new user instance.
func NewUser(id, displayName, publicKey string) *User {
	return &User{
		ID:          id,
		DisplayName: displayName,
		PublicKey:   publicKey,
		LastSeen:    time.Now().UTC(),
	}
}

// UpdateProfile updates the user's profile information.
func (u *User) UpdateProfile(displayName, profilePicURL string) {
	u.mu.Lock()
	defer u.mu.Unlock()
	if displayName != "" {
		u.DisplayName = displayName
	}
	if profilePicURL != "" {
		u.ProfilePictureURL = profilePicURL
	}
}

// UserRepository defines the interface for user persistence.
// This allows us to swap implementations (e.g., in-memory vs. Redis).
type UserRepository interface {
	Add(ctx context.Context, user *User) error
	GetByID(ctx context.Context, id string) (*User, error)
	Remove(ctx context.Context, id string) error
	GetAll(ctx context.Context) ([]*User, error)
}
