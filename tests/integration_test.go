package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"os/signal"
	"syscall"
	"testing"

	"github.com/Saaghh/hezzl-hr/internal/apiserver"
	"github.com/Saaghh/hezzl-hr/internal/config"
	"github.com/Saaghh/hezzl-hr/internal/logger"
	"github.com/Saaghh/hezzl-hr/internal/model"
	"github.com/Saaghh/hezzl-hr/internal/service"
	"github.com/Saaghh/hezzl-hr/internal/store/nats"
	"github.com/Saaghh/hezzl-hr/internal/store/pg"
	"github.com/Saaghh/hezzl-hr/internal/store/rdb"
	"github.com/google/go-querystring/query"
	migrate "github.com/rubenv/sql-migrate"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
)

const (
	bindAddr         = "http://localhost:8080/api/v1"
	createEndpoint   = "/good/create"
	updateEndpoint   = "/good/update"
	deleteEndpoint   = "/good/remove"
	getEndpoint      = "/good/list"
	priorityEndpoint = "/good/reprioritize"
)

type QueryRequestParams struct {
	ID        int64 `url:"id,omitempty"`
	ProjectID int64 `url:"projectId"`
}

type QueryListParams struct {
	Offset int `url:"offset"`
	Limit  int `url:"limit"`
}

type IntegrationTestSuite struct {
	suite.Suite
	store             *pg.Postgres
	standardProjectID int64
}

func TestIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}

func (s *IntegrationTestSuite) SetupSuite() {
	ctx, _ := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	cfg := config.New()

	logger.InitLogger(logger.Config{Level: cfg.LogLevel})

	// no error handling for now
	// check https://github.com/uber-go/zap/issues/991
	//nolint: errcheck
	defer zap.L().Sync()

	pgStore, err := pg.New(ctx, cfg)
	if err != nil {
		s.Require().NoError(err)
	}

	s.store = pgStore

	if err = pgStore.Migrate(migrate.Up); err != nil {
		s.Require().NoError(err)
	}

	zap.L().Info("successful postgres migration")

	cashdb := rdb.New(cfg)

	mb, err := nats.NewPublisher()

	serviceLayer := service.New(pgStore, cashdb, mb)

	project, err := serviceLayer.CreateProject(ctx, model.Project{Name: "Первая запись"})
	s.Require().NoError(err)

	s.standardProjectID = project.ID

	server := apiserver.New(
		apiserver.Config{BindAddress: cfg.BindAddress},
		serviceLayer)

	go func() {
		err = server.Run(ctx)
		s.Require().NoError(err)
	}()
}

func (s *IntegrationTestSuite) TearDownSuite() {
	err := s.store.TruncateAllTables(context.Background())
	s.Require().NoError(err)
}

func (s *IntegrationTestSuite) TestAPI() {
	var goods1, goods2, goods3 *model.Goods

	s.Run("empty list", func() {
	})

	s.Run("POST:/goods/create", func() {
		goods1 = s.createGood("first good")
		goods2 = s.createGood("second good")
		goods3 = s.createGood("third good")

		s.Require().NotNil(goods1)
		s.Require().NotNil(goods2)
		s.Require().NotNil(goods3)
	})

	s.Run("PATCH:/good/update", func() {
		s.Run("400, empty name", func() {
			goods1.Name = ""

			resp := s.sendRequest(
				context.Background(),
				http.MethodPatch,
				updateEndpoint,
				goods1,
				nil,
				QueryRequestParams{
					ID:        goods1.ID,
					ProjectID: goods1.ProjectID,
				})

			s.Require().Equal(http.StatusBadRequest, resp.StatusCode)
		})

		s.Run("404", func() {
			var responseData apiserver.ErrorResponse

			type nameUpdateRequest struct {
				Name string `json:"name"`
			}

			resp := s.sendRequest(
				context.Background(),
				http.MethodPatch,
				updateEndpoint,
				nameUpdateRequest{Name: "any"},
				&responseData,
				QueryRequestParams{
					ID:        -1,
					ProjectID: -1,
				})

			s.Require().Equal(http.StatusNotFound, resp.StatusCode)
			s.Require().Equal(3, responseData.Code)
			s.Require().Equal("errors.good.notFound", responseData.Message)
		})

		s.Run("200, no desc", func() {
			type nameUpdateRequest struct {
				Name string `json:"name"`
			}

			newName := "better name"

			resp := s.sendRequest(
				context.Background(),
				http.MethodPatch,
				updateEndpoint,
				nameUpdateRequest{Name: newName},
				&goods1,
				QueryRequestParams{
					ID:        goods1.ID,
					ProjectID: goods1.ProjectID,
				})

			s.Require().Equal(http.StatusOK, resp.StatusCode)
			s.Require().Equal(newName, goods1.Name)
		})

		s.Run("200, desc", func() {
			newName := "best name"
			newDesc := "cool desc"

			goods1.Name = newName
			goods1.Description = newDesc

			resp := s.sendRequest(
				context.Background(),
				http.MethodPatch,
				updateEndpoint,
				goods1,
				&goods1,
				QueryRequestParams{
					ID:        goods1.ID,
					ProjectID: goods1.ProjectID,
				})

			s.Require().Equal(http.StatusOK, resp.StatusCode)
			s.Require().Equal(newName, goods1.Name)
			s.Require().Equal(newDesc, goods1.Description)
		})
	})

	s.Run("DELETE:/good/delete", func() {
		s.Run("404", func() {
			var responseData apiserver.ErrorResponse

			resp := s.sendRequest(
				context.Background(),
				http.MethodDelete,
				deleteEndpoint,
				nil,
				&responseData,
				QueryRequestParams{
					ID:        -1,
					ProjectID: s.standardProjectID,
				})

			s.Require().Equal(http.StatusNotFound, resp.StatusCode)
			s.Require().Equal(3, responseData.Code)
			s.Require().Equal("errors.good.notFound", responseData.Message)
		})

		s.Run("200", func() {
			var deletedGood model.Goods
			resp := s.sendRequest(
				context.Background(),
				http.MethodDelete,
				deleteEndpoint,
				nil,
				&deletedGood,
				QueryRequestParams{
					ID:        goods1.ID,
					ProjectID: goods1.ProjectID,
				})

			s.Require().Equal(http.StatusOK, resp.StatusCode)
			s.Require().Equal(goods1.ID, deletedGood.ID)
			s.Require().Equal(goods1.ProjectID, deletedGood.ProjectID)
			s.Require().Equal(true, deletedGood.Removed)
		})
	})

	s.Run("GET:/good/list", func() {
		var responseData apiserver.GetListResponse

		resp := s.sendRequest(
			context.Background(),
			http.MethodGet,
			getEndpoint,
			nil,
			&responseData,
			QueryListParams{
				Offset: 0,
				Limit:  10,
			},
		)

		s.Require().Equal(http.StatusOK, resp.StatusCode)
		s.Require().Equal(3, responseData.Meta.Total)
		s.Require().Equal(1, responseData.Meta.Removed)
		s.Require().Equal(2, cap(*responseData.Goods))
	})

	s.Run("PATCH:/good/reprioritize", func() {
		var responseData apiserver.ReprioritizeResponse

		resp := s.sendRequest(
			context.Background(),
			http.MethodPatch,
			priorityEndpoint,
			model.UpdatePriorityRequest{Priority: 1},
			&responseData,
			QueryRequestParams{
				ID:        goods2.ID,
				ProjectID: goods2.ProjectID,
			})

		s.Require().Equal(http.StatusOK, resp.StatusCode)
		s.Require().Greater(len(*responseData.Priorities), 1)
	})
}

func (s *IntegrationTestSuite) createGood(name string) *model.Goods {
	var goods model.Goods

	resp := s.sendRequest(
		context.Background(),
		http.MethodPost,
		createEndpoint,
		model.Goods{Name: name},
		&goods,
		QueryRequestParams{ProjectID: s.standardProjectID},
	)

	s.Require().Equal(http.StatusCreated, resp.StatusCode)
	s.Require().NotEqual(0, goods.ID)
	s.Require().Equal(name, goods.Name)

	return &goods
}

func (s *IntegrationTestSuite) sendRequest(ctx context.Context, method, endpoint string, body interface{}, dest interface{}, params any) *http.Response {
	s.T().Helper()

	reqBody, err := json.Marshal(body)
	s.Require().NoError(err)

	queryParamsValues, err := query.Values(params)
	s.Require().NoError(err)

	req, err := http.NewRequestWithContext(ctx, method, bindAddr+endpoint+"?"+queryParamsValues.Encode(), bytes.NewReader(reqBody))
	s.Require().NoError(err)

	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	s.Require().NoError(err)

	defer func() {
		err = resp.Body.Close()
		s.Require().NoError(err)
	}()

	if dest != nil {
		err = json.NewDecoder(resp.Body).Decode(&dest)
		s.Require().NoError(err)
	}

	return resp
}
