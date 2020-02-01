package logrotate

import "sync"

type Writer struct {
	// path to directory where log files will be stored
	dir string

	mux sync.Mutex
}

func (w *Writer) Write(p []byte) (n int, err error) {
	panic("implement me")
}

func (w *Writer) Close() error {
	panic("implement me")
}

func New(dir string) *Writer {
	return &Writer{dir: dir}
}
