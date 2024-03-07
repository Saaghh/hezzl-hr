package apiserver

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/Saaghh/hezzl-hr/internal/model"
	"go.uber.org/zap"
)

type service interface {
	CreateGoods(ctx context.Context, goods model.Goods) (*model.Goods, error)
	UpdateGoods(ctx context.Context, request model.UpdateGoodsRequest) (*model.Goods, error)
	DeleteGoods(ctx context.Context, goods model.Goods) (*model.Goods, error)
	GetGoods(ctx context.Context, params model.ListParams) (*model.GetListResponse, error)
	ReprioritizeGoods(ctx context.Context, goods model.UpdatePriorityRequest) (*[]model.Goods, error)
}

type ErrorResponse struct {
	Code    int                    `json:"code"`
	Message string                 `json:"message"`
	Details map[string]interface{} `json:"details"`
}

type GetListResponse struct {
	Meta  *model.ListParams `json:"meta"`
	Goods *[]model.Goods    `json:"goods"`
}

type ReprioritizeResponse struct {
	Priorities *[]model.Goods `json:"priorities"`
}

func (s *APIServer) createGood(w http.ResponseWriter, r *http.Request) {
	var requestGoods model.Goods

	if err := json.NewDecoder(r.Body).Decode(&requestGoods); err != nil {
		writeErrorResponse(w, http.StatusBadRequest, 0, "error.FailedToReadBody", make(map[string]any))

		return
	}

	if err := model.DecodeQueryParams(*r.URL, &requestGoods); err != nil {
		writeErrorResponse(w, http.StatusBadRequest, 0, "error.FailedToReadQuery", make(map[string]any))

		return
	}

	goods, err := s.service.CreateGoods(r.Context(), requestGoods)
	if err != nil {
		zap.L().With(zap.Error(err)).Warn("createGood/s.service.CreateGoods(r.Context(), requestGoods)")
		writeErrorResponse(w, http.StatusInternalServerError, 5, "errors.InternalServerError", make(map[string]any))

		return
	}

	writeOkResponse(w, http.StatusCreated, goods)
}

func (s *APIServer) updateGoods(w http.ResponseWriter, r *http.Request) {
	var updateRequest model.UpdateGoodsRequest

	err := json.NewDecoder(r.Body).Decode(&updateRequest)
	if err != nil {
		writeErrorResponse(w, http.StatusBadRequest, 0, "error.FailedToReadBody", make(map[string]any))

		return
	}

	if err = model.DecodeQueryParams(*r.URL, &updateRequest); err != nil {
		writeErrorResponse(w, http.StatusBadRequest, 0, "error.FailedToReadQuery", make(map[string]any))

		return
	}

	goods, err := s.service.UpdateGoods(r.Context(), updateRequest)

	switch {
	case errors.Is(err, model.ErrGoodNotFound):
		writeErrorResponse(w, http.StatusNotFound, 3, "errors.good.notFound", make(map[string]any))

		return
	case errors.Is(err, model.ErrBlankName):
		writeErrorResponse(w, http.StatusBadRequest, 1, "errors.good.blankName", make(map[string]any))

		return
	case err != nil:
		zap.L().With(zap.Error(err)).Warn("updateGoods/s.service.UpdateGoods(r.Context(), updateRequest)")
		writeErrorResponse(w, http.StatusInternalServerError, 5, "errors.InternalServerError", make(map[string]any))

		return
	}

	writeOkResponse(w, http.StatusOK, goods)
}

func (s *APIServer) removeGoods(w http.ResponseWriter, r *http.Request) {
	var (
		goods model.Goods
		err   error
	)

	if err = model.DecodeQueryParams(*r.URL, &goods); err != nil {
		writeErrorResponse(w, http.StatusBadRequest, 0, "error.FailedToReadQuery", make(map[string]any))

		return
	}

	deletedGoods, err := s.service.DeleteGoods(r.Context(), goods)

	switch {
	case errors.Is(err, model.ErrGoodNotFound):
		writeErrorResponse(w, http.StatusNotFound, 3, "errors.good.notFound", make(map[string]any))

		return
	case err != nil:
		zap.L().With(zap.Error(err)).Warn("removeGoods/s.service.DeleteGoods(r.Context(), goods)")
		writeErrorResponse(w, http.StatusInternalServerError, 5, "errors.InternalServerError", make(map[string]any))

		return
	}

	writeOkResponse(w, http.StatusOK, deletedGoods)
}

func (s *APIServer) getGoods(w http.ResponseWriter, r *http.Request) {
	var params model.ListParams
	if err := model.DecodeQueryParams(*r.URL, &params); err != nil {
		writeErrorResponse(w, http.StatusBadRequest, 0, "error.FailedToReadQuery", make(map[string]any))

		return
	}

	result, err := s.service.GetGoods(r.Context(), params)
	if err != nil {
		zap.L().With(zap.Error(err)).Warn("getGoods/s.service.GetGoods(r.Context(), *params)")

		writeErrorResponse(w, http.StatusInternalServerError, 5, "errors.InternalServerError", make(map[string]any))

		return
	}

	writeOkResponse(w, http.StatusOK, result)
}

func (s *APIServer) reprioritizeGood(w http.ResponseWriter, r *http.Request) {
	var (
		goods model.UpdatePriorityRequest
		err   error
	)

	err = json.NewDecoder(r.Body).Decode(&goods)
	if err != nil {
		writeErrorResponse(w, http.StatusBadRequest, 0, "error.FailedToReadBody", make(map[string]any))

		return
	}

	if err = model.DecodeQueryParams(*r.URL, &goods); err != nil {
		writeErrorResponse(w, http.StatusBadRequest, 0, "error.FailedToReadQuery", make(map[string]any))

		return
	}

	changedGoods, err := s.service.ReprioritizeGoods(r.Context(), goods)
	if err != nil {
		zap.L().With(zap.Error(err)).Warn("reprioritizeGood/s.service.ReprioritizeGoods(r.Context(), goods)")
		writeErrorResponse(w, http.StatusInternalServerError, 5, "errors.InternalServerError", make(map[string]any))

		return
	}

	writeOkResponse(w, http.StatusOK, ReprioritizeResponse{Priorities: changedGoods})
}

func writeOkResponse(w http.ResponseWriter, statusCode int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	err := json.NewEncoder(w).Encode(data)
	if err != nil {
		zap.L().With(zap.Error(err)).Warn(
			"writeOkResponse/json.NewEncoder(w).Encode(data)")
	}
}

func writeErrorResponse(w http.ResponseWriter, statusCode int, errorCode int, message string, details map[string]any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := ErrorResponse{
		Code:    errorCode,
		Message: message,
		Details: details,
	}

	err := json.NewEncoder(w).Encode(response)
	if err != nil {
		zap.L().With(zap.Error(err)).Warn(
			"writeErrorResponse/json.NewEncoder(w).Encode(data)")
	}
}
