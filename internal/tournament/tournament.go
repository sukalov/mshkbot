package tournament

import (
	"context"
	"fmt"
	"sync"

	"github.com/sukalov/mshkbot/internal/redis"
	"github.com/sukalov/mshkbot/internal/types"
)

type StateManager struct {
	mu     sync.RWMutex
	list   []types.Player
	limit  int
	exists bool
}

type ByTimeAdded []types.Player

func (a ByTimeAdded) Len() int           { return len(a) }
func (a ByTimeAdded) Less(i, j int) bool { return a[i].TimeAdded.Before(a[j].TimeAdded) }
func (a ByTimeAdded) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }

func NewStateManager() *StateManager {
	return &StateManager{
		list:   []types.Player{},
		limit:  26,
		exists: false,
	}
}

func (sm *StateManager) Init() error {
	ctx := context.Background()
	sm.mu.Lock()
	defer sm.mu.Unlock()
	// Retrieve the list from Redis
	list, err := redis.GetList(ctx)
	limit, err2 := redis.GetLimit(ctx)
	if err != nil {
		return err
	}
	if err2 != nil {
		return err2
	}
	sm.list = list
	sm.limit = limit
	return nil
}

func (sm *StateManager) AddPlayer(ctx context.Context, player types.Player) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.list = append(sm.list, player)
	if err := redis.SetList(ctx, sm.list); err != nil {
		fmt.Printf("error happened while adding to redis list: %s", err)
		return err
	}
	return nil
}

func (sm *StateManager) CreateTournament(ctx context.Context) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.exists = true
	if err := redis.SetExists(ctx, true); err != nil {
		fmt.Printf("error happened while saving list state to redis: %s", err)
		return err
	}
	return nil
}

func (sm *StateManager) RemoveTournament(ctx context.Context) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.exists = false
	if err := sm.Clear(ctx); err != nil {
		fmt.Printf("error happened while clearing the redis list: %s", err)
		return err
	}
	if err := redis.SetExists(ctx, false); err != nil {
		fmt.Printf("error happened while saving list state to redis: %s", err)
		return err
	}
	return nil
}

func (sm *StateManager) GetAll() []types.Player {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.list
}

func (sm *StateManager) Clear(ctx context.Context) error {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	sm.list = []types.Player{}
	if err := redis.SetList(ctx, sm.list); err != nil {
		fmt.Printf("error happened while clearing the redis list: %s", err)
	}
	return nil
}

func (sm *StateManager) EditState(ctx context.Context, stateID int, newState types.Player) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	for i, state := range sm.list {
		if state.ID == stateID {
			sm.list[i] = newState
			if err := redis.SetList(ctx, sm.list); err != nil {
				fmt.Printf("error happened while updating the redis list: %s", err)
				return err
			}
			return nil
		}
	}

	return fmt.Errorf("state with ID %d not found", stateID)
}

func (sm *StateManager) Sync(ctx context.Context) error {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	if err := redis.SetList(ctx, sm.list); err != nil {
		fmt.Printf("error happened while updating the redis list: %s", err)
		return err
	}
	return nil
}

func (sm *StateManager) RemoveState(ctx context.Context, stateID int) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	result := []types.Player{}
	for _, state := range sm.list {
		if state.ID != stateID {
			result = append(result, state)
		}
	}
	sm.list = result
	if err := redis.SetList(ctx, result); err != nil {
		fmt.Printf("error happened while updating the redis list: %s", err)
		return err
	}
	return nil
}

func (sm *StateManager) GetLimit() int {
	return sm.limit
}

func (sm *StateManager) SetLimit(ctx context.Context, limit int) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.limit = limit
	if err := redis.SetLimit(ctx, limit); err != nil {
		fmt.Printf("error happened while updating the redis limit: %s", err)
		return err
	}
	return nil
}
