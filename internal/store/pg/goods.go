package pg

import (
	"context"
	"errors"
	"fmt"

	"github.com/Saaghh/hezzl-hr/internal/model"
	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"
)

func (p *Postgres) CreateProject(ctx context.Context, project model.Project) (*model.Project, error) {
	query := `
	INSERT INTO projects (name) 
	VALUES ($1)
	RETURNING id, created_at`

	err := p.db.QueryRow(
		ctx,
		query,
		project.Name,
	).Scan(
		&project.ID,
		&project.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("p.db.QueryRow(...).Scan(...): %w", err)
	}

	return &project, nil
}

func (p *Postgres) CreateGoods(ctx context.Context, goods model.Goods) (*model.Goods, error) {
	query := `SELECT COALESCE(MAX(priority), 0) FROM goods`

	err := p.db.QueryRow(
		ctx,
		query,
	).Scan(&goods.Priority)
	if err != nil {
		return nil, fmt.Errorf("p.db.QueryRow(...).Scan(&maxPriority): %w", err)
	}

	goods.Priority++

	query = `
	INSERT INTO goods (project_id, name, priority)
	VALUES ($1, $2, $3)
	RETURNING id, description, removed, created_at`

	err = p.db.QueryRow(
		ctx,
		query,
		goods.ProjectID,
		goods.Name,
		goods.Priority,
	).Scan(
		&goods.ID,
		&goods.Description,
		&goods.Removed,
		&goods.CreatedAt,
	)
	// TODO: add foreign key violation check
	if err != nil {
		return nil, fmt.Errorf("p.db.QueryRow().Scan(): %w", err)
	}

	return &goods, nil
}

func (p *Postgres) UpdateGoods(ctx context.Context, request model.UpdateGoodsRequest) (*model.Goods, error) {
	tx, err := p.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("p.db.Begin(ctx): %w", err)
	}

	defer func() {
		err := tx.Rollback(ctx)
		if err != nil && !errors.Is(err, pgx.ErrTxClosed) {
			zap.L().With(zap.Error(err)).Warn("UpdateWallet/tx.Rollback(ctx)")
		}
	}()

	query := `
	SELECT * FROM goods WHERE id = $1 AND project_id = $2 FOR UPDATE`

	rows, err := tx.Query(
		ctx,
		query,
		request.ID,
		request.ProjectID,
	)
	rows.Close()

	switch {
	case errors.Is(err, pgx.ErrNoRows):
		return nil, model.ErrGoodNotFound
	case err != nil:
		return nil, fmt.Errorf("tx.QueryRow(ctx, query, request.ID, request.ProjectID).Scan(): %w", err)
	}

	query = `
	UPDATE goods
	SET name = $1
	WHERE removed = false and id = $2 and project_id = $3
	RETURNING id, project_id, name, description, priority, removed, created_at`

	var goods model.Goods

	err = tx.QueryRow(
		ctx,
		query,
		request.Name,
		request.ID,
		request.ProjectID,
	).Scan(
		&goods.ID,
		&goods.ProjectID,
		&goods.Name,
		&goods.Description,
		&goods.Priority,
		&goods.Removed,
		&goods.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("tx.QueryRow(...).Scan(...): %w", err)
	}

	if request.Description != nil {
		query = `
		UPDATE goods
		SET description = $1
		WHERE removed = false and id = $2 and project_id = $3
		RETURNING description`

		err = tx.QueryRow(
			ctx,
			query,
			request.Description,
			request.ID,
			request.ProjectID,
		).Scan(
			&goods.Description,
		)
		if err != nil {
			return nil, fmt.Errorf("tx.QueryRow(...).Scan(...): %w", err)
		}
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("tx.Commit(ctx): %w", err)
	}

	return &goods, nil
}

func (p *Postgres) DeleteGoods(ctx context.Context, goods model.Goods) (*model.Goods, error) {
	query := `
	UPDATE goods
	SET removed = true
	WHERE removed = false and id = $1 and project_id = $2
	RETURNING removed`

	err := p.db.QueryRow(
		ctx,
		query,
		goods.ID,
		goods.ProjectID,
	).Scan(
		&goods.Removed,
	)

	switch {
	case errors.Is(err, pgx.ErrNoRows):
		return nil, model.ErrGoodNotFound
	case err != nil:
		return nil, fmt.Errorf("p.db.QueryRow(...).Scan(): %w", err)
	}

	return &goods, nil
}

func (p *Postgres) GetMetaData(ctx context.Context) (*model.GetMetaData, error) {
	var totalRecords model.GetMetaData

	query := `SELECT COALESCE(COUNT(*), 0) FROM goods`

	err := p.db.QueryRow(
		ctx,
		query,
	).Scan(
		&totalRecords.Total,
	)
	if err != nil {
		return nil, fmt.Errorf("p.db.QueryRow(): %w", err)
	}

	query += " WHERE removed = true"

	err = p.db.QueryRow(
		ctx,
		query,
	).Scan(
		&totalRecords.Removed,
	)
	if err != nil {
		return nil, fmt.Errorf("p.db.QueryRow(): %w", err)
	}

	return &totalRecords, nil
}

func (p *Postgres) GetGoods(ctx context.Context, params model.GetParams) (*[]model.Goods, error) {
	goods := make([]model.Goods, 0, 1)

	query := `
	SELECT id, project_id, name, description, priority, removed, created_at
	FROM goods
	WHERE removed = false
	LIMIT $1 OFFSET $2`

	rows, err := p.db.Query(
		ctx,
		query,
		params.Limit,
		params.Offset)
	if err != nil {
		return nil, fmt.Errorf("p.db.Query(...): %w", err)
	}

	for rows.Next() {
		var good model.Goods

		err = rows.Scan(
			&good.ID,
			&good.ProjectID,
			&good.Name,
			&good.Description,
			&good.Priority,
			&good.Removed,
			&good.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("rows.Scan(...): %w", err)
		}

		goods = append(goods, good)
	}

	return &goods, nil
}

func (p *Postgres) ReprioritizeGoods(ctx context.Context, goods model.UpdatePriorityRequest) (*[]model.Goods, error) {
	tx, err := p.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("p.db.Begin(ctx): %w", err)
	}

	var changedGoods []model.Goods

	defer func() {
		err := tx.Rollback(ctx)
		if err != nil && !errors.Is(err, pgx.ErrTxClosed) {
			zap.L().With(zap.Error(err)).Warn("UpdateWallet/tx.Rollback(ctx)")
		}
	}()

	query := `
	UPDATE goods
	SET priority = $1
	WHERE id = $2 and project_id = $3 and removed = false
	RETURNING priority`

	err = tx.QueryRow(
		ctx,
		query,
		goods.Priority,
		goods.ID,
		goods.ProjectID,
	).Scan(
		&goods.Priority,
	)
	if err != nil {
		return nil, fmt.Errorf("tx.QueryRow(...): %w", err)
	}

	changedGoods = append(changedGoods, model.Goods{ID: goods.ID, Priority: goods.Priority})

	query = `
	UPDATE goods
	SET priority = priority + 1
	WHERE priority > $1 and removed = false
	RETURNING id, priority`

	rows, err := tx.Query(
		ctx,
		query,
		goods.Priority)
	if err != nil {
		return nil, fmt.Errorf("tx.Query(...): %w", err)
	}

	for rows.Next() {
		var good model.Goods

		err = rows.Scan(
			&good.ID,
			&good.Priority)
		if err != nil {
			return nil, fmt.Errorf("rows.Scan(...): %w", err)
		}

		changedGoods = append(changedGoods, good)
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("tx.Commit(ctx): %w", err)
	}

	return &changedGoods, nil
}
