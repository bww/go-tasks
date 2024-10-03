package exec

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"maps"
	"net/url"
	"os"
	"runtime/debug"
	"slices"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/bww/go-tasks/v1"
	"github.com/bww/go-tasks/v1/router"
	"github.com/bww/go-tasks/v1/transport"
	"github.com/bww/go-tasks/v1/worklog"

	"github.com/bww/go-alert/v1"
	"github.com/bww/go-ident/v1"
	"github.com/bww/go-metrics/v1"
	errutil "github.com/bww/go-util/v1/errors"
	"github.com/bww/go-util/v1/ext"
	sliceutil "github.com/bww/go-util/v1/slices"
	"github.com/bww/go-util/v1/text"
	"github.com/dustin/go-humanize"
	cmap "github.com/orcaman/concurrent-map/v2"
)

var (
	ErrStarted       = errors.New("Already started")
	ErrStopped       = errors.New("Not running")
	ErrMissingIdent  = errors.New("Missing identifier")
	ErrTaskFailed    = errors.New("Task failed")
	ErrUnsupported   = errors.New("Unsupported operation")
	ErrInvalidConfig = errors.New("Invalid configuration")
	ErrUnimplemented = errors.New("Unimplemented")
)

// the default timeout for operations
const defaultTimeout = time.Second * 10

type taskSpec struct {
	context context.Context
	cancel  context.CancelFunc
	message *transport.Message
	entry   *worklog.Entry
}

type Executor struct {
	sync.Mutex
	router.Router
	starter sync.Once
	worklog worklog.Worklog

	nodename string
	inflight cmap.ConcurrentMap[string, taskSpec]
	cn       int
	ttl      time.Duration
	queue    *tasks.Queue
	subscr   string
	log      *slog.Logger
	errs     chan error
	verbose  bool
	debug    bool
	runid    uint64

	metrics            *metrics.Metrics
	taskSuccessCounter metrics.Counter
	taskFailureCounter metrics.Counter
	taskExecSampler    metrics.Sampler
}

func New(q *tasks.Queue, s string, opts ...Option) (*Executor, error) {
	return NewWithConfig(Config{
		Queue:        q,
		Subscription: s,
	}.WithOptions(opts))
}

func NewWithConfig(conf Config) (*Executor, error) {
	if conf.Queue == nil {
		return nil, fmt.Errorf("%w: No queue provided", ErrInvalidConfig)
	}
	if conf.Subscription == "" {
		return nil, fmt.Errorf("%w: No subscription provided", ErrInvalidConfig)
	}
	if conf.Logger == nil {
		conf.Logger = slog.Default()
	}

	var (
		nodename string
		err      error
	)
	if n := conf.Nodename; n != "" {
		nodename = n
	} else {
		nodename, err = os.Hostname()
		if err != nil {
			return nil, fmt.Errorf("No node name provided and could not obtain host name: %w", err)
		}
	}

	enableVerbose := text.Coalesce(os.Getenv("VERBOSE_WORKER"), os.Getenv("VERBOSE")) != ""
	enableDebug := text.Coalesce(os.Getenv("DEBUG_WORKER"), os.Getenv("DEBUG")) != ""

	r := router.New()
	w := &Executor{
		Router:   r,
		nodename: nodename,
		inflight: cmap.New[taskSpec](),
		cn:       max(1, conf.Concurrency),
		ttl:      max(time.Minute, conf.EntryTTL), // entry TTL; must be at least a minute
		queue:    conf.Queue,
		worklog:  conf.Worklog,
		subscr:   conf.Subscription,
		log:      conf.Logger.With("system", "tasks"),
		verbose:  enableVerbose,
		debug:    enableDebug,
	}

	if w.metrics != nil {
		w.taskSuccessCounter = w.metrics.RegisterCounter("task_success", "Successful tasks", nil)
		w.taskFailureCounter = w.metrics.RegisterCounter("task_failure", "Failed tasks", nil)
		w.taskExecSampler = w.metrics.RegisterSampler("task_exec", "Task execution duration", nil)
	}

	return w, nil
}

func (w *Executor) Verbose() bool {
	w.Lock()
	defer w.Unlock()
	return w.verbose || w.debug
}

func (w *Executor) Debug() bool {
	w.Lock()
	defer w.Unlock()
	return w.debug
}

func (w *Executor) Run(cxt context.Context) error {
	w.Lock()
	cn := w.cn
	subscr := w.subscr
	w.Unlock()
	return w.run(cxt, cn, subscr)
}

func (w *Executor) nextRun() string {
	n := atomic.AddUint64(&w.runid, 1)
	return tasks.Run(w.nodename, n)
}

func (w *Executor) run(cxt context.Context, cn int, name string) error {
	log := w.log
	cxt, cancel := context.WithCancel(cxt)
	defer cancel()

	wg := &sync.WaitGroup{}
	sem := make(chan struct{}, cn)
	var inflight, total int64

	recv, err := w.queue.Consume(cxt, name)
	if err != nil {
		return err
	}

outer:
	for {
		var dlv tasks.Delivery
		var ok bool
		select {
		case <-cxt.Done():
			break outer
		case dlv, ok = <-recv:
			if !ok {
				break outer
			}
		}

		msg, err := dlv.Message()
		if err != nil {
			w.report(err)
			continue
		}

		// ack immediately to avoid repeated delivery; if this message is managed,
		// the pending state should already have been recorded in the worklog and
		// the task may be retried from there should execution fail
		dlv.Ack()

		t := atomic.AddInt64(&total, 1)
		log := msgLog(slog.Default(), msg)
		if w.Verbose() {
			f := atomic.LoadInt64(&inflight)
			log.With(
				"total", t,
				"in_flight", f,
			).Info("Received task")
		}

		sem <- struct{}{}
		wg.Add(1)
		atomic.AddInt64(&inflight, 1)

		go func(msg *transport.Message) {
			defer func() { <-sem; wg.Done(); atomic.AddInt64(&inflight, -1) }()
			now := time.Now()
			var err error
			switch msg.Type {
			case transport.Managed:
				err = w.handleManaged(cxt, msg, now)
			case transport.Oneshot:
				err = w.handleOneshot(cxt, msg, now)
			case transport.Cronjob:
				err = w.handleCronjob(cxt, msg, now)
			default:
				err = fmt.Errorf("Task type is not supported: %v", msg.Type)
			}
			if err != nil {
				err := errutil.Reference(err)
				if msg.Type == transport.Oneshot {
					logerr(log, fmt.Errorf("Task failed: %v", err))
				} else {
					alert.Error(fmt.Errorf("Task failed: %w", err), alert.WithTags(msgTags(msg)))
				}
				w.report(err)
			}
		}(msg)
	}

	log.Info(fmt.Sprintf("Waiting for %d tasks to complete...\n", atomic.LoadInt64(&inflight)))
	wg.Wait()
	return ErrStopped
}

func (w *Executor) report(err error) {
	log := w.log
	w.Lock()
	errs := w.errs
	w.Unlock()
	if errs == nil {
		log.With("cause", err).Error("Error receiving message")
	} else {
		select {
		case errs <- err:
			// error propagated to handlers
		case <-time.After(defaultTimeout):
			log.With("cause", err).Error("Timeout while propagating error; moving on")
		}
	}
}

func (w *Executor) Errors() <-chan error {
	w.Lock()
	defer w.Unlock()
	if w.errs == nil {
		w.errs = make(chan error, 1)
	}
	return w.errs
}

func (w *Executor) handleCronjob(cxt context.Context, msg *transport.Message, now time.Time) error {
	if w.worklog == nil {
		return fmt.Errorf("%w: Worklog is not available, cannot manage tasks", ErrUnsupported)
	}
	// A "cronjob" message is a managed task message that isn't fully initialized
	// because it was created by the cron service by directly enqueuing to the
	// task queue. As such it doesn't have an identifier util we assign one to
	// it.
	if msg.Id == ident.Zero {
		msg.Id = ident.New()
	}

	ent := &worklog.Entry{
		TaskId:  msg.Id,
		UTD:     msg.UTD,
		State:   worklog.Pending,
		Created: now,
	}
	err := w.worklog.CreateEntry(cxt, ent)
	if err != nil && !errors.Is(err, worklog.ErrNotFound) {
		return fmt.Errorf("Could not initialize worklog entry: %v", err)
	}

	return w.handleManaged(cxt, msg, now)
}

func (w *Executor) handleManaged(cxt context.Context, msg *transport.Message, now time.Time) error {
	if w.worklog == nil {
		return fmt.Errorf("%w: Worklog is not available, cannot manage tasks", ErrUnsupported)
	}
	if msg.Id == ident.Zero {
		return ErrMissingIdent
	}

	ent, err := w.worklog.FetchLatestEntryForTask(cxt, msg.Id)
	if err != nil && !errors.Is(err, worklog.ErrNotFound) {
		return fmt.Errorf("Could not fetch worklog entry: %v", err)
	}

	var next *worklog.Entry
	if ent != nil {
		if ent.State == worklog.Complete {
			return fmt.Errorf("Task is already completed")
		} else if ent.State == worklog.Running && ent.Valid(now) {
			return fmt.Errorf("Task is already running since: %v", ent.Created)
		}
		next = ent.NextWithAttrs(worklog.Running, msg.Data, msg.Attrs)
	} else {
		next = &worklog.Entry{
			TaskId:  msg.Id,
			UTD:     msg.UTD,
			State:   worklog.Running,
			Data:    msg.Data,
			Attrs:   msg.Attrs,
			Created: now,
		}
	}

	err = w.worklog.StoreEntry(cxt, next) // Entry must be initialized
	if err != nil {
		if ent != nil {
			return fmt.Errorf("Could not store worklog entry on run (%d → %d): %w", ent.TaskSeq, next.TaskSeq, err)
		} else {
			return fmt.Errorf("Could not store worklog entry on init: %w", err)
		}
	}

	taskId := next.TaskId.String()
	cxt, cancel := context.WithCancel(cxt)
	defer func() {
		cancel()
		w.inflight.Remove(taskId)
	}()

	w.inflight.Set(taskId, taskSpec{
		context: cxt,
		cancel:  cancel,
		message: msg,
		entry:   next,
	})

	res, err := w.Proc(cxt, msg, next)
	if err == nil {
		next = next.Next(worklog.Complete, res.State)
	} else {
		next = next.Next(stateForError(err), res.State).SetRetry(errutil.Recoverable(err))
		errdat, suberr := json.Marshal(jsonError{Err: err})
		if suberr != nil {
			alert.Error(fmt.Errorf("Could not marshal worklog error on failure: %v", suberr), alert.WithTags(msgTags(msg, alert.Tags{"task_seq": fmt.Sprint(next.TaskSeq)})))
		} else {
			next = next.SetError(errdat)
		}
	}

	// As a special case, we create a new context for storing state. if the
	// original context was canceled or timed out, we don't want that to affect
	// this operation
	subcxt, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()
	suberr := w.worklog.StoreEntry(subcxt, next)
	if suberr != nil {
		alert.Error(fmt.Errorf("Could not store worklog entry on success: %w", suberr), alert.WithTags(msgTags(msg, alert.Tags{"task_seq": fmt.Sprint(next.TaskSeq)})))
	}

	return err
}

func (w *Executor) handleOneshot(cxt context.Context, msg *transport.Message, now time.Time) error {
	_, err := w.Proc(cxt, msg, nil)
	return err
}

func (w *Executor) Proc(cxt context.Context, msg *transport.Message, ent *worklog.Entry) (res tasks.Result, err error) {
	now := time.Now()
	log := msgLog(w.log, msg)
	if ent != nil {
		log = log.With("worklog", ent.String())
	}
	if w.Verbose() {
		log.Info("Running task")
	}

	defer func() {
		if w.taskExecSampler != nil {
			w.taskExecSampler.Observe(float64(time.Since(now)))
		}
		if r := recover(); r != nil {
			err = errors.New(fmt.Sprintf("panic: %v\n%s", r, string(debug.Stack())))
		}
		if err != nil {
			if w.taskFailureCounter != nil {
				w.taskFailureCounter.Inc()
			}
		} else {
			if w.taskSuccessCounter != nil {
				w.taskSuccessCounter.Inc()
			}
		}
	}()

	u, err := url.Parse(msg.UTD)
	if err != nil {
		return res, fmt.Errorf("Invalid UTD: %w", err)
	}

	cxt, cancel := context.WithCancel(cxt)
	defer cancel()

	if ent != nil {
		go func() {
			for {
				select {
				case <-cxt.Done():
					return
				case <-time.After(w.ttl / 2):
					if w.Verbose() || w.Debug() {
						log.Debug("Renew lease", "entry", ent, "window", w.ttl)
					}
					prv := ent
					ent, err = w.worklog.RenewEntry(cxt, ent, now.Add(w.ttl))
					if err != nil {
						alert.Error(fmt.Errorf("Could not renew worklog entry: %v", err), alert.WithTags(msgTags(msg, alert.Tags{"worklog": prv.String()})))
					}
				}
			}
		}()
	}

	res, err = w.Router.Exec(cxt, &tasks.Request{
		Run:    w.nextRun(),
		UTD:    u,
		Entity: msg.Data,
	})
	if errors.Is(err, tasks.ErrUnsupported) {
		return res, err
	} else if errors.Is(err, context.Canceled) {
		return res, err
	} else if err != nil {
		return res, fmt.Errorf("Handler error: %w", err)
	}

	if w.Verbose() {
		log.Debug("Task completed", "duration", time.Since(now))
	}
	return res, nil
}

func msgLog(base *slog.Logger, msg *transport.Message) *slog.Logger {
	if base == nil {
		base = slog.Default()
	}
	alen := len(msg.Attrs)
	return base.With(
		"utd", msg.UTD,
		"task_id", msg.Id.String(),
		"task_type", msg.Type.String(),
		"task_data", humanize.Bytes(uint64(len(msg.Data))),
		"task_attrs", ext.Choose(alen > 0, sliceutil.Summary(slices.Collect(maps.Keys(msg.Attrs)), ",", "...", 3), strconv.Itoa(alen)),
	)
}

func msgTags(msg *transport.Message, merge ...alert.Tags) alert.Tags {
	alen := len(msg.Attrs)
	t := alert.Tags{
		"utd":        msg.UTD,
		"task_id":    msg.Id,
		"task_type":  msg.Type,
		"task_data":  humanize.Bytes(uint64(len(msg.Data))),
		"task_attrs": ext.Choose(alen > 0, sliceutil.Summary(slices.Collect(maps.Keys(msg.Attrs)), ",", "...", 3), strconv.Itoa(alen)),
	}
	for _, m := range merge {
		for k, v := range m {
			t[k] = v
		}
	}
	return t
}

func logerr(log *slog.Logger, err error) {
	ref := errutil.Refstr(err)
	if ref != "" {
		log = log.With("ref", ref)
	}
	log.Error(err.Error())
}
