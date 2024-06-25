package service

import (
	"github.com/bww/go-metrics/v1"
	"github.com/bww/go-tasks/v1"
	"github.com/bww/go-tasks/v1/exec"
)

type Config struct {
	Addr    string
	Secret  []byte
	Prefix  string
	Queue   *tasks.Queue
	Exec    *exec.Executor
	Metrics *metrics.Metrics
	Debug   bool
	Verbose bool
}
