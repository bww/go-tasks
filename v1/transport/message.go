package transport

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/bww/go-tasks/v1/worklog"

	"github.com/bww/go-ident/v1"
	"github.com/bww/go-queue/v1"
)

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

const utdMaxLength = 1024

type Message struct {
	Id   ident.Ident `json:"id"`
	Seq  int64       `json:"seq"` // generally speaking, don't mess with the sequence
	Type Type        `json:"type"`
	UTD  string      `json:"utd"`
	Data []byte      `json:"data,omitempty"`
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

func NewFromTask(t *worklog.Task) *Message {
	return &Message{
		Id:   t.Id,
		Type: Managed,
		UTD:  t.UTD,
		Data: t.Data,
	}
}

func Parse(m *queue.Message) (*Message, error) {
	var err error
	c := &Message{}

	var inline bool
	if a := m.Attributes; len(a) > 0 {
		// if we are using inline encoding, we decode the message from the
		// queue payload; otherwise we extract headers
		if v, ok := a[attrMime]; ok && v == mimeInline {
			inline = true
			err = json.Unmarshal(m.Data, c)
			if err != nil {
				return nil, fmt.Errorf("Unable to decode inline data: %w", err)
			}
		} else {
			if v, ok := a[attrId]; ok {
				c.Id, err = ident.Parse(v)
				if err != nil {
					return nil, fmt.Errorf("Invalid identifier: %v", err)
				}
			}
			if v, ok := a[attrType]; ok {
				c.Type, err = ParseType(v)
				if err != nil {
					return nil, err
				}
			}
			if v, ok := a[attrSeq]; ok {
				c.Seq, err = strconv.ParseInt(v, 10, 64)
				if err != nil {
					return nil, fmt.Errorf("Invalid sequence: %v", err)
				}
			}
			if v, ok := a[attrUTD]; ok {
				c.UTD = v
			}
		}
	}

	// If no UTD is set at this point, we assume the payload is the UTD;
	// this was used to support the use of Cloud Scheduler to schedule
	// tasks before it supported configuring attributes. It should not
	// actually be used anymore, although we haven't gone through to
	// check all the configurations to ensure that is the case.
	if c.UTD == "" {
		c.UTD = string(m.Data)
	} else if !inline {
		c.Data = m.Data
	}

	return c, nil
}

func (m *Message) SetData(d []byte) *Message {
	m.Data = d
	return m
}

func (m *Message) Encode() (*queue.Message, error) {
	return m.encode(utdMaxLength)
}

func (m *Message) encode(maxlen int) (*queue.Message, error) {
	// if the UTD's length exceeds the maximum allowed length (this is a
	// limitation of PubSub's message attributes), we use the alternate
	// inline encoding.
	if len(m.UTD) > maxlen {
		data, err := json.Marshal(m)
		if err != nil {
			return nil, fmt.Errorf("Could not encode message (inline): %w", err)
		}
		return &queue.Message{
			Attributes: queue.Attributes{
				attrMime: mimeInline,
			},
			Data: data,
		}, nil
	} else {
		return &queue.Message{
			Attributes: queue.Attributes{
				attrId:   m.Id.String(),
				attrType: m.Type.String(),
				attrUTD:  m.UTD,
				attrSeq:  strconv.FormatInt(m.Seq, 10),
				attrMime: mimeHeader, // we provide this as an advisory, but it is also assumed
			},
			Data: m.Data,
		}, nil
	}
}

func (m *Message) String() string {
	return fmt.Sprintf("<%v [%s] %v %s>", m.Id, m.UTD, m.Type, base64.StdEncoding.EncodeToString(m.Data))
}
