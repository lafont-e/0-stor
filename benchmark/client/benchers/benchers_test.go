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

	// test without collecting interval data
	wg.Add(1)
	go func() {
		dataAggregator(&wb.result, -1, signal)
		wg.Done()
	}()

	for i := 0; i < runs; i++ {
		signal <- struct{}{}
	}

	close(signal)

	timedout := waitTimeout(&wg, 30*time.Second)
	if timedout {
		require.FailNow("Timed out waiting for aggregator to close")
	}

	require.Equal(runs, wb.result.Count)
	require.Empty(wb.result.PerInterval, "Interval list should be empty")

	// test with collecting interval data
	signal = make(chan struct{})
	wg.Add(1)
	go func() {
		dataAggregator(&wb.result, 1*time.Second, signal)
		wg.Done()
	}()

	// wait for some data to be aggregated
	time.Sleep(2 * time.Second)
	close(signal)

	timedout = waitTimeout(&wg, 30*time.Second)
	if timedout {
		require.FailNow("Timed out waiting for aggregator to close")
	}
	require.NotEmpty(wb.result.PerInterval, "Interval list should not be empty")
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
