package service

import (
	"context"
	"fmt"

	"github.com/Saaghh/hezzl-hr/internal/model"
)

type Service struct {
	db store
}

type store interface {
	CreateProject(ctx context.Context, project model.Project) (*model.Project, error)

	CreateGoods(ctx context.Context, goods model.Goods) (*model.Goods, error)
	UpdateGoods(ctx context.Context, request model.UpdateGoodsRequest) (*model.Goods, error)
	DeleteGoods(ctx context.Context, goods model.Goods) (*model.Goods, error)
	GetGoods(ctx context.Context, params model.GetParams) (*[]model.Goods, error)
	GetMetaData(ctx context.Context) (*model.GetMetaData, error)
	ReprioritizeGoods(ctx context.Context, goods model.UpdatePriorityRequest) (*[]model.Goods, error)
}

func New(db store) *Service {
	return &Service{
		db: db,
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

	// TODO: invalidate redis data
	// TODO: clickhouse log

	return result, nil
}

func (s *Service) DeleteGoods(ctx context.Context, goods model.Goods) (*model.Goods, error) {
	result, err := s.db.DeleteGoods(ctx, goods)
	if err != nil {
		return nil, fmt.Errorf("s.db.DeleteGoods(ctx, goods): %w", err)
	}

	// TODO: invalidate redis data
	// TODO: clickhouse log

	return result, nil
}

func (s *Service) GetGoods(ctx context.Context, params model.GetParams) (*[]model.Goods, *model.GetMetaData, error) {
	// TODO: check for data in redis
	result, err := s.db.GetGoods(ctx, params)
	if err != nil {
		return nil, nil, fmt.Errorf("s.db.GetGoods(ctx, params): %w", err)
	}

	metaData, err := s.db.GetMetaData(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("s.db.GetMetaData(ctx): %w", err)
	}

	metaData.Offset = params.Offset
	metaData.Limit = params.Limit

	// TODO: cash data in redis

	return result, metaData, nil
}

func (s *Service) ReprioritizeGoods(ctx context.Context, goods model.UpdatePriorityRequest) (*[]model.Goods, error) {
	if goods.Priority < 1 {
		return nil, model.ErrWrongPriority
	}

	result, err := s.db.ReprioritizeGoods(ctx, goods)
	if err != nil {
		return nil, fmt.Errorf("s.db.ReprioritizeGoods(ctx, goods): %w", err)
	}

	// TODO: invalidate Redis data
	// TODO: clickhouse log

	return result, nil
}
