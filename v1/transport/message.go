package transport

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/bww/go-ident/v1"
	"github.com/bww/go-queue/v1"
	"github.com/bww/go-tasks/v1/attrs"
	"github.com/bww/go-tasks/v1/worklog"
)

var errEncodingNotSupported = errors.New("Encoding is not longer supported")

const (
	attrId   = "id"
	attrType = "type"
	attrSeq  = "seq"
	attrUTD  = "utd"
	attrMime = "mime" // message encoding
)

const (
	mimeHeader = "tasks/header" // mimeHeader is the normal header-oriented encoding
	mimeInline = "tasks/inline" // mimeInline is the MIME type for the inlined encoding format
)

type Message struct {
	Id       ident.Ident      `json:"id"`
	Seq      int64            `json:"seq"` // generally speaking, don't mess with the sequence
	Type     Type             `json:"type"`
	UTD      string           `json:"utd" check:"len(self) > 0" invalid:"Task UTD is required"`
	Data     []byte           `json:"data,omitempty"`
	Attrs    attrs.Attributes `json:"attrs,omitempty"`
	Triggers worklog.Triggers `json:"triggers,omitempty"`
}

func New(utd string) *Message {
	return &Message{
		Type: Managed,
		UTD:  utd,
	}
}

func NewWithId(id ident.Ident, utd string) *Message {
	return &Message{
		Id:   id,
		Type: Managed,
		UTD:  utd,
	}
}

func Parse(m *queue.Message) (*Message, error) {
	// if we are using inline encoding, we decode the message from the
	// queue payload; otherwise we extract headers
	if v, ok := m.Attributes[attrMime]; ok && v != mimeInline {
		return nil, fmt.Errorf("%w: Header-based attributes", errEncodingNotSupported)
	}

	c := Message{}
	err := json.Unmarshal(m.Data, &c)
	if err != nil {
		return nil, fmt.Errorf("Unable to decode inline message data: %w", err)
	}
	if c.UTD == "" {
		return nil, fmt.Errorf("%w: Payload-based UTD", errEncodingNotSupported)
	}

	return &c, nil
}

func (m *Message) SetData(d []byte) *Message {
	m.Data = d
	return m
}

func (m *Message) SetAttrs(a attrs.Attributes) *Message {
	m.Attrs = a
	return m
}

func (m *Message) SetAttr(k, v string) *Message {
	if m.Attrs == nil {
		m.Attrs = make(attrs.Attributes)
	}
	m.Attrs[k] = v
	return m
}

func (m *Message) SetTriggers(t worklog.Triggers) *Message {
	m.Triggers = t
	return m
}

func (m *Message) AddTrigger(s worklog.State, utds ...string) *Message {
	if len(utds) > 0 {
		if m.Triggers == nil {
			m.Triggers = map[worklog.State][]string{s: utds}
		} else {
			m.Triggers[s] = append(m.Triggers[s], utds...)
		}
	}
	return m
}

func (m *Message) TriggerForState(s worklog.State) (string, worklog.Triggers, bool) {
	if m.Triggers == nil {
		return "", nil, false
	} else if t := m.Triggers[s]; len(t) > 0 {
		return t[0], worklog.Triggers{s: t[1:]}, true
	} else {
		return "", m.Triggers, false
	}
}

func (m *Message) Encode() (*queue.Message, error) {
	// only inline encoding is supported now
	data, err := json.Marshal(m)
	if err != nil {
		return nil, fmt.Errorf("Could not encode message: %w", err)
	}
	return &queue.Message{
		Attributes: queue.Attributes{
			attrId:   m.Id.String(),
			attrType: m.Type.String(),
			attrMime: mimeInline,
		},
		Data: data,
	}, nil
}

func (m *Message) Entry(state worklog.State, when time.Time) *worklog.Entry {
	return &worklog.Entry{
		TaskId:   m.Id,
		UTD:      m.UTD,
		State:    state,
		Data:     m.Data,
		Attrs:    m.Attrs,
		Triggers: m.Triggers, // we retain triggers in the initial case
		Created:  when,
	}
}

func (m *Message) String() string {
	return fmt.Sprintf("<%v [%s] %v %s>", m.Id, m.UTD, m.Type, base64.StdEncoding.EncodeToString(m.Data))
}
