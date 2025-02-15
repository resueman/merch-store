package closer

import (
	"bytes"
	"log"
	"os"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
)

func TestCloserCallCallbacksBySignal(t *testing.T) {
	t.Parallel()
	closer := NewCloser(os.Interrupt)

	closeFuncCalled := false
	closer.Add(
		func() error {
			closeFuncCalled = true
			return nil
		},
	)
	closer.Signal()
	closer.Wait()

	require.True(t, closeFuncCalled, "Function added to closer wasn't called after signal")
}

func TestCloserReportClosingFuncError(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer log.SetOutput(os.Stderr)

	closer := NewCloser(os.Interrupt)

	closer.Add(
		func() error {
			return errors.New("test: closing func returns error")
		},
	)
	closer.Signal()
	closer.Wait()

	require.Contains(t, buf.String(), "test: closing func returns error")
}

func TestCloserCloseOnce(t *testing.T) {
	t.Parallel()
	closer := NewCloser(os.Interrupt)

	count := 0
	closer.Add(
		func() error {
			count++
			return nil
		},
	)
	closer.CloseAll()
	closer.Wait()
	closer.CloseAll()

	require.Equal(t, 1, count)
}
