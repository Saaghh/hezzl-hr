package model

import (
	"fmt"
	"net/url"
	"time"

	"github.com/gorilla/schema"
)

type Goods struct {
	ID          int64      `json:"id" schema:"id"`
	ProjectID   int64      `json:"projectId,omitempty" schema:"projectId"`
	Name        string     `json:"name,omitempty"`
	Description string     `json:"description,omitempty"`
	Priority    int        `json:"priority,omitempty"`
	Removed     bool       `json:"removed"`
	CreatedAt   *time.Time `json:"createdAt,omitempty"`
}

type GoodsEvent struct {
	Goods
	EventTime time.Time `json:"eventTime"`
}

type Project struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"createdAt"`
}

type UpdateGoodsRequest struct {
	ID          int64   `json:"id"`
	ProjectID   int64   `json:"projectId"`
	Name        string  `json:"name"`
	Description *string `json:"description"`
}

type UpdatePriorityRequest struct {
	ID        int64 `json:"id" schema:"id"`
	ProjectID int64 `json:"projectId"`
	Priority  int   `json:"newPriority"`
}

type ListParams struct {
	Limit   int `json:"limit,omitempty" schema:"limit"`
	Offset  int `json:"offset,omitempty" schema:"offset"`
	Total   int `json:"total,omitempty"`
	Removed int `json:"removed,omitempty"`
}

type GetListResponse struct {
	Meta      ListParams `json:"meta"`
	GoodsList []Goods    `json:"goods"`
}

func DecodeQueryParams(url url.URL, target any) error {
	err := schema.NewDecoder().Decode(target, url.Query())
	if err != nil {
		return fmt.Errorf("schema.NewDecoder().Decode(target, url.Query()):%w", err)
	}

	return nil
}
