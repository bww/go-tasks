package tasks

import (
	"context"
	"net/url"
)

type Request struct {
	UTD    *url.URL
	Entity []byte
}

func NewRequest(utd *url.URL) *Request {
	return &Request{UTD: utd}
}

func (r *Request) WithEntity(data []byte) *Request {
	d := *r
	d.Entity = data
	return &d
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
