package service

import (
	"github.com/bww/go-metrics/v1"
	"github.com/bww/go-tasks/v1"
)

type Config struct {
	Addr    string
	Secret  []byte
	Prefix  string
	Queue   *tasks.Queue
	Metrics *metrics.Metrics
	Debug   bool
	Verbose bool
}
