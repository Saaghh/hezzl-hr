package model

import (
	"fmt"
	"net/url"
	"strconv"
	"time"
)

type Goods struct {
	ID          int64     `json:"id"`
	ProjectID   int64     `json:"projectId"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Priority    int       `json:"priority"`
	Removed     bool      `json:"removed"`
	CreatedAt   time.Time `json:"createdAt"`
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
	ID        int64 `json:"id"`
	ProjectID int64 `json:"projectId"`
	Priority  int   `json:"newPriority"`
}

type GetParams struct {
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
}

type GetMetaData struct {
	GetParams
	Total   int `json:"total"`
	Removed int `json:"removed"`
}

func URLToGetParams(url url.URL) (*GetParams, error) {
	// TODO: make better. Check for empty params
	var err error

	params := GetParams{
		Limit:  10,
		Offset: 1,
	}

	params.Offset, err = strconv.Atoi(url.Query().Get("offset"))
	if err != nil {
		return nil, fmt.Errorf("strconv.Atoi(url.Query().Get(\"offset\")): %w", err)
	}

	params.Limit, err = strconv.Atoi(url.Query().Get("limit"))
	if err != nil {
		return nil, fmt.Errorf("strconv.Atoi(url.Query().Get(\"limit\")): %w", err)
	}

	return &params, nil
}
