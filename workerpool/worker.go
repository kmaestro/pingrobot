package workerpool

import (
	"log"
	"net/http"
	"sync"
	"time"
)

type worker struct {
	client *http.Client
}

func newWorker(timeout time.Duration) *worker {
	return &worker{
		&http.Client{
			Timeout: timeout,
		},
	}
}

func (w worker) process(j Job) Result {
	result := Result{URL: j.URL}

	now := time.Now()

	resp, err := w.client.Get(j.URL)
	if err != nil {
		result.Error = err

		return result
	}

	result.StatusCode = resp.StatusCode
	result.ResponseTime = time.Since(now)

	return result
}

type Pool struct {
	worker       *worker
	workersCount int

	jobs    chan Job
	results chan Result

	wg      *sync.WaitGroup
	stopped bool
}

func New(workersCount int, timeout time.Duration, results chan Result) *Pool {
	return &Pool{
		worker:       newWorker(timeout),
		workersCount: workersCount,
		jobs:         make(chan Job),
		results:      results,
		wg:           new(sync.WaitGroup),
	}
}

func (p *Pool) Init() {
	for i := 0; i < p.workersCount; i++ {
		go p.initWorker(i)
	}
}

func (p *Pool) Push(j Job) {
	if p.stopped {
		return
	}

	p.jobs <- j
	p.wg.Add(1)
}

func (p *Pool) Stop() {
	p.stopped = true
	close(p.jobs)
	p.wg.Wait()
}

func (p *Pool) initWorker(id int) {
	for job := range p.jobs {
		time.Sleep(time.Second)
		p.results <- p.worker.process(job)
		p.wg.Done()
	}

	log.Printf("[worker ID %d] finished proccesing", id)
}
