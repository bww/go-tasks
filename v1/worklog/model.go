package worklog

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/bww/go-ident/v1"
)

type Task struct {
	Id      ident.Ident // This must be generated in lexical descending order
	UTD     string
	Data    []byte
	Created time.Time
}

func NewTask(utd string, data []byte) *Task {
	now := time.Now()
	return &Task{
		Id:      ident.DscWithTime(now),
		UTD:     utd,
		Data:    data,
		Created: now,
	}
}

type Entry struct {
	Seq     int64
	TaskId  ident.Ident
	UTD     string
	State   State
	Data    []byte
	Error   json.RawMessage
	Retry   bool
	Created time.Time
	Expires *time.Time
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

func (e *Entry) Next(s State, d []byte) *Entry {
	return &Entry{
		Seq:     e.Seq + 1,
		TaskId:  e.TaskId,
		UTD:     e.UTD,
		State:   s,
		Data:    d,
		Retry:   e.Retry,
		Created: time.Now(),
	}
}

func (e *Entry) SetSeq(n int64) *Entry {
	e.Seq = n
	return e
}

func (e *Entry) SetData(d []byte) *Entry {
	e.Data = d
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
	return fmt.Sprintf("%v:%d", e.TaskId, e.Seq)
}
