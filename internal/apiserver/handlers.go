package apiserver

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/Saaghh/hezzl-hr/internal/model"
	"go.uber.org/zap"
)

type service interface {
	CreateGoods(ctx context.Context, goods model.Goods) (*model.Goods, error)
	UpdateGoods(ctx context.Context, request model.UpdateGoodsRequest) (*model.Goods, error)
	DeleteGoods(ctx context.Context, goods model.Goods) (*model.Goods, error)
	GetGoods(ctx context.Context, params model.GetParams) (*[]model.Goods, *model.GetMetaData, error)
	ReprioritizeGoods(ctx context.Context, goods model.UpdatePriorityRequest) (*[]model.Goods, error)
}

type GetListResponse struct {
	Meta  *model.GetMetaData `json:"meta"`
	Goods *[]model.Goods     `json:"goods"`
}

func (s *APIServer) createGood(w http.ResponseWriter, r *http.Request) {
	var requestGoods model.Goods

	projectID, err := strconv.ParseInt(r.URL.Query().Get("projectID"), 10, 64)
	if err != nil {
		writeErrorResponse(w, http.StatusBadRequest, "failed to read projectID")

		return
	}

	if err = json.NewDecoder(r.Body).Decode(&requestGoods); err != nil {
		writeErrorResponse(w, http.StatusBadRequest, "failed to read body")

		return
	}

	requestGoods.ProjectID = projectID

	goods, err := s.service.CreateGoods(r.Context(), requestGoods)
	if err != nil {
		zap.L().With(zap.Error(err)).Warn("createGood/s.service.CreateGoods(r.Context(), requestGoods)")
		writeErrorResponse(w, http.StatusInternalServerError, "internal server error")

		return
	}

	writeOkResponse(w, http.StatusCreated, goods)
}

func (s *APIServer) updateGoods(w http.ResponseWriter, r *http.Request) {
	var updateRequest model.UpdateGoodsRequest

	err := json.NewDecoder(r.Body).Decode(&updateRequest)
	if err != nil {
		writeErrorResponse(w, http.StatusBadRequest, "failed to read body")

		return
	}

	updateRequest.ProjectID, err = strconv.ParseInt(r.URL.Query().Get("projectId"), 10, 64)
	if err != nil {
		writeErrorResponse(w, http.StatusBadRequest, "failed to read projectID")

		return
	}

	updateRequest.ID, err = strconv.ParseInt(r.URL.Query().Get("id"), 10, 64)
	if err != nil {
		writeErrorResponse(w, http.StatusBadRequest, "failed to read projectID")

		return
	}

	goods, err := s.service.UpdateGoods(r.Context(), updateRequest)
	if err != nil {
		zap.L().With(zap.Error(err)).Warn("updateGoods/s.service.UpdateGoods(r.Context(), updateRequest)")
		writeErrorResponse(w, http.StatusInternalServerError, "internal server error")

		return
	}
	// TODO update response
	writeOkResponse(w, http.StatusOK, goods)
}

func (s *APIServer) removeGoods(w http.ResponseWriter, r *http.Request) {
	var (
		goods model.Goods
		err   error
	)

	goods.ProjectID, err = strconv.ParseInt(r.URL.Query().Get("projectId"), 10, 64)
	if err != nil {
		writeErrorResponse(w, http.StatusBadRequest, "failed to read projectID")

		return
	}

	goods.ID, err = strconv.ParseInt(r.URL.Query().Get("id"), 10, 64)
	if err != nil {
		writeErrorResponse(w, http.StatusBadRequest, "failed to read projectID")

		return
	}

	deletedGoods, err := s.service.DeleteGoods(r.Context(), goods)
	if err != nil {
		zap.L().With(zap.Error(err)).Warn("removeGoods/s.service.DeleteGoods(r.Context(), goods)")
		writeErrorResponse(w, http.StatusInternalServerError, "internal server error")

		return
	}

	writeOkResponse(w, http.StatusOK, deletedGoods)
}

func (s *APIServer) getGoods(w http.ResponseWriter, r *http.Request) {
	params, err := model.URLToGetParams(*r.URL)
	if err != nil {
		writeErrorResponse(w, http.StatusBadRequest, "error getting params")

		return
	}

	// TODO: compose response
	goods, metaData, err := s.service.GetGoods(r.Context(), *params)
	if err != nil {
		zap.L().With(zap.Error(err)).Warn("getGoods/s.service.GetGoods(r.Context(), *params)")

		writeErrorResponse(w, http.StatusInternalServerError, "internal server error")

		return
	}

	writeOkResponse(w, http.StatusOK, GetListResponse{
		Meta:  metaData,
		Goods: goods,
	})
}

func (s *APIServer) reprioritizeGood(w http.ResponseWriter, r *http.Request) {
	var (
		goods model.UpdatePriorityRequest
		err   error
	)

	err = json.NewDecoder(r.Body).Decode(&goods)
	if err != nil {
		writeErrorResponse(w, http.StatusBadRequest, "failed to read body")

		return
	}

	goods.ProjectID, err = strconv.ParseInt(r.URL.Query().Get("projectId"), 10, 64)
	if err != nil {
		writeErrorResponse(w, http.StatusBadRequest, "failed to read projectID")

		return
	}

	goods.ID, err = strconv.ParseInt(r.URL.Query().Get("id"), 10, 64)
	if err != nil {
		writeErrorResponse(w, http.StatusBadRequest, "failed to read projectID")

		return
	}

	changedGoods, err := s.service.ReprioritizeGoods(r.Context(), goods)
	if err != nil {
		zap.L().With(zap.Error(err)).Warn("reprioritizeGood/s.service.ReprioritizeGoods(r.Context(), goods)")
		writeErrorResponse(w, http.StatusInternalServerError, "internal server error")

		return
	}

	writeOkResponse(w, http.StatusOK, changedGoods)
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

func writeErrorResponse(w http.ResponseWriter, statusCode int, description string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	err := json.NewEncoder(w).Encode(description)
	if err != nil {
		zap.L().With(zap.Error(err)).Warn(
			"writeErrorResponse/json.NewEncoder(w).Encode(data)")
	}
}
