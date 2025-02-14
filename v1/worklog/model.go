package worklog

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/bww/go-ident/v1"
	"github.com/bww/go-tasks/v1/attrs"
)

type NextConfig struct {
	Attrs    attrs.Attributes
	Triggers Triggers
}

func (c NextConfig) WithOptions(opts []NextOption) NextConfig {
	for _, opt := range opts {
		c = opt(c)
	}
	return c
}

type NextOption func(NextConfig) NextConfig

func WithAttributes(v attrs.Attributes) NextOption {
	return func(c NextConfig) NextConfig {
		c.Attrs = v
		return c
	}
}

func WithTriggers(v Triggers) NextOption {
	return func(c NextConfig) NextConfig {
		c.Triggers = v
		return c
	}
}

const (
	AttrRetries = "retries"
)

type Entry struct {
	TaskId   ident.Ident
	TaskSeq  int64
	State    State
	StateSeq int64
	UTD      string
	Data     []byte
	Attrs    attrs.Attributes
	Error    json.RawMessage
	Triggers Triggers
	Retry    bool
	Created  time.Time
	Expires  *time.Time
}

func (e *Entry) Valid(when time.Time) bool {
	if e.Resolved() {
		return true // resolved status are terminal and remain valid
	} else if x := e.Expires; x != nil {
		return x.After(when) // otherwise check the expiration
	} else {
		return true // if no expiration is set, then we are valid
	}
}

func (e *Entry) Resolved() bool {
	return e.State.Resolved()
}

func (e *Entry) Clone() *Entry {
	d := *e
	return &d
}

func (e *Entry) Next(s State, d []byte, opts ...NextOption) *Entry {
	// NOTE: triggers do not inherit; do this explicitly if it's what you want
	conf := NextConfig{
		Attrs: e.Attrs,
	}.WithOptions(opts)
	sseq := e.StateSeq
	if s != e.State {
		sseq++ // increment state sequence if the state changes
	}
	return &Entry{
		TaskId:   e.TaskId,
		TaskSeq:  e.TaskSeq + 1,
		State:    s,
		StateSeq: sseq,
		UTD:      e.UTD,
		Data:     d,
		Attrs:    conf.Attrs,
		Triggers: conf.Triggers,
		Retry:    e.Retry,
		Created:  time.Now(),
	}
}

func (e *Entry) IncTaskSeq() *Entry {
	e.TaskSeq++
	return e
}

func (e *Entry) SetTaskSeq(n int64) *Entry {
	e.TaskSeq = n
	return e
}

func (e *Entry) SetStateSeq(n int64) *Entry {
	e.StateSeq = n
	return e
}

func (e *Entry) SetData(d []byte) *Entry {
	e.Data = d
	return e
}

func (e *Entry) SetAttrs(a attrs.Attributes) *Entry {
	e.Attrs = a
	return e
}

func (e *Entry) SetTriggers(t Triggers) *Entry {
	e.Triggers = t
	return e
}

func (e *Entry) SetRetry(r bool) *Entry {
	e.Retry = r
	return e
}

func (e *Entry) SetError(v []byte) *Entry {
	e.Error = v
	return e
}

func (e *Entry) SetExpires(t time.Time) *Entry {
	e.Expires = &t
	return e
}

func (e *Entry) SetCreated(t time.Time) *Entry {
	e.Created = t
	return e
}

func (e *Entry) String() string {
	return fmt.Sprintf("%v:%d", e.TaskId, e.TaskSeq)
}
