package serverkit

import (
	"errors"
	"fmt"

	"github.com/hashicorp/go-multierror"
	//"gitlab.qinjian.com/tech-server/q-gateway/lib/log"
	"sync"
	"time"
)

var (
	ErrStartTimeout = errors.New("lifecycle: start timeout")
	ErrStopTimeout  = errors.New("lifecycle: stop timeout")
	startTimeout    = 5 * time.Second
	stopTimeout     = 5 * time.Second
)

const (
	svcStInit int = iota
	svcStRunning
	svcStFailed
	svcStStopped
)

type Service interface {
	Start() error
	Stop()
}

type svcWithID interface {
	ServiceID() string
}

type Context struct {
	sync.Mutex
	startAsync bool
	svcs       map[string]*svcHandle
}

type svcHandle struct {
	sync.Mutex
	id       string
	svc      Service
	state    int
	startErr error // start error
	stopErr  error
}

func NewContext() *Context {
	return &Context{
		startAsync: false,
		svcs:       make(map[string]*svcHandle),
	}
}

func (c *Context) Register(svc Service) {
	var id string
	if sid, ok := svc.(svcWithID); ok {
		id = sid.ServiceID()
	} else {
		id = fmt.Sprintf("%s", svc)
	}

	c.svcs[id] = &svcHandle{id: id, svc: svc, state: svcStInit}
}

func (c *Context) Start() error {
	c.Lock()
	defer c.Unlock()

	wg := &sync.WaitGroup{}
	wg.Add(len(c.svcs))

	for _, svc := range c.svcs {
		if c.startAsync {
			go start(svc, wg)
		} else {
			start(svc, wg)
		}
	}

	wg.Wait()
	var err *multierror.Error
	for _, svc := range c.svcs {
		if svc.startErr != nil {
			err = multierror.Append(err, svc.startErr)
			//log.Warn("[lifecycle] Failed to start service", "service", id, "error", svc.startErr)
		}
	}

	return err.ErrorOrNil()
}

func (c *Context) Stop() error {
	//log.Info("[lifecycle] Stopping")
	c.Lock()
	defer c.Unlock()

	wg := &sync.WaitGroup{}

	for _, svc := range c.svcs {
		if svc.state == svcStRunning {
			wg.Add(1)
			go stop(svc, wg)
		}
	}

	wg.Wait()
	var err *multierror.Error
	for _, svc := range c.svcs {
		if svc.stopErr != nil {
			err = multierror.Append(err, svc.stopErr)
			//log.Warn("[lifecycle] Failed to stop service", "service", id, "error", svc.startErr)
		}
	}

	return err
}

func start(h *svcHandle, wg *sync.WaitGroup) {
	//log.Debug("[lifecycle] Starting service", "service", h.id)
	defer wg.Done()

	h.Lock()
	defer h.Unlock()

	ch := make(chan error, 1)
	go func() {
		ch <- h.svc.Start()
	}()

	select {
	case h.startErr = <-ch:
		if h.startErr == nil {
			h.state = svcStRunning
			//log.Info("[lifecycle] Service started", "id", h.id)
		} else {
			h.state = svcStFailed
		}
	case <-time.After(startTimeout):
		h.startErr = ErrStartTimeout
		h.state = svcStFailed
	}
}

func stop(h *svcHandle, wg *sync.WaitGroup) {
	defer wg.Done()

	h.Lock()
	defer h.Unlock()

	timeout := time.After(startTimeout)
	ch := make(chan error, 1)

	go func() {
		h.svc.Stop()
		ch <- nil
	}()

	select {
	case <-ch:
		h.state = svcStStopped
	case <-timeout:
		h.stopErr = ErrStopTimeout
		h.state = svcStFailed
	}
}
