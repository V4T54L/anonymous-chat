package application

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"chat-app/server/internal/domain"
	"github.com/google/uuid"
)

var (
	ErrUserNotFound  = errors.New("user not found")
	ErrGroupNotFound = errors.New("group not found")
	ErrAlreadyExists = errors.New("already exists")
)

// ChatService handles the core application logic (use cases).
type ChatService struct {
	userRepo  domain.UserRepository
	groupRepo domain.GroupRepository
}

// NewChatService creates a new ChatService.
func NewChatService(userRepo domain.UserRepository, groupRepo domain.GroupRepository) *ChatService {
	return &ChatService{
		userRepo:  userRepo,
		groupRepo: groupRepo,
	}
}

// RegisterUser creates or retrieves a user.
func (s *ChatService) RegisterUser(ctx context.Context, userID, displayName, publicKey string) (*domain.User, error) {
	// In this ephemeral system, we just add the user. A real system might check for existence.
	user, err := s.userRepo.GetByID(ctx, userID)
	if err == nil {
		return user, nil // User already connected in another session, which is fine.
	}

	newUser := domain.NewUser(userID, displayName, publicKey)
	if err := s.userRepo.Add(ctx, newUser); err != nil {
		return nil, fmt.Errorf("failed to add user: %w", err)
	}
	return newUser, nil
}

// GetUser retrieves a user by their ID.
func (s *ChatService) GetUser(ctx context.Context, userID string) (*domain.User, error) {
    user, err := s.userRepo.GetByID(ctx, userID)
    if err != nil {
        return nil, ErrUserNotFound
    }
    return user, nil
}


// UnregisterUser removes a user from the system.
func (s *ChatService) UnregisterUser(ctx context.Context, userID string) error {
	return s.userRepo.Remove(ctx, userID)
}

// CreateGroup creates a new group and adds the creator as the first member.
func (s *ChatService) CreateGroup(ctx context.Context, name, joinTag, ownerID string) (*domain.Group, error) {
	owner, err := s.userRepo.GetByID(ctx, ownerID)
	if err != nil {
		return nil, ErrUserNotFound
	}
	
	// Check if join tag is unique
	if _, err := s.groupRepo.GetByTag(ctx, joinTag); err == nil {
		return nil, fmt.Errorf("join tag '%s' is already in use", joinTag)
	}

	groupID := uuid.New().String()
	group := domain.NewGroup(groupID, name, joinTag, owner.ID)
	group.AddMember(owner)

	if err := s.groupRepo.Add(ctx, group); err != nil {
		return nil, fmt.Errorf("failed to create group: %w", err)
	}
	return group, nil
}

// JoinGroup adds a user to an existing group.
func (s *ChatService) JoinGroup(ctx context.Context, groupID, userID string) (*domain.Group, error) {
	group, err := s.groupRepo.GetByID(ctx, groupID)
	if err != nil {
		return nil, ErrGroupNotFound
	}

	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, ErrUserNotFound
	}

	group.AddMember(user)
	if err := s.groupRepo.Save(ctx, group); err != nil {
		return nil, fmt.Errorf("failed to save group after joining: %w", err)
	}
	return group, nil
}

// LeaveGroup removes a user from a group.
func (s *ChatService) LeaveGroup(ctx context.Context, groupID, userID string) (*domain.Group, string, error) {
	group, err := s.groupRepo.GetByID(ctx, groupID)
	if err != nil {
		return nil, "", ErrGroupNotFound
	}

	newOwnerID, err := group.RemoveMember(userID)
	if err != nil {
		return nil, "", fmt.Errorf("failed to remove member: %w", err)
	}

	if group.IsEmpty() {
		// The hub will handle scheduling deletion after a timeout
		return group, newOwnerID, nil
	}
	
	if err := s.groupRepo.Save(ctx, group); err != nil {
		return nil, "", fmt.Errorf("failed to save group after leaving: %w", err)
	}
	return group, newOwnerID, nil
}

// FindGroupsByTag performs a simple search for groups by their join tag.
func (s *ChatService) FindGroupsByTag(ctx context.Context, tagQuery string) ([]*domain.Group, error) {
	allGroups, err := s.groupRepo.GetAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not retrieve groups: %w", err)
	}

	var matchedGroups []*domain.Group
	for _, group := range allGroups {
		// Simple exact match for this implementation
		if strings.EqualFold(group.JoinTag, tagQuery) {
			matchedGroups = append(matchedGroups, group)
		}
	}
	return matchedGroups, nil
}

// UpdateGroupDetails updates a group's info.
func (s *ChatService) UpdateGroupDetails(ctx context.Context, groupID, name, profilePicURL string) (*domain.Group, error) {
    group, err := s.groupRepo.GetByID(ctx, groupID)
    if err != nil {
        return nil, ErrGroupNotFound
    }
    group.UpdateDetails(name, profilePicURL)
    if err := s.groupRepo.Save(ctx, group); err != nil {
        return nil, fmt.Errorf("failed to save group details: %w", err)
    }
    return group, nil
}

// UpdateUserProfile updates a user's profile.
func (s *ChatService) UpdateUserProfile(ctx context.Context, userID, displayName, profilePicURL string) (*domain.User, error) {
    user, err := s.userRepo.GetByID(ctx, userID)
    if err != nil {
        return nil, ErrUserNotFound
    }
    user.UpdateProfile(displayName, profilePicURL)
    // The user object is a pointer, so the in-memory repo is updated directly.
    // A DB-backed repo would need a `Save` call here.
    return user, nil
}
