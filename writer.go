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

func DefaultFilenameFunc() string {
	return fmt.Sprintf("%s-%s.log", time.Now().UTC().Format(time.RFC3339), RandomHash(3))
}

// Options define rotation behavior
type Options struct {
	// Directory defines the directory where log files will be written to.
	// If the directory does not exist, it will be created.
	Directory string

	// MaximumFileSize defines the maximum size of each log file in bytes.
	// When MaximumFileSize == 0, no upper bound will be enforced.
	// No file will be greater than MaximumFileSize. A Write() which would
	// exceed MaximumFileSize will instead cause a new file to be created.
	MaximumFileSize int

	// MaximumLifetime defines the maximum amount of time a file will
	// be written to before a rotation occurs.
	// When MaximumLifetime == 0, no log rotation will occur.
	MaximumLifetime time.Duration

	// FileNameFunc specifies the name a new file will take.
	// FileNameFunc must ensure collisions in filenames do not occur.
	// Do not rely on timestamps to be unique, high throughput writes
	// may fall on the same timestamp.
	// Eg.
	// 	2020-03-28_15-00-945-<random-hash>.log
	// When FileNameFunc is not specified, DefaultFilenameFunc will be used.
	FileNameFunc func() string
}

type Writer struct {
	logger *log.Logger

	// opts are the configuration options for this Writer
	opts Options

	// f is the currently open file used for appends.
	// Writes to f are only synchronized once Close() is called,
	// or when files are being rotated.
	f *os.File

	// queue of entries awaiting to be written
	queue chan []byte
	// synchronize write which have started but not been queued up
	pending sync.WaitGroup
	// singal the writer should close
	closing chan struct{}
	// signal the writer has finished writing all queued up entries.
	done chan struct{}
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
			path := filepath.Join(w.opts.Directory, w.opts.FileNameFunc())
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

func New(logger *log.Logger, opts Options) (*Writer, error) {
	if _, err := os.Stat(opts.Directory); os.IsNotExist(err) {
		if err := os.MkdirAll(opts.Directory, 0644); err != nil {
			return nil, errors.Wrapf(err, "directory %v does not exist and could not be created", opts.Directory)
		}
	}

	if opts.FileNameFunc == nil {
		opts.FileNameFunc = DefaultFilenameFunc
	}

	w := &Writer{
		logger:  logger,
		opts:    opts,
		queue:   make(chan []byte, 1024),
		closing: make(chan struct{}),
		done:    make(chan struct{}),
	}

	go w.listen()

	return w, nil
}

func newFile(path string) (*os.File, error) {
	return os.OpenFile(path, os.O_CREATE|os.O_WRONLY, 0644)
}
