package tasks

import (
	"net/url"
	"strconv"
)

type PublishConfig struct {
	StateSeq int64
}

func PublishConfigFromParams(params url.Values) (PublishConfig, error) {
	var c PublishConfig
	if v := params.Get("state_seq"); v != "" {
		x, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return c, err
		}
		c.StateSeq = x
	}
	return c, nil
}

func (c PublishConfig) WithOptions(opts []PublishOption) PublishConfig {
	for _, opt := range opts {
		c = opt(c)
	}
	return c
}

func (c PublishConfig) Params() url.Values {
	params := make(url.Values)
	if c.StateSeq != 0 {
		params.Set("state_seq", strconv.FormatInt(c.StateSeq, 10))
	}
	return params
}

func (c PublishConfig) Query() string {
	params := c.Params()
	if len(params) > 0 {
		return "?" + params.Encode()
	} else {
		return ""
	}
}

type PublishOption func(PublishConfig) PublishConfig

func UseConfig(v PublishConfig) PublishOption {
	return func(c PublishConfig) PublishConfig {
		return v // replace config
	}
}

func WithStateSeq(seq int64) PublishOption {
	return func(c PublishConfig) PublishConfig {
		c.StateSeq = seq
		return c
	}
}
