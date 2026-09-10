package logging

import (
	"fmt"
	"io"
	"sync"
	"sync/atomic"
	"time"
)

const asyncQueueDepth = 4096

const asyncCloseTimeout = 2 * time.Second

type AsyncWriter struct {
	queue     chan []byte
	stop      chan struct{}
	done      chan struct{}
	closeOnce sync.Once
	dropped   atomic.Uint64
}

func NewAsyncWriter(dst io.Writer) *AsyncWriter {
	w := &AsyncWriter{
		queue: make(chan []byte, asyncQueueDepth),
		stop:  make(chan struct{}),
		done:  make(chan struct{}),
	}
	go w.run(dst)
	return w
}

func (w *AsyncWriter) Write(record []byte) (int, error) {
	buf := make([]byte, len(record))
	copy(buf, record)

	select {
	case w.queue <- buf:
	default:
		w.dropped.Add(1)
	}
	return len(record), nil
}

func (w *AsyncWriter) Dropped() uint64 {
	return w.dropped.Load()
}

func (w *AsyncWriter) Close() error {
	w.closeOnce.Do(func() {
		close(w.stop)
		select {
		case <-w.done:
		case <-time.After(asyncCloseTimeout):
		}
	})
	return nil
}

func (w *AsyncWriter) run(dst io.Writer) {
	defer close(w.done)

	var reported uint64
	writeRecord := func(buf []byte) {
		if n := w.dropped.Load(); n != reported {
			_, _ = fmt.Fprintf(dst,
				"logging: dropped %d records; the log destination is not keeping up\n",
				n-reported)
			reported = n
		}
		_, _ = dst.Write(buf)
	}

	for {
		select {
		case buf := <-w.queue:
			writeRecord(buf)
		case <-w.stop:
			for {
				select {
				case buf := <-w.queue:
					writeRecord(buf)
				default:
					return
				}
			}
		}
	}
}
