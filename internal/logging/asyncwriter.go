package logging

import (
	"fmt"
	"io"
	"sync"
	"sync/atomic"
	"time"
)

// asyncQueueDepth bounds how many formatted log records may wait for a slow
// destination before further records are dropped. Each entry is one slog record
// of a few hundred bytes, so a full queue holds on the order of a megabyte.
const asyncQueueDepth = 4096

// asyncCloseTimeout caps how long Close waits for the destination to drain. A
// destination wedged this long is not coming back, and Close must not keep the
// process from exiting.
const asyncCloseTimeout = 2 * time.Second

// AsyncWriter decouples logging from the log destination. Write copies the
// record, hands it to a background goroutine, and returns immediately; the
// goroutine is the only thing that ever calls the destination's Write. When the
// destination stops draining (a full pipe to a slow log collector) the queue
// fills and further records are dropped and counted rather than blocking.
//
// slog serializes every record through a single mutex that it holds across the
// destination Write. Without this indirection a destination that blocks would
// stall every goroutine that logs — including every in-flight request handler.
type AsyncWriter struct {
	queue     chan []byte
	stop      chan struct{}
	done      chan struct{}
	closeOnce sync.Once
	dropped   atomic.Uint64
}

// NewAsyncWriter starts a background goroutine writing to dst and returns a
// non-blocking io.Writer in front of it. Call Close to flush and stop it.
func NewAsyncWriter(dst io.Writer) *AsyncWriter {
	w := &AsyncWriter{
		queue: make(chan []byte, asyncQueueDepth),
		stop:  make(chan struct{}),
		done:  make(chan struct{}),
	}
	go w.run(dst)
	return w
}

// Write queues record for the background writer. It never blocks and never
// fails: if the queue is full the record is dropped and Dropped is incremented.
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

// Dropped reports how many records have been discarded because the destination
// could not keep up.
func (w *AsyncWriter) Dropped() uint64 {
	return w.dropped.Load()
}

// Close flushes records already queued and stops the background goroutine. It
// waits at most asyncCloseTimeout for a flush so a wedged destination cannot
// keep the process from exiting. Writes after Close are accepted but never
// flushed.
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
		// If records were dropped while a previous Write was blocked, say so in
		// the stream before the next record -- otherwise the loss is silent.
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
