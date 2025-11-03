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
	return Client.Set(ctx, "tournament_list_dev", listJSON, 0).Err()
}

func GetList(ctx context.Context) ([]types.Player, error) {
	data, err := Client.Get(ctx, "tournament_list_dev").Bytes()
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

func SetMetadata(ctx context.Context, metadata types.TournamentMetadata) error {
	metadataJSON, err := json.Marshal(metadata)
	if err != nil {
		return err
	}
	return Client.Set(ctx, "tournament_metadata_dev", metadataJSON, 0).Err()
}

func GetMetadata(ctx context.Context) (types.TournamentMetadata, error) {
	data, err := Client.Get(ctx, "tournament_metadata_dev").Bytes()
	if err != nil {
		if err == redisClient.Nil {
			return types.TournamentMetadata{}, nil
		}
		return types.TournamentMetadata{}, err
	}
	var metadata types.TournamentMetadata
	if err := json.Unmarshal(data, &metadata); err != nil {
		return types.TournamentMetadata{}, err
	}
	return metadata, nil
}
