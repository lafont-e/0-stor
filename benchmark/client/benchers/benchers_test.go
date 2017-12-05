package benchers

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

const (
	runs = 5
)

func TestAggregator(t *testing.T) {
	require := require.New(t)
	wb := new(WriteBencher)
	signal := make(chan struct{})
	var wg sync.WaitGroup

	wg.Add(1)

	go dataAggregator(&wb.result, wb.aggregationInterval, signal, &wg)

	for i := 0; i < runs; i++ {
		signal <- struct{}{}
	}

	close(signal)

	timedout := waitTimeout(&wg, 30*time.Second)
	if timedout {
		require.FailNow("Timed out waiting for aggregator to close")
	}

	require.Equal(runs, wb.result.Count)
}

func waitTimeout(wg *sync.WaitGroup, timeout time.Duration) bool {
	c := make(chan struct{})
	go func() {
		defer close(c)
		wg.Wait()
	}()
	select {
	case <-c:
		return false // completed normally
	case <-time.After(timeout):
		return true // timed out
	}
}
