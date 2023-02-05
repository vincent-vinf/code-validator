package mq

import (
	"fmt"
	"github.com/streadway/amqp"
	"github.com/vincent-vinf/code-validator/pkg/util/config"
)

const (
	queueName    = "code"
	exchangeName = "code-direct"

	prefetchCnt = 2
)

type Client struct {
	*PubClient
	q amqp.Queue
}

type PubClient struct {
	conn *amqp.Connection
	ch   *amqp.Channel
}

func NewPubClient(cfg config.RabbitMQ) (*PubClient, error) {
	c := &PubClient{}
	var err error
	c.conn, err = amqp.Dial(fmt.Sprintf("amqp://%s:%s@%s:%s/", cfg.User, cfg.Passwd, cfg.Host, cfg.Port))
	if err != nil {
		return nil, err
	}

	c.ch, err = c.conn.Channel()
	if err != nil {
		return nil, err
	}

	if err = c.ch.Qos(prefetchCnt, 0, false); err != nil {
		return nil, err
	}

	if err = c.ch.ExchangeDeclare(
		exchangeName,
		"direct",
		true,
		false,
		false,
		false,
		nil,
	); err != nil {
		return nil, err
	}

	return c, nil
}

func NewClient(routeKey string, cfg config.RabbitMQ) (*Client, error) {
	var err error
	c := &Client{}
	c.PubClient, err = NewPubClient(cfg)
	if err != nil {
		return nil, err
	}

	c.q, err = c.ch.QueueDeclare(
		fmt.Sprintf("%s-%s", queueName, routeKey),
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return nil, err
	}

	if err = c.ch.QueueBind(
		c.q.Name,
		routeKey,
		exchangeName,
		false,
		nil,
	); err != nil {
		return nil, err
	}

	return c, nil
}

func (c *Client) Close() {
	if c.ch != nil {
		c.ch.Close()
	}
	if c.conn != nil {
		c.conn.Close()
	}
}

func (c *PubClient) Publish(routeKey string, data []byte) error {
	return c.ch.Publish(
		exchangeName,
		routeKey,
		false, // mandatory
		false, // immediate
		amqp.Publishing{
			ContentType: "application/octet-stream",
			Body:        data,
		})
}

func (c *Client) Consume(f func(data []byte)) error {
	msgs, err := c.ch.Consume(
		c.q.Name,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return err
	}

	for msg := range msgs {
		go func(msg amqp.Delivery) {
			f(msg.Body)
			_ = msg.Ack(false)
		}(msg)
	}

	return fmt.Errorf("mq channel closed abnormally")
}
