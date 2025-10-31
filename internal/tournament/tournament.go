package tournament

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/sukalov/mshkbot/internal/redis"
	"github.com/sukalov/mshkbot/internal/types"
)

type TournamentManager struct {
	mu     sync.RWMutex
	List   []types.Player
	Limit  int
	Exists bool
}

type ByTimeAdded []types.Player

func (a ByTimeAdded) Len() int           { return len(a) }
func (a ByTimeAdded) Less(i, j int) bool { return a[i].TimeAdded.Before(a[j].TimeAdded) }
func (a ByTimeAdded) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }

func (tm *TournamentManager) Init() error {
	ctx := context.Background()
	fmt.Println("initializing tournament")
	tm.mu.Lock()
	defer tm.mu.Unlock()
	// Retrieve the list from Redis
	list, err := redis.GetList(ctx)
	limit, err2 := redis.GetLimit(ctx)
	exists, err3 := redis.GetExists(ctx)
	if err != nil {
		return err
	}
	if err2 != nil {
		return err2
	}
	if err3 != nil {
		return err3
	}
	tm.List = list
	tm.Limit = limit
	tm.Exists = exists
	if !tm.Exists && len(tm.List) > 0 {
		fmt.Println("tournament does not exist but list is not empty, clearing list")
		if err := tm.removeTournament(ctx); err != nil {
			return err
		}
	}
	fmt.Println("tournament initialized")
	return nil
}

func (tm *TournamentManager) AddPlayer(ctx context.Context, player types.Player) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	tm.List = append(tm.List, player)
	if err := redis.SetList(ctx, tm.List); err != nil {
		fmt.Printf("error happened while adding to redis list: %s", err)
		return err
	}
	return nil
}

func (tm *TournamentManager) CreateTournament(ctx context.Context, limit int, ratingLimit int) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	tm.Exists = true
	tm.List = []types.Player{}
	if err := redis.SetExists(ctx, true); err != nil {
		fmt.Printf("error happened while saving list state to redis: %s", err)
		return err
	}
	if err := redis.SetList(ctx, tm.List); err != nil {
		fmt.Printf("error happened while saving list state to redis: %s", err)
		return err
	}
	return nil
}

func (tm *TournamentManager) RemoveTournament(ctx context.Context) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	tm.Exists = false
	if err := tm.clearList(ctx); err != nil {
		fmt.Printf("error happened while clearing the redis list: %s", err)
		return err
	}
	if err := redis.SetExists(ctx, false); err != nil {
		fmt.Printf("error happened while saving list state to redis: %s", err)
		return err
	}
	return nil
}

func (tm *TournamentManager) removeTournament(ctx context.Context) error {
	tm.Exists = false
	if err := tm.clearList(ctx); err != nil {
		fmt.Printf("error happened while clearing the redis list: %s", err)
		return err
	}
	if err := redis.SetExists(ctx, false); err != nil {
		fmt.Printf("error happened while saving list state to redis: %s", err)
		return err
	}
	return nil
}

func (tm *TournamentManager) clearList(ctx context.Context) error {
	tm.List = []types.Player{}
	fmt.Printf("clearing list KUYFIUFUT: %v", tm.List)
	if err := redis.SetList(ctx, tm.List); err != nil {
		fmt.Printf("error happened while clearing the redis list: %s", err)
	}
	return nil
}

func (tm *TournamentManager) EditPlayer(ctx context.Context, playerID int, updatedPlayer types.Player) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	for i, player := range tm.List {
		if player.ID == playerID {
			tm.List[i] = updatedPlayer
			if err := redis.SetList(ctx, tm.List); err != nil {
				fmt.Printf("error happened while updating the redis list: %s", err)
				return err
			}
			return nil
		}
	}

	return fmt.Errorf("player with ID %d not found in list", playerID)
}

func (tm *TournamentManager) RemovePlayer(ctx context.Context, playerID int) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	for i, player := range tm.List {
		if player.ID == playerID {
			tm.List = append(tm.List[:i], tm.List[i+1:]...)
			if err := redis.SetList(ctx, tm.List); err != nil {
				fmt.Printf("error happened while updating the redis list: %s", err)
				return err
			}
			return nil
		}
	}

	return fmt.Errorf("player with ID %d not found in list", playerID)
}

func (tm *TournamentManager) Sync(ctx context.Context) error {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	if err := redis.SetList(ctx, tm.List); err != nil {
		fmt.Printf("error happened while updating the redis list: %s", err)
		return err
	}
	if err := redis.SetExists(ctx, tm.Exists); err != nil {
		fmt.Printf("error happened while updating the redis exists: %s", err)
		return err
	}
	if err := redis.SetLimit(ctx, tm.Limit); err != nil {
		fmt.Printf("error happened while updating the redis limit: %s", err)
		return err
	}
	return nil
}

func (tm *TournamentManager) RemoveState(ctx context.Context, stateID int) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	result := []types.Player{}
	for _, state := range tm.List {
		if state.ID != stateID {
			result = append(result, state)
		}
	}
	tm.List = result
	if err := redis.SetList(ctx, result); err != nil {
		fmt.Printf("error happened while updating the redis list: %s", err)
		return err
	}
	return nil
}
func (tm *TournamentManager) GetTournamentJSON() (string, error) {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	data := struct {
		List   []types.Player `json:"players"`
		Limit  int            `json:"limit"`
		Exists bool           `json:"exists"`
	}{
		List:   tm.List,
		Limit:  tm.Limit,
		Exists: tm.Exists,
	}

	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal tournament state: %w", err)
	}

	return string(jsonData), nil
}

func (tm *TournamentManager) SetLimit(ctx context.Context, limit int) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	tm.Limit = limit
	if err := redis.SetLimit(ctx, limit); err != nil {
		fmt.Printf("error happened while updating the redis limit: %s", err)
		return err
	}
	return nil
}
