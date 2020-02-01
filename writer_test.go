package logrotate

import (
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"log"
	"os"
	"testing"
)

func TestWriter(t *testing.T) {
	dir, err := ioutil.TempDir("", "")
	require.NoError(t, err)

	defer func() {
		require.NoError(t, os.RemoveAll(dir))
	}()

	logger := log.New(os.Stderr, "", log.LstdFlags)

	w := New(logger, dir)

	_, err = w.Write([]byte("foo"))
	require.NoError(t, err)

	require.NoError(t, w.Close(), "must close")

	files, err := ioutil.ReadDir(dir)
	require.NoError(t, err)

	require.Len(t, files, 1, "must write a single file")
}
