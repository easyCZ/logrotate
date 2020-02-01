package logrotate

import (
	"fmt"
	"github.com/pkg/errors"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type Writer struct {
	logger *log.Logger
	// path to directory where log files will be stored
	dir string

	mux sync.Mutex

	queue   chan []byte
	pending sync.WaitGroup
	closing chan struct{}
	done    chan struct{}

	f *os.File
}

func (w *Writer) Write(p []byte) (n int, err error) {
	select {
	case <-w.closing:
		return 0, errors.Wrap(err, "writer is closing")
	default:
		w.pending.Add(1)
		defer w.pending.Done()
	}

	w.queue <- p

	return len(p), nil
}

func (w *Writer) Close() error {
	close(w.closing)
	w.pending.Wait()

	close(w.queue)
	<-w.done

	if w.f != nil {
		if err := w.f.Sync(); err != nil {
			return errors.Wrap(err, "failed to sync current log file")
		}

		if err := w.f.Close(); err != nil {
			return errors.Wrap(err, "failed to close current log file")
		}
	}

	return nil
}

func (w *Writer) listen() {
	for b := range w.queue {
		if w.f == nil {
			path := filepath.Join(w.dir, fmt.Sprintf("%s.log", time.Now().UTC().Format(time.RFC3339)))
			f, err := newFile(path)
			if err != nil {
				w.logger.Println(fmt.Sprintf("Failed to create new file at %v", path), err)
			}
			w.f = f
		}

		if _, err := w.f.Write(b); err != nil {
			w.logger.Println("Failed to write to file.", err)
		}
	}

	close(w.done)
}

func New(logger *log.Logger, dir string) *Writer {
	w := &Writer{
		logger:  logger,
		dir:     dir,
		queue:   make(chan []byte, 1024),
		closing: make(chan struct{}),
		done:    make(chan struct{}),
	}

	go w.listen()

	return w
}

func newFile(path string) (*os.File, error) {
	return os.OpenFile(path, os.O_CREATE|os.O_WRONLY, 0644)
}
