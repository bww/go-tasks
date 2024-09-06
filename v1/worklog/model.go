package worklog

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/bww/go-ident/v1"
	"github.com/bww/go-tasks/v1/attrs"
)

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

func (e *Entry) Next(s State, d []byte) *Entry {
	return e.NextWithAttrs(s, d, nil)
}

func (e *Entry) NextWithAttrs(s State, d []byte, a attrs.Attributes) *Entry {
	sseq := e.StateSeq
	if s != e.State {
		sseq++ // increment state sequence if the state changes
	}
	if a == nil {
		a = e.Attrs // prefer provided attributes to inherited ones
	}
	return &Entry{
		TaskId:   e.TaskId,
		TaskSeq:  e.TaskSeq + 1,
		State:    s,
		StateSeq: sseq,
		UTD:      e.UTD,
		Data:     d,
		Attrs:    a,
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
