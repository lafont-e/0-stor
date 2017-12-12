//Package benchers provides methods to run benchmarking
package benchers

import (
	"math/rand"
	"time"

	"github.com/zero-os/0-stor/benchmark/client/config"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

const (
	defaultOperations = 1000000
)

var (
	// Methods represent name mapping for benchmarking methods
	Methods = map[string]func(scenarioID string, conf *config.Scenario) (Method, error){
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

// Method represents benchmarking methods
type Method interface {
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

func generatedata(len int) []byte {
	data := make([]byte, len)
	rand.Read(data)
	return data
}

//dataAggregator aggregates generated data to provided result
func dataAggregator(result *Result, interval time.Duration, signal <-chan struct{}) {
	var totalCount int
	var alreadyCounted int

	defer func(totalCount *int) {
		result.Count = *totalCount
	}(&totalCount)

	tick := make(<-chan time.Time)

	if interval >= time.Second {
		tick = time.Tick(interval)
	}

	for {
		select {
		case <-tick:
			// aggregate data
			result.PerInterval = append(result.PerInterval, totalCount-alreadyCounted)
			alreadyCounted = totalCount
		case _, ok := <-signal:
			if !ok {
				if totalCount != alreadyCounted && interval >= time.Second {
					result.PerInterval = append(result.PerInterval, totalCount-alreadyCounted)
				}
				return
			}
			totalCount++
		}
	}
}
