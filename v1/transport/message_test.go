package transport

import (
	"testing"

	"github.com/bww/go-queue/v1"
	"github.com/stretchr/testify/assert"
)

func TestEncodeMessage(t *testing.T) {
	tests := []struct {
		Name   string
		Input  *Message
		Expect *queue.Message
		Error  error
	}{
		{
			Name: "No data, normal encoding",
			Input: &Message{
				Type: Oneshot,
				UTD:  "example://utd",
			},
			Expect: &queue.Message{
				Attributes: queue.Attributes{
					attrId:   "00000000000000000000",
					attrType: Oneshot.String(),
					attrUTD:  "example://utd",
					attrSeq:  "0",
					attrMime: mimeHeader,
				},
			},
		},
		{
			Name: "With data, normal encoding",
			Input: &Message{
				Type: Oneshot,
				UTD:  "example://utd",
				Data: []byte("Got your data right here..."),
			},
			Expect: &queue.Message{
				Attributes: queue.Attributes{
					attrId:   "00000000000000000000",
					attrType: Oneshot.String(),
					attrUTD:  "example://utd",
					attrSeq:  "0",
					attrMime: mimeHeader,
				},
				Data: []byte("Got your data right here..."),
			},
		},
		{
			Name: "UTD too long, no data",
			Input: &Message{
				Type: Oneshot,
				UTD:  "example://utd?a=AAAAAAAAAAAAAAAAAAAA&b=BBBBBBBBBBBBBBBBBBBB&c=CCCCCCCCCCCCCCCCCCCC&d=DDDDDDDDDDDDDDDDDDDD&e=EEEEEEEEEEEEEEEEEEEE",
			},
			Expect: &queue.Message{
				Attributes: queue.Attributes{
					attrMime: "tasks/inline",
				},
				Data: []byte(`{"id":null,"seq":0,"type":"oneshot","utd":"example://utd?a=AAAAAAAAAAAAAAAAAAAA\u0026b=BBBBBBBBBBBBBBBBBBBB\u0026c=CCCCCCCCCCCCCCCCCCCC\u0026d=DDDDDDDDDDDDDDDDDDDD\u0026e=EEEEEEEEEEEEEEEEEEEE"}`),
			},
		},
		{
			Name: "UTD too long, with data",
			Input: &Message{
				Type: Oneshot,
				UTD:  "example://utd?a=AAAAAAAAAAAAAAAAAAAA&b=BBBBBBBBBBBBBBBBBBBB&c=CCCCCCCCCCCCCCCCCCCC&d=DDDDDDDDDDDDDDDDDDDD&e=EEEEEEEEEEEEEEEEEEEE",
				Data: []byte("Got your data right here..."),
			},
			Expect: &queue.Message{
				Attributes: queue.Attributes{
					attrMime: "tasks/inline",
				},
				Data: []byte(`{"id":null,"seq":0,"type":"oneshot","utd":"example://utd?a=AAAAAAAAAAAAAAAAAAAA\u0026b=BBBBBBBBBBBBBBBBBBBB\u0026c=CCCCCCCCCCCCCCCCCCCC\u0026d=DDDDDDDDDDDDDDDDDDDD\u0026e=EEEEEEEEEEEEEEEEEEEE","data":"R290IHlvdXIgZGF0YSByaWdodCBoZXJlLi4u"}`),
			},
		},
	}
	for i, e := range tests {
		enc, err := e.Input.encode(100)
		if e.Error != nil {
			assert.ErrorIs(t, err, e.Error)
			continue
		} else if assert.NoError(t, err, "#%d: %s (enc)", i, e.Name) {
			if !assert.Equal(t, e.Expect, enc, "#%d: %s (enc)", i, e.Name) {
				continue
			}
		}

		dec, err := Parse(enc)
		if assert.NoError(t, err, "#%d: %s (dec)", i, e.Name) {
			if !assert.Equal(t, e.Input, dec, "#%d: %s (dec)", i, e.Name) {
				continue
			}
		}
	}
}
