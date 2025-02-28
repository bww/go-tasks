package tasks

import (
	"context"
	"time"

	"github.com/bww/go-tasks/v1/transport"
	"github.com/bww/go-tasks/v1/worklog"

	"github.com/bww/go-ident/v1"
	"github.com/bww/go-queue/v1"
)

type Delivery struct {
	d   queue.Delivery
	m   *transport.Message
	err error
}

func (d Delivery) Message() (*transport.Message, error) {
	if d.err != nil {
		return nil, d.err
	} else {
		return d.m, nil
	}
}

func (d Delivery) Ack() {
	if d.d != nil {
		d.d.Ack()
	}
}

func (d Delivery) Nack() {
	if d.d != nil {
		d.d.Nack()
	}
}

// Publisher is an abstraction for publishing tasks which both [github.com/bww/go-tasks/v1.Queue]
// and [github.com/bww/go-tasks/v1/client.Client] conform to.
type Publisher interface {
	Submit(context.Context, *transport.Message, ...PublishOption) error
}

type Queue struct {
	queue.Queue
	log worklog.Worklog
}

func NewQueue(q queue.Queue, w worklog.Worklog) *Queue {
	return &Queue{q, w}
}

func (q *Queue) Worklog() worklog.Worklog {
	return q.log
}

// Submit conforms to Publisher; it has the same effect as Publish in Queue
func (q *Queue) Submit(cxt context.Context, msg *transport.Message, opts ...PublishOption) error {
	return q.Publish(cxt, msg, opts...)
}

func (q *Queue) Publish(cxt context.Context, msg *transport.Message, opts ...PublishOption) error {
	conf := PublishConfig{
		StateSeq: 0,
	}.WithOptions(opts)

	if msg.Id == ident.Zero {
		msg.Id = ident.New()
	}
	c, err := msg.Encode()
	if err != nil {
		return err
	}

	if msg.Type == transport.Managed && q.log != nil {
		ent := &worklog.Entry{
			TaskId:   msg.Id,
			TaskSeq:  msg.Seq,
			State:    worklog.Pending,
			StateSeq: conf.StateSeq,
			UTD:      msg.UTD,
			Data:     msg.Data,
			Attrs:    msg.Attrs,
			Created:  time.Now(),
		}
		var err error
		if ent.TaskSeq == 0 {
			err = q.log.CreateEntry(cxt, ent)
		} else {
			err = q.log.StoreEntry(cxt, ent)
		}
		if err != nil {
			return err
		}
	}

	return q.Queue.Publish(c)
}

func (q *Queue) Consume(cxt context.Context, name string) (<-chan Delivery, error) {
	c, err := q.Queue.Consumer(name)
	if err != nil {
		return nil, err
	}

	r := make(chan Delivery, 10)
	go func() {
		defer func() { close(r) }()
		for {
			select {
			case <-cxt.Done():
				return
			default:
				// continue
			}

			d, err := c.ReceiveWithTimeout(time.Second * 10)
			if err == queue.ErrTimeout {
				continue // we do this to catch cancellation after a reasonable period of time
			} else if err == queue.ErrClosed {
				break
			} else if err != nil {
				r <- Delivery{d: nil, err: err}
				break
			}

			var x Delivery
			m, err := transport.Parse(d.Message())
			if err != nil {
				x = Delivery{d: d, err: err}
			} else {
				x = Delivery{d: d, m: m}
			}

			r <- x
		}
	}()

	return r, nil
}
