package client

import (
	"context"

	api "github.com/bww/go-apiclient/v1"
	"github.com/bww/go-tasks/v1"
	"github.com/bww/go-tasks/v1/transport"
)

var jsonContentType = api.WithHeader("Content-Type", "application/json")

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

// Submit conforms to tasks.Publisher; it either enqueues the task or executes
// it synchronously, depending on the configuration of the client.
func (c *Client) Submit(cxt context.Context, msg *transport.Message, opts ...tasks.PublishOption) error {
	if c.sync {
		return c.Execute(cxt, msg, opts...)
	} else {
		return c.Publish(cxt, msg, opts...)
	}
}

func (c *Client) Execute(cxt context.Context, msg *transport.Message, opts ...tasks.PublishOption) error {
	conf := tasks.PublishConfig{}.WithOptions(opts)
	_, err := c.Post(cxt, "v1/tasks"+conf.Query(), msg, nil, jsonContentType)
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) Publish(cxt context.Context, msg *transport.Message, opts ...tasks.PublishOption) error {
	conf := tasks.PublishConfig{}.WithOptions(opts)
	_, err := c.Post(cxt, "v1/queue"+conf.Query(), msg, nil, jsonContentType)
	if err != nil {
		return err
	}
	return nil
}
