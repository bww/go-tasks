package client

import (
	"context"

	api "github.com/bww/go-apiclient/v1"
	"github.com/bww/go-tasks/v1/transport"
)

var (
	jsonContentType = api.WithHeader("Content-Type", "application/json")
)

type Client struct {
	*api.Client
	sync bool
}

func NewWithConfig(conf Config) *Client {
	return &Client{
		Client: conf.Client,
		sync:   conf.Sync,
	}
}

func (c *Client) Submit(cxt context.Context, msg *transport.Message) error {
	if c.sync {
		return c.Execute(cxt, msg)
	} else {
		return c.Publish(cxt, msg)
	}
}

func (c *Client) Execute(cxt context.Context, msg *transport.Message) error {
	_, err := c.Client.Post(cxt, "v1/tasks", msg, nil, jsonContentType)
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) Publish(cxt context.Context, msg *transport.Message) error {
	_, err := c.Client.Post(cxt, "v1/queue", msg, nil, jsonContentType)
	if err != nil {
		return err
	}
	return nil
}
