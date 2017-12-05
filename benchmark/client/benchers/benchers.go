//Package benchers provides methods to run benchmarking
package benchers

import (
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/zero-os/0-stor/benchmark/client/config"
	"github.com/zero-os/0-stor/client"
)

const (
	defaultOperations          = 1000000
	defaultAggregationInterval = time.Second
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
	Duration    time.Duration
	PerInterval []int
}

func generatedata(len int) []byte {
	data := make([]byte, len)
	rand.Read(data)
	return data
}

// stripJWT sets JWT fields to empty
// this is to prevent the zstor client creating a JWT token
// as we benchmark with a non authenticating zstordb
func stripJWT(p *client.Policy) {
	p.Organization = ""
	p.IYOAppID = ""
	p.IYOSecret = ""
}

//dataAggregator aggregates generated data to provided result
func dataAggregator(result *Result, interval time.Duration, signal <-chan struct{}, wg *sync.WaitGroup) {
	defer wg.Done()
	var totalCount int
	var alreadyCounted int

	defer func(totalCount *int) {
		result.Count = *totalCount
		fmt.Println("defer total time", totalCount)
	}(&totalCount)

	tick := time.Tick(interval)

	for {
		select {
		case <-tick:
			// aggregate data
			result.PerInterval = append(result.PerInterval, totalCount-alreadyCounted)
			alreadyCounted = totalCount
		case _, ok := <-signal:
			if !ok {
				if totalCount != alreadyCounted {
					result.PerInterval = append(result.PerInterval, totalCount-alreadyCounted)
				}
				return
			}
			totalCount++
		}
	}

}
