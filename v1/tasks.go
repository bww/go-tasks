package tasks

import (
	"context"
	"fmt"
	"log/slog"
	"net/url"

	"github.com/dustin/go-humanize"
)

func Run(node string, run uint64) string {
	return fmt.Sprintf("%s:%d", node, run)
}

type Request struct {
	Run    string // the execution run identifier
	UTD    *url.URL
	Entity []byte
}

func NewRequest(utd *url.URL) *Request {
	return &Request{UTD: utd}
}

func (r *Request) WithRun(node string, run uint64) *Request {
	d := *r
	d.Run = Run(node, run)
	return &d
}

func (r *Request) WithEntity(data []byte) *Request {
	d := *r
	d.Entity = data
	return &d
}

func (r *Request) Logger(base *slog.Logger) *slog.Logger {
	return base.With("run", r.Run, "utd", r.UTD.String(), "size", humanize.Bytes(uint64(len(r.Entity))))
}

type Result struct {
	State []byte `json:"state"`
}

func (r Result) WithState(b []byte) Result {
	return Result{State: b}
}

type Params struct {
	Vars map[string]string
}

type Task interface {
	Exec(context.Context, *Request, Params) (Result, error)
}

type TaskFunc func(context.Context, *Request, Params) (Result, error)

func (f TaskFunc) Exec(cxt context.Context, req *Request, params Params) (Result, error) {
	return f(cxt, req, params)
}
