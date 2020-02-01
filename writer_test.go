package logrotate

import (
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"testing"
)

func TestWriter(t *testing.T) {
	dir, err := ioutil.TempDir("", "")
	require.NoError(t, err)

	w := New(dir)

	_, err = w.Write([]byte("foo"))
	require.NoError(t, err)

	require.NoError(t,w.Close(), "must close")
}
