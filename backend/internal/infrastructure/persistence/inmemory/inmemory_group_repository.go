package inmemory

import (
	"context"
	"fmt"
	"sync"
	"chat-app/server/internal/domain"
)

// InMemoryGroupRepository is an in-memory implementation of GroupRepository.
type InMemoryGroupRepository struct {
	groups  map[string]*domain.Group
	tags    map[string]string // joinTag -> groupID
	mu      sync.RWMutex
}

// NewInMemoryGroupRepository creates a new in-memory group repository.
func NewInMemoryGroupRepository() *InMemoryGroupRepository {
	return &InMemoryGroupRepository{
		groups: make(map[string]*domain.Group),
		tags:   make(map[string]string),
	}
}

func (r *InMemoryGroupRepository) Add(ctx context.Context, group *domain.Group) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.groups[group.ID]; ok {
		return fmt.Errorf("group with ID %s already exists", group.ID)
	}
	if _, ok := r.tags[group.JoinTag]; ok {
		return fmt.Errorf("group with tag %s already exists", group.JoinTag)
	}
	r.groups[group.ID] = group
	r.tags[group.JoinTag] = group.ID
	return nil
}

func (r *InMemoryGroupRepository) GetByID(ctx context.Context, id string) (*domain.Group, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	group, ok := r.groups[id]
	if !ok {
		return nil, fmt.Errorf("group with ID %s not found", id)
	}
	return group, nil
}

func (r *InMemoryGroupRepository) GetByTag(ctx context.Context, tag string) (*domain.Group, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	groupID, ok := r.tags[tag]
	if !ok {
		return nil, fmt.Errorf("group with tag %s not found", tag)
	}
	return r.groups[groupID], nil
}

func (r *InMemoryGroupRepository) Remove(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	group, ok := r.groups[id]
	if !ok {
		return fmt.Errorf("group with ID %s not found", id)
	}
	delete(r.tags, group.JoinTag)
	delete(r.groups, id)
	return nil
}

func (r *InMemoryGroupRepository) GetAll(ctx context.Context) ([]*domain.Group, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	groups := make([]*domain.Group, 0, len(r.groups))
	for _, group := range r.groups {
		groups = append(groups, group)
	}
	return groups, nil
}

func (r *InMemoryGroupRepository) Save(ctx context.Context, group *domain.Group) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	// In-memory, the object is already updated via its pointer.
	// This check just ensures it exists. A real DB would perform an UPDATE here.
	if _, ok := r.groups[group.ID]; !ok {
		return fmt.Errorf("cannot save group with ID %s: not found", group.ID)
	}
	// No-op for in-memory, but crucial for other implementations.
	return nil
}
