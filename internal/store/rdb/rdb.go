package rdb

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/Saaghh/hezzl-hr/internal/config"
	"github.com/Saaghh/hezzl-hr/internal/model"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

type Redis struct {
	client         *redis.Client
	defaultTimeout time.Duration
}

func New(cfg *config.Config) *Redis {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
	})

	return &Redis{
		client:         client,
		defaultTimeout: cfg.RedisDefaultTimeout,
	}
}

func (r *Redis) StoreGetResponse(ctx context.Context, response model.GetListResponse) error {
	key := r.getParamsKey(model.ListParams{
		Limit:  response.Meta.Limit,
		Offset: response.Meta.Offset,
	})

	serializedResponse, err := json.Marshal(response)
	if err != nil {
		return fmt.Errorf("json.Marshal(response): %w", err)
	}

	if err = r.client.Set(ctx, key, serializedResponse, r.defaultTimeout).Err(); err != nil {
		return fmt.Errorf("r.client.Set(ctx, key, serializedResponse, r.defaultTimeout).Err(): %w", err)
	}

	zap.L().Debug("successfully saved data to redis", zap.String("key", key))

	return nil
}

func (r *Redis) InvalidateAllData(ctx context.Context) error {
	if err := r.client.FlushDB(ctx).Err(); err != nil {
		return fmt.Errorf("r.client.FlushDB(ctx).Err(): %w", err)
	}

	zap.L().Debug("successfully invalidated all data")

	return nil
}

func (r *Redis) GetListResponse(ctx context.Context, params model.ListParams) (*model.GetListResponse, error) {
	res, err := r.client.Get(ctx, r.getParamsKey(params)).Result()
	if err != nil {
		return nil, fmt.Errorf("r.client.Get(ctx, r.getParamsKey(params)).Result(): %w", err)
	}

	var resultList model.GetListResponse
	if err = json.Unmarshal([]byte(res), &resultList); err != nil {
		return nil, fmt.Errorf("json.Unmarshal([]byte(res), &resultList): %w", err)
	}

	zap.L().Debug("successfully returned data from redis", zap.Int("length", resultList.Meta.Total))

	return &resultList, nil
}

func (r *Redis) getParamsKey(params model.ListParams) string {
	return strconv.Itoa(params.Offset) + "-" + strconv.Itoa(params.Limit)
}
