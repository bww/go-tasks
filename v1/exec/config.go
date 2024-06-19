package exec

import (
	"log/slog"
	"time"

	"github.com/bww/go-tasks/v1"
	"github.com/bww/go-tasks/v1/worklog"
)

type Config struct {
	Queue        *tasks.Queue
	Worklog      worklog.Worklog
	Subscription string
	Concurrency  int
	EntryTTL     time.Duration // how long are non-terminal entries valid until they expire?
	Logger       *slog.Logger
	Debug        bool
	Verbose      bool
}

func (c Config) WithOptions(opts []Option) Config {
	for _, opt := range opts {
		c = opt(c)
	}
	return c
}

type Option func(Config) Config

func WithQueue(v *tasks.Queue) Option {
	return func(c Config) Config {
		c.Queue = v
		return c
	}
}

func WithWorklog(v worklog.Worklog) Option {
	return func(c Config) Config {
		c.Worklog = v
		return c
	}
}

func WithSubscription(v string) Option {
	return func(c Config) Config {
		c.Subscription = v
		return c
	}
}

func WithConcurrency(v int) Option {
	return func(c Config) Config {
		c.Concurrency = v
		return c
	}
}

func WithEntryTTL(v time.Duration) Option {
	return func(c Config) Config {
		c.EntryTTL = v
		return c
	}
}

func WithLogger(v *slog.Logger) Option {
	return func(c Config) Config {
		c.Logger = v
		return c
	}
}
