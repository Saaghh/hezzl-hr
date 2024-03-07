package store

import (
	"context"
	"fmt"

	"github.com/Saaghh/hezzl-hr/internal/model"
	"go.uber.org/zap"
)

func (c *Clickhouse) SaveGoodsEvents(ctx context.Context, goods *[]model.GoodsEvent) error {
	zap.L().Debug("Starting batch insert for goods events")

	tx, err := c.conn.Begin()
	if err != nil {
		return fmt.Errorf("c.conn.Begin(): %w", err)
	}

	stmt, err := tx.PrepareContext(ctx, "INSERT INTO goods_logs (id, project_id, name, description, priority, removed, event_time) VALUES (?, ?, ?, ?, ?, ?, ?)")
	if err != nil {
		return fmt.Errorf("tx.PrepareContext(...): %w", err)
	}

	defer func() {
		err = stmt.Close()
		if err != nil {
			zap.L().With(zap.Error(err)).Warn("SaveGoodsEvents/stmt.Close()")
		}
	}()

	for _, event := range *goods {
		removed := 0
		if event.Removed {
			removed = 1
		}

		if _, err = stmt.ExecContext(ctx,
			event.ID,
			event.ProjectID,
			event.Name,
			event.Description,
			event.Priority,
			removed,
			event.EventTime,
		); err != nil {
			if err := tx.Rollback(); err != nil {
				zap.L().Error("tx.Rollback()", zap.Error(err))
			}

			return fmt.Errorf("stmt.ExecContext(...): %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("tx.Commit(): %w", err)
	}

	zap.L().Debug("Batch insert for goods events completed successfully")

	return nil
}
