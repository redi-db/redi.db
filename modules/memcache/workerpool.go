package memcache

import (
	"sync"
)

type WorkerPool struct {
	taskQueue chan func()
	wg        sync.WaitGroup
	once      sync.Once
}

func NewWorkerPool(workerCount int, taskQueueSize int) *WorkerPool {
	pool := &WorkerPool{
		taskQueue: make(chan func(), taskQueueSize),
	}

	for i := 0; i < workerCount; i++ {
		go pool.worker()
	}

	return pool
}

func (wp *WorkerPool) worker() {
	for task := range wp.taskQueue {
		task()
		wp.wg.Done()
	}
}

func (wp *WorkerPool) Submit(taskFunc func()) {
	wp.wg.Add(1)
	wp.taskQueue <- taskFunc
}

func (wp *WorkerPool) Shutdown() {
	wp.once.Do(func() {
		close(wp.taskQueue)
	})

	wp.wg.Wait()
}
