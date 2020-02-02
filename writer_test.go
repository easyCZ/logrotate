package logrotate

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestWriter(t *testing.T) {
	logger := log.New(os.Stderr, "", log.LstdFlags)

	setup := func(t *testing.T) (string, func()) {
		dir, err := ioutil.TempDir("", "")
		require.NoError(t, err)

		cleanup := func() {
			require.NoError(t, os.RemoveAll(dir))
		}

		return dir, cleanup
	}

	t.Run("creates target directory if it does not exist", func(t *testing.T) {
		dir, cleanup := setup(t)
		defer cleanup()

		dir = filepath.Join(dir, "foo")
		w, err := New(logger, Options{
			Directory: dir,
		})
		require.NoError(t, err)
		require.NoError(t, w.Close(), "must close writer")

		f, err := os.Stat(dir)
		require.NoError(t, err)
		require.True(t, f.IsDir(), "must create directory")
	})

	t.Run("create, write, close", func(t *testing.T) {
		dir, cleanup := setup(t)
		defer cleanup()

		w, err := New(logger, Options{
			Directory: dir,
		})
		require.NoError(t, err, "must construct writer")

		message := []byte("message")
		_, err = w.Write(message)
		require.NoError(t, err, "write must succeed")
		require.NoError(t, w.Close(), "must close writer")

		files, err := ioutil.ReadDir(dir)
		require.NoError(t, err)

		require.Len(t, files, 1, "must write exactly one file")
		written, err := ioutil.ReadFile(filepath.Join(dir, files[0].Name()))
		require.NoError(t, err, "must read file")
		require.Equal(t, message, written)
	})

	t.Run("rotates on file size", func(t *testing.T) {
		dir, cleanup := setup(t)
		defer cleanup()

		max := 128
		w, err := New(logger, Options{
			Directory:       dir,
			MaximumFileSize: int64(max),
		})
		require.NoError(t, err)

		// fill up the first file
		_, err = w.Write([]byte(strings.Repeat("a", max)))
		require.NoError(t, err, "must write")

		// write more, should create a new file
		_, err = w.Write([]byte("b"))
		require.NoError(t, err)

		require.NoError(t, w.Close())

		files, err := ioutil.ReadDir(dir)
		require.NoError(t, err)

		require.Len(t, files, 2, "must produce 2 files")
	})

	t.Run("rotates on lifetime", func(t *testing.T) {
		dir, cleanup := setup(t)
		defer cleanup()

		lifetime := time.Second
		w, err := New(logger, Options{
			Directory:       dir,
			MaximumLifetime: lifetime,
		})
		require.NoError(t, err)

		// keep writing until lifetime + half of lifetime (middle of ticks) elapses
		end := time.Now().Add(lifetime + lifetime/2)
		for time.Now().Before(end) {
			_, err = w.Write([]byte("message"))
			require.NoError(t, err)
		}

		require.NoError(t, w.Close())
		files, err := ioutil.ReadDir(dir)
		require.NoError(t, err)
		require.Len(t, files, 2, "should produce 2 files")
	})

	t.Run("concurrent writes", func(t *testing.T) {
		dir, cleanup := setup(t)
		defer cleanup()

		w, err := New(logger, Options{
			Directory: dir,
		})
		require.NoError(t, err)
		rows := 10000
		writers := 10
		messageSize := 10

		var wg sync.WaitGroup
		for i := 0; i < writers; i++ {
			wg.Add(1)
			go func(i int) {
				for j := 0; j < rows; j++ {
					_, err := w.Write([]byte(strings.Repeat(fmt.Sprintf("%d", i), messageSize)))
					require.NoError(t, err)
				}
				wg.Done()
			}(i)
		}

		wg.Wait()
		require.NoError(t, w.Close(), "must close")

		files, err := ioutil.ReadDir(dir)
		require.NoError(t, err)
		require.Len(t, files, 1, "must write a single file")
		require.Equal(t, int64(rows*writers*messageSize), files[0].Size(), "must write all bytes")
	})
}

func benchmarkWriter(b *testing.B, messages int, messageSize int, writers int) {
	logger := log.New(os.Stderr, "", log.LstdFlags)

	dir, err := ioutil.TempDir("", "")
	if err != nil {
		b.Fatalf("err: %v", err)
	}
	defer os.RemoveAll(dir)

	for n := 0; n < b.N; n++ {
		w, err := New(logger, Options{
			Directory: dir,
		})
		if err != nil {
			b.Fatalf("err: %v", err)
		}

		var wg sync.WaitGroup
		for i := 0; i < writers; i++ {
			wg.Add(1)
			go func(i int) {
				for j := 0; j < messages; j++ {
					_, err := w.Write([]byte(strings.Repeat(fmt.Sprintf("%d", i), messageSize)))
					if err != nil {
						b.Fatalf("err: %v", err)
					}
				}
				wg.Done()
			}(i)
		}

		wg.Wait()
		if err := w.Close(); err != nil {
			b.Fatalf("err: %v", err)
		}
	}
}

func Benchmark_1000Messages_100BytesPerMessage_1Writer(b *testing.B) {
	benchmarkWriter(b, 1000, 100, 1)
}

func Benchmark_1000Messages_100BytesPerMessage_2Writers(b *testing.B) {
	benchmarkWriter(b, 1000, 100, 2)
}

func Benchmark_1000Messages_100BytesPerMessage_4Writer(b *testing.B) {
	benchmarkWriter(b, 1000, 100, 4)
}

func Benchmark_100000Messages_100BytesPerMessage_1Writer(b *testing.B) {
	benchmarkWriter(b, 100000, 100, 1)
}

func Benchmark_100000Messages_100BytesPerMessage_2Writers(b *testing.B) {
	benchmarkWriter(b, 100000, 100, 2)
}

func Benchmark_100000Messages_100BytesPerMessage_4Writer(b *testing.B) {
	benchmarkWriter(b, 100000, 100, 4)
}
