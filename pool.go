package goPool

import (
	"errors"
	"os"
	"runtime"
	"strconv"
	"sync"
	"time"
)

var (
	ErrWorkerClosed = errors.New("no Active Jobs")
	ErrJobTimedOut  = errors.New("request Timed Out")

	once sync.Once
	pool *Pool
)

type Pool struct {
	WorkerChan  chan int
	RequestChan chan *RequestChannel
}

func Initialize(nRoutines int) *Pool {
	once.Do(func() {
		initializeCpuUtilization()
		pool = &Pool{
			WorkerChan:  make(chan int, nRoutines),
			RequestChan: make(chan *RequestChannel),
		}
		go pool.initWorker()
	})
	return pool
}

func (p *Pool) initWorker() {
	for {
		req, closed := <-p.RequestChan
		if closed {
			return
		} else {
			go p.worker(req)
			p.WorkerChan <- len(p.WorkerChan) + 1
		}
	}
}

func (p *Pool) worker(req *RequestChannel) {
	defer func() {
		workerId := <-p.WorkerChan
		req.workerId = workerId
	}()

	req.retChan <- req.worker.Process(req.input)
}

func (p *Pool) Process(input ReqPayload, f func(payload ReqPayload) RetPayload) (RetPayload, error) {
	req := RequestChannel{
		input: input,
		worker: func() Worker {
			return &InitializeWorker{
				processor: f,
			}
		}(),
		retChan: make(chan RetPayload),
	}
	p.RequestChan <- &req

	retPayload, closed := <-req.retChan
	if closed {
		return nil, ErrWorkerClosed
	}
	return retPayload, nil
}

func (p *Pool) ProcessWithExpiry(reqPayload interface{}, timeout time.Duration, f func(payload ReqPayload) RetPayload) (RetPayload, error) {
	req := RequestChannel{
		input: reqPayload,
		worker: func() Worker {
			return &InitializeWorker{
				processor: f,
			}
		}(),
		retChan: make(chan RetPayload),
	}

	tout := time.NewTimer(timeout)

	var retPayload RetPayload
	var closed bool

	select {
	case p.RequestChan <- &req:
	case <-tout.C:
		return nil, ErrJobTimedOut
	}

	select {
	case retPayload, closed = <-req.retChan:
		if closed {
			return nil, ErrWorkerClosed
		}
	case <-tout.C:
		return nil, ErrJobTimedOut
	}

	tout.Stop()
	return retPayload, nil
}

func (p *Pool) Close() {
	close(p.RequestChan)
	close(p.WorkerChan)
}

func initializeCpuUtilization() {
	n, err := strconv.Atoi(os.Getenv("MAX_CPU_UTILIZATION"))
	if err == nil && n != 0 && n < runtime.NumCPU() {
		runtime.GOMAXPROCS(n)
	}
}
