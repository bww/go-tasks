package client

import (
	api "github.com/bww/go-apiclient/v1"
)

type Config struct {
	Client *api.Client
	Sync   bool
}
