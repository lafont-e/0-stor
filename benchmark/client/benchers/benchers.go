//Package benchers provides methods to run benchmarking
package benchers

import (
	"math/rand"
	"time"

	"github.com/zero-os/0-stor/benchmark/client/config"
)

func init() {
	// seed random generator
	rand.Seed(time.Now().UnixNano())
}

const (
	// defaultOperations is set when BenchConf.Operations was not provided
	defaultOperations = 10000
)

var (
	// Methods represent name mapping for benchmarking methods
	benchers = map[string]BencherCtor{
		"read":  nil,
		"write": NewWriteBencher,
	}
	// ResultOptions represent name mapping for benchmarking methods
	ResultOptions = map[string]time.Duration{
		"per_second": time.Second,
		"per_minute": time.Minute,
		"per_hour":   time.Hour,
	}
)

// BencherCtor represents a benchmarker constructor
type BencherCtor func(scenarioID string, conf *config.Scenario) (Benchmarker, error)

// Benchmarker represents benchmarking methods
type Benchmarker interface {
	// RunBenchmark starts the benchmarking
	RunBenchmark() (*Result, error)
}

// Result represents a benchmark result
type Result struct {
	Count       int
	Duration    Duration
	PerInterval []int
}

// Duration represents a duration of a test result
// used for custom YAML output
type Duration struct {
	T time.Duration
}

// MarshalYAML implements yaml.Marshaler.MarshalYAML
func (d Duration) MarshalYAML() (interface{}, error) {
	return d.T.Seconds(), nil
}

// GetBencherCtor returns a BencherCtor that belongs to the provided method string
// if benchmarking method was not found, nil is returned
func GetBencherCtor(benchMethod string) BencherCtor {
	benchConstructor, ok := benchers[benchMethod]
	if !ok {
		return nil
	}
	return benchConstructor
}

// generateData generates a byteslice of provided length
// filled with random data
func generateData(len int) []byte {
	data := make([]byte, len)
	rand.Read(data)
	return data
}

// dataAggregator aggregates  data received from channel.
// Keeps track of total count of signals from channel
// as count of signals per time interval provided.
func dataAggregator(result *Result, interval time.Duration, signal <-chan struct{}) {
	var totalCount int
	var intervalCount int

	defer func() {
		result.Count = totalCount
	}()

	// setup ticker
	tick := make(<-chan time.Time)
	if interval > 1 {
		tick = time.Tick(interval)
	}

	for {
		select {
		case <-tick:
			// aggregate interval data
			result.PerInterval = append(result.PerInterval, intervalCount)
			intervalCount = 0
		case _, ok := <-signal:
			// channel is closed
			if !ok {
				if intervalCount != 0 && interval >= time.Second {
					result.PerInterval = append(result.PerInterval, intervalCount)
				}
				return
			}
			// add received signal
			intervalCount++
			totalCount++
		}
	}
}
