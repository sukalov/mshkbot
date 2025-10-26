package redis

import (
	"context"
	"encoding/json"

	redisClient "github.com/go-redis/redis/v8"
	"github.com/sukalov/mshkbot/internal/types"
)

func SetList(ctx context.Context, list []types.Player) error {
	listJSON, err := json.Marshal(list)
	if err != nil {
		return err
	}
	return Client.Set(ctx, "tournament", listJSON, 0).Err()
}

func GetList(ctx context.Context) ([]types.Player, error) {
	data, err := Client.Get(ctx, "tournament").Bytes()
	if err != nil {
		if err == redisClient.Nil {
			return []types.Player{}, nil
		}
		return nil, err
	}
	var list []types.Player
	if err := json.Unmarshal(data, &list); err != nil {
		return nil, err
	}
	return list, nil
}

func GetLimit(ctx context.Context) (int, error) {
	data, err := Client.Get(ctx, "tournament_limit").Bytes()
	if err != nil {
		if err == redisClient.Nil {
			return 0, nil
		}
		return 0, err
	}
	var limit int
	if err := json.Unmarshal(data, &limit); err != nil {
		return 0, err
	}
	return limit, nil
}

func SetLimit(ctx context.Context, limit int) error {
	limitJSON, err := json.Marshal(limit)
	if err != nil {
		return err
	}
	return Client.Set(ctx, "tournament_limit", limitJSON, 0).Err()
}

func GetExists(ctx context.Context) (bool, error) {
	data, err := Client.Get(ctx, "tournament_exists").Bytes()
	if err != nil {
		if err == redisClient.Nil {
			return false, nil
		}
		return false, err
	}
	var exists bool
	if err := json.Unmarshal(data, &exists); err != nil {
		return false, err
	}
	return exists, nil
}

func SetExists(ctx context.Context, exists bool) error {
	existsJSON, err := json.Marshal(exists)
	if err != nil {
		return err
	}
	return Client.Set(ctx, "tournament_exists", existsJSON, 0).Err()
}
