package logging_test

import (
	"bytes"
	"sync"
	"testing"
	"time"

	"github.com/larssonoliver/inundated/internal/logging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type writerFunc func([]byte) (int, error)

func (f writerFunc) Write(p []byte) (int, error) { return f(p) }

// blockingWriter refuses to complete a Write until release is closed.
type blockingWriter struct{ release chan struct{} }

func (b *blockingWriter) Write(p []byte) (int, error) {
	<-b.release
	return len(p), nil
}

// gatedWriter blocks every Write until gate is closed, then records everything.
type gatedWriter struct {
	gate chan struct{}
	mu   sync.Mutex
	buf  bytes.Buffer
}

func (g *gatedWriter) Write(p []byte) (int, error) {
	<-g.gate
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.buf.Write(p)
}

func (g *gatedWriter) String() string {
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.buf.String()
}

func TestAsyncWriterDoesNotBlockWhenDestinationStalls(t *testing.T) {
	dst := &blockingWriter{release: make(chan struct{})}
	w := logging.NewAsyncWriter(dst)
	t.Cleanup(func() {
		close(dst.release)
		_ = w.Close()
	})

	done := make(chan struct{})
	go func() {
		for i := 0; i < 100_000; i++ {
			_, _ = w.Write([]byte("a log record\n"))
		}
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("Write blocked while the destination was stalled")
	}
}

func TestAsyncWriterFlushesQueuedRecordsInOrderOnClose(t *testing.T) {
	var mu sync.Mutex
	var got bytes.Buffer
	w := logging.NewAsyncWriter(writerFunc(func(p []byte) (int, error) {
		mu.Lock()
		defer mu.Unlock()
		return got.Write(p)
	}))

	for _, line := range []string{"one\n", "two\n", "three\n"} {
		_, _ = w.Write([]byte(line))
	}
	require.NoError(t, w.Close())

	mu.Lock()
	defer mu.Unlock()
	assert.Equal(t, "one\ntwo\nthree\n", got.String())
}

func TestAsyncWriterCountsRecordsDroppedWhenTheQueueOverflows(t *testing.T) {
	dst := &blockingWriter{release: make(chan struct{})}
	w := logging.NewAsyncWriter(dst)
	t.Cleanup(func() {
		close(dst.release)
		_ = w.Close()
	})

	const attempts = 50_000
	for i := 0; i < attempts; i++ {
		_, _ = w.Write([]byte("x\n"))
	}

	dropped := w.Dropped()
	assert.Positive(t, dropped, "a stalled destination must cause drops, not blocking")
	assert.Less(t, dropped, uint64(attempts), "the in-flight queue should still absorb some records")
}

func TestAsyncWriterReportsDroppedRecordsOnceTheDestinationRecovers(t *testing.T) {
	// Silent log loss during an incident is the worst time to lose logs, so the
	// drop count must show up in the stream, not just via Dropped().
	g := &gatedWriter{gate: make(chan struct{})}
	w := logging.NewAsyncWriter(g)

	for i := 0; i < 50_000; i++ {
		_, _ = w.Write([]byte("a record\n"))
	}
	require.Positive(t, w.Dropped(), "precondition: the stalled destination caused drops")

	close(g.gate)
	require.NoError(t, w.Close())

	assert.Contains(t, g.String(), "dropped",
		"the recovered stream must carry a notice about the dropped records")
}

func TestAsyncWriterCloseReturnsEvenWhenTheDestinationIsWedged(t *testing.T) {
	// Close runs on the process-exit path; a permanently blocked destination
	// must not keep the process from shutting down.
	dst := &blockingWriter{release: make(chan struct{})}
	t.Cleanup(func() { close(dst.release) })
	w := logging.NewAsyncWriter(dst)
	_, _ = w.Write([]byte("stuck in the destination\n"))

	returned := make(chan struct{})
	go func() {
		_ = w.Close()
		close(returned)
	}()

	select {
	case <-returned:
	case <-time.After(10 * time.Second):
		t.Fatal("Close blocked on a wedged destination")
	}
}

func TestAsyncWriterCopiesTheRecordBeforeQueueing(t *testing.T) {
	// slog hands its reusable internal buffer to Write and overwrites it on the
	// next record, so the async writer must copy rather than retain the slice.
	var mu sync.Mutex
	var got bytes.Buffer
	release := make(chan struct{})
	w := logging.NewAsyncWriter(writerFunc(func(p []byte) (int, error) {
		<-release
		mu.Lock()
		defer mu.Unlock()
		return got.Write(p)
	}))

	record := []byte("original")
	_, _ = w.Write(record)
	copy(record, []byte("SCRIBBLED"))
	close(release)
	require.NoError(t, w.Close())

	mu.Lock()
	defer mu.Unlock()
	assert.Equal(t, "original", got.String())
}
