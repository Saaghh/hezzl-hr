package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Saaghh/hezzl-hr/internal/model"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

type Service struct {
	db   store
	cash cashdb
	bl   brokerLogger
}

type cashdb interface {
	StoreGetResponse(ctx context.Context, response model.GetListResponse) error
	InvalidateAllData(ctx context.Context) error
	GetListResponse(ctx context.Context, params model.ListParams) (*model.GetListResponse, error)
}

type brokerLogger interface {
	PublishEvent(event model.GoodsEvent) error
}

type store interface {
	CreateProject(ctx context.Context, project model.Project) (*model.Project, error)

	CreateGoods(ctx context.Context, goods model.Goods) (*model.Goods, error)
	UpdateGoods(ctx context.Context, request model.UpdateGoodsRequest) (*model.Goods, error)
	DeleteGoods(ctx context.Context, goods model.Goods) (*model.Goods, error)
	GetGoods(ctx context.Context, params model.ListParams) (*[]model.Goods, error)
	GetMetaData(ctx context.Context) (*model.ListParams, error)
	ReprioritizeGoods(ctx context.Context, goods model.UpdatePriorityRequest) (*[]model.Goods, error)
}

func New(db store, cash cashdb, bl brokerLogger) *Service {
	return &Service{
		db:   db,
		cash: cash,
		bl:   bl,
	}
}

func (s *Service) CreateProject(ctx context.Context, project model.Project) (*model.Project, error) {
	result, err := s.db.CreateProject(ctx, project)
	if err != nil {
		return nil, fmt.Errorf("s.db.CreateProject(ctx, project): %w", err)
	}

	return result, nil
}

func (s *Service) CreateGoods(ctx context.Context, goods model.Goods) (*model.Goods, error) {
	resultGood, err := s.db.CreateGoods(ctx, goods)
	if err != nil {
		return nil, fmt.Errorf("s.db.CreateGoods(ctx, goods): %w", err)
	}

	if err = s.cash.InvalidateAllData(ctx); err != nil {
		zap.L().With(zap.Error(err)).Warn("CreateGoods/s.cash.InvalidateAllData(ctx)")
	}

	return resultGood, nil
}

func (s *Service) UpdateGoods(ctx context.Context, request model.UpdateGoodsRequest) (*model.Goods, error) {
	if request.Name == "" {
		return nil, model.ErrBlankName
	}

	result, err := s.db.UpdateGoods(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("s.db.UpdateGoods(ctx, request): %w", err)
	}

	if err = s.cash.InvalidateAllData(ctx); err != nil {
		zap.L().With(zap.Error(err)).Warn("UpdateGoods/s.cash.InvalidateAllData(ctx)")
	}

	err = s.bl.PublishEvent(model.GoodsEvent{
		Goods:     *result,
		EventTime: time.Now(),
	})
	if err != nil {
		zap.L().With(zap.Error(err)).Warn("UpdateGoods/s.bl.PublishEvent(...)")
	}

	return result, nil
}

func (s *Service) DeleteGoods(ctx context.Context, goods model.Goods) (*model.Goods, error) {
	result, err := s.db.DeleteGoods(ctx, goods)
	if err != nil {
		return nil, fmt.Errorf("s.db.DeleteGoods(ctx, goods): %w", err)
	}

	if err = s.cash.InvalidateAllData(ctx); err != nil {
		zap.L().With(zap.Error(err)).Warn("DeleteGoods/s.cash.InvalidateAllData(ctx)")
	}

	err = s.bl.PublishEvent(model.GoodsEvent{
		Goods:     *result,
		EventTime: time.Now(),
	})
	if err != nil {
		zap.L().With(zap.Error(err)).Warn("DeleteGoods/s.bl.PublishEvent(...)")
	}

	return result, nil
}

func (s *Service) GetGoods(ctx context.Context, params model.ListParams) (*model.GetListResponse, error) {
	cashedGoods, err := s.cash.GetListResponse(ctx, params)

	switch {
	case err == nil:
		return cashedGoods, nil
	case errors.Is(err, redis.Nil):
		break
	case err != nil:
		zap.L().With(zap.Error(err)).Warn("GetGoods/s.cash.GetListResponse(ctx, params)")
	}

	result, err := s.db.GetGoods(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("s.db.GetGoods(ctx, params): %w", err)
	}

	metaData, err := s.db.GetMetaData(ctx)
	if err != nil {
		return nil, fmt.Errorf("s.db.GetMetaData(ctx): %w", err)
	}

	metaData.Offset = params.Offset
	metaData.Limit = params.Limit

	listResponse := model.GetListResponse{
		Meta:      *metaData,
		GoodsList: *result,
	}

	if err = s.cash.StoreGetResponse(ctx, listResponse); err != nil {
		zap.L().With(zap.Error(err)).Warn("GetGoods/s.cash.StoreGetResponse(ctx, listResponse)")
	}

	return &listResponse, nil
}

func (s *Service) ReprioritizeGoods(ctx context.Context, goods model.UpdatePriorityRequest) (*[]model.Goods, error) {
	if goods.Priority < 1 {
		return nil, model.ErrWrongPriority
	}

	result, err := s.db.ReprioritizeGoods(ctx, goods)
	if err != nil {
		return nil, fmt.Errorf("s.db.ReprioritizeGoods(ctx, goods): %w", err)
	}

	if err = s.cash.InvalidateAllData(ctx); err != nil {
		zap.L().With(zap.Error(err)).Warn("ReprioritizeGoods/s.cash.InvalidateAllData(ctx)")
	}

	for _, value := range *result {
		err = s.bl.PublishEvent(model.GoodsEvent{
			Goods:     value,
			EventTime: time.Now(),
		})
		if err != nil {
			zap.L().With(zap.Error(err)).Warn("ReprioritizeGoods/s.bl.PublishEvent(...)")
		}
	}

	return result, nil
}
