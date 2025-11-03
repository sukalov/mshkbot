package redis

import (
	"context"
	"encoding/json"
	"strconv"

	redisClient "github.com/go-redis/redis/v8"
	"github.com/sukalov/mshkbot/internal/types"
)

func SetList(ctx context.Context, list []types.Player) error {
	listJSON, err := json.Marshal(list)
	if err != nil {
		return err
	}
	return Client.Set(ctx, "tournament_list", listJSON, 0).Err()
}

func GetList(ctx context.Context) ([]types.Player, error) {
	data, err := Client.Get(ctx, "tournament_list").Bytes()
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
	return Client.Set(ctx, "tournament_metadata", metadataJSON, 0).Err()
}

func GetMetadata(ctx context.Context) (types.TournamentMetadata, error) {
	data, err := Client.Get(ctx, "tournament_metadata").Bytes()
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

func StoreMessageMapping(ctx context.Context, userMessageID int, botMessageID int) error {
	key := "message_mapping"
	field := strconv.Itoa(userMessageID)
	return Client.HSet(ctx, key, field, botMessageID).Err()
}

func GetBotMessageID(ctx context.Context, userMessageID int) (int, error) {
	key := "message_mapping"
	field := strconv.Itoa(userMessageID)
	result, err := Client.HGet(ctx, key, field).Int()
	if err != nil {
		if err == redisClient.Nil {
			return 0, nil
		}
		return 0, err
	}
	return result, nil
}

func DeleteMessageMapping(ctx context.Context, userMessageID int) error {
	key := "message_mapping"
	field := strconv.Itoa(userMessageID)
	return Client.HDel(ctx, key, field).Err()
}

func TrimMessageMappings(ctx context.Context, maxSize int64) error {
	key := "message_mapping"
	size, err := Client.HLen(ctx, key).Result()
	if err != nil {
		return err
	}

	if size <= maxSize {
		return nil
	}

	toDelete := size - maxSize
	fields, err := Client.HKeys(ctx, key).Result()
	if err != nil {
		return err
	}

	if int64(len(fields)) > toDelete {
		fieldsToDelete := fields[:toDelete]
		return Client.HDel(ctx, key, fieldsToDelete...).Err()
	}

	return nil
}
