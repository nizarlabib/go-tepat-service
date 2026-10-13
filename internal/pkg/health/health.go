package health

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"runtime"
	"sync"
	"time"
)

// New instantiates and build new health check container
func New(opts ...Option) (*Health, error) {
	h := &Health{
		checks:        make(map[string]Config),
		maxConcurrent: runtime.NumCPU(),
	}

	for _, o := range opts {
		if err := o(h); err != nil {
			return nil, err
		}
	}

	return h, nil
}

// Register registers a check config to be performed.
func (h *Health) Register(c Config) error {
	if c.Timeout == 0 {
		c.Timeout = time.Second * 2
	}

	if c.Name == "" {
		return errors.New("health check must have a name to be registered")
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	if _, ok := h.checks[c.Name]; ok {
		return fmt.Errorf("health check %q is already registered", c.Name)
	}

	h.checks[c.Name] = c

	return nil
}

// Handler returns an HTTP handler (http.HandlerFunc).
func (h *Health) Handler() http.Handler {
	return http.HandlerFunc(h.HandlerFunc)
}

// HandlerFunc is the HTTP handler function.
func (h *Health) HandlerFunc(w http.ResponseWriter, r *http.Request) {
	c := h.Measure(r.Context())

	w.Header().Set("Content-Type", "application/json")
	data, err := json.Marshal(c)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	code := http.StatusOK
	if c.Status == StatusUnavailable {
		code = http.StatusServiceUnavailable
	}
	w.WriteHeader(code)
	w.Write(data)
}

// Measure runs all the registered health checks and returns summary status
func (h *Health) Measure(ctx context.Context) Check {
	h.mu.Lock()
	defer h.mu.Unlock()

	status := StatusOK
	failures := make(map[string]string)
	checksStatus := make(map[string]CheckStatus)

	limiterCh := make(chan bool, h.maxConcurrent)
	defer close(limiterCh)

	var (
		wg sync.WaitGroup
		mu sync.Mutex
	)
	for _, c := range h.checks {
		limiterCh <- true
		wg.Add(1)

		startTime := time.Now()
		checksStatus[c.Name] = CheckStatus{
			Name:    c.Name,
			Status:  StatusOK,
			Latency: 0,
		}

		go func(c Config) {
			defer func() {
				<-limiterCh
				wg.Done()
			}()

			resCh := make(chan error, 1)

			go func() {
				resCh <- c.Check(ctx)
				defer close(resCh)
			}()

			timeout := time.NewTimer(c.Timeout)

			select {
			case <-timeout.C:
				mu.Lock()
				defer mu.Unlock()

				failures[c.Name] = string(StatusTimeout)
				status = getAvailability(status, c.SkipOnErr)
				checksStatus[c.Name] = getCheckStatus(checksStatus[c.Name], startTime, failures[c.Name])
			case res := <-resCh:
				if !timeout.Stop() {
					<-timeout.C
				}

				mu.Lock()
				defer mu.Unlock()

				if res != nil {
					failures[c.Name] = res.Error()
					status = getAvailability(status, c.SkipOnErr)
					checksStatus[c.Name] = getCheckStatus(checksStatus[c.Name], startTime, failures[c.Name])
				} else {
					checksStatus[c.Name] = getCheckStatus(checksStatus[c.Name], startTime, "")
				}
			}
		}(c)
	}

	wg.Wait()

	var systemMetrics *System
	if h.systemInfoEnabled {
		systemMetrics = newSystemMetrics()
	}

	if !h.outputChecksEnabled {
		checksStatus = nil
	}

	return newCheck(h.component, status, systemMetrics, failures, checksStatus)
}

func newCheck(c Component, s Status, system *System, failures map[string]string, checks map[string]CheckStatus) Check {
	return Check{
		Status:    s,
		Timestamp: time.Now(),
		Failures:  failures,
		System:    system,
		Component: c,
		Checks:    checks,
	}
}

func newSystemMetrics() *System {
	s := runtime.MemStats{}
	runtime.ReadMemStats(&s)

	return &System{
		Version:          runtime.Version(),
		GoroutinesCount:  runtime.NumGoroutine(),
		TotalAllocBytes:  int(s.TotalAlloc),
		HeapObjectsCount: int(s.HeapObjects),
		AllocBytes:       int(s.Alloc),
	}
}

func getAvailability(s Status, skipOnErr bool) Status {
	if skipOnErr && s != StatusUnavailable {
		return StatusPartiallyAvailable
	}

	return StatusUnavailable
}

func getCheckStatus(status CheckStatus, startTime time.Time, err string) CheckStatus {
	status.Latency = time.Since(startTime).Milliseconds()
	if err != "" {
		status.Status = StatusUnavailable
	} else {
		status.Status = StatusOK
	}
	return status
}
