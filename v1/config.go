package tasks

type PublishConfig struct {
	StateSeq int64
}

func (c PublishConfig) WithOptions(opts []PublishOption) PublishConfig {
	for _, opt := range opts {
		c = opt(c)
	}
	return c
}

type PublishOption func(PublishConfig) PublishConfig

func WithStateSeq(seq int64) PublishOption {
	return func(c PublishConfig) PublishConfig {
		c.StateSeq = seq
		return c
	}
}
