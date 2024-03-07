package nats

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"sync"

	"github.com/Saaghh/hezzl-hr/internal/model"
	"github.com/nats-io/nats.go"
	"go.uber.org/zap"
)

type Store interface {
	SaveGoodsEvents(ctx context.Context, goods *[]model.GoodsEvent) error
}

type Subscriber struct {
	conn      *nats.Conn
	queue     []model.GoodsEvent
	store     Store
	queueSize int
	mu        *sync.Mutex
}

func NewGoodsEventSubscriber(store Store, queueSize int, bindAddr string) (*Subscriber, error) {
	conn, err := nats.Connect(bindAddr)
	if err != nil {
		return nil, fmt.Errorf("nats.Connect(nats.DefaultURL): %w", err)
	}

	return &Subscriber{
		queue:     make([]model.GoodsEvent, 0, queueSize),
		conn:      conn,
		store:     store,
		queueSize: queueSize,
		mu:        new(sync.Mutex),
	}, nil
}

func (s *Subscriber) SubscribeEventLogger(subject string) (*nats.Subscription, error) {
	sub, err := s.conn.Subscribe(subject, s.processEvent)
	if err != nil {
		return nil, fmt.Errorf("nc.Subscribe(subject, processEvent): %w", err)
	}

	return sub, nil
}

func (s *Subscriber) processEvent(m *nats.Msg) {
	zap.L().Debug("processing event")

	var goods model.GoodsEvent
	if err := json.Unmarshal(m.Data, &goods); err != nil {
		zap.L().With(zap.Error(err)).Warn("processEvent/json.Unmarshal(m.Data, &goods)", zap.String("msg", string(m.Data)))
	}

	s.mu.Lock()

	s.queue = append(s.queue, goods)

	if len(s.queue) >= s.queueSize {
		if err := s.flushQueue(); err != nil {
			zap.L().With(zap.Error(err)).Warn("processEvent/s.flushQueue()", zap.String("events", strconv.Itoa(len(s.queue))))
		}
	}

	s.mu.Unlock()
}

func (s *Subscriber) flushQueue() error {
	if err := s.store.SaveGoodsEvents(context.Background(), &s.queue); err != nil {
		return fmt.Errorf("s.store.SaveGoodsEvents(s.ctx, &s.queue): %w", err)
	}

	s.queue = s.queue[:0]

	return nil
}
