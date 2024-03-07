package nats

import (
	"encoding/json"
	"fmt"

	"github.com/Saaghh/hezzl-hr/internal/model"
	"github.com/nats-io/nats.go"
	"go.uber.org/zap"
)

const subject = "goods_logs"

type Publisher struct {
	conn *nats.Conn
}

func NewPublisher() (*Publisher, error) {
	nc, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		return nil, fmt.Errorf("nats.Connect(nats.DefaultURL): %w", err)
	}

	return &Publisher{
		conn: nc,
	}, nil
}

func (p *Publisher) PublishEvent(event model.GoodsEvent) error {
	eventString, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("json.Marshal(event): %w", err)
	}

	err = p.conn.Publish(subject, eventString)
	if err != nil {
		return fmt.Errorf("p.conn.Publish(subject, []byte(message)): %w", err)
	}

	zap.L().Debug("successfully sent event to nats", zap.String("event", string(eventString)))

	return nil
}
