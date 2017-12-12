package benchers

import (
	"fmt"
	"sync"
	"time"

	"github.com/siddontang/go/log"
	"github.com/zero-os/0-stor/benchmark/client/config"
	"github.com/zero-os/0-stor/client"
)

//ReadBencher represents a reading benchmarker
type ReadBencher struct {
	client              *client.Client
	scenario            *config.Scenario
	scenarioID          string
	keys                [][]byte
	value               []byte
	opsEmpty            bool
	result              Result
	aggregationInterval time.Duration
}

// NewReadBencher returns a new ReadBencher
func NewReadBencher(scenarioID string, scenario *config.Scenario) (Method, error) {
	rb := new(ReadBencher)

	// validate scenario config
	err := scenario.Validate()
	if err != nil {
		return nil, fmt.Errorf("Scenario %s failed: %v", scenarioID, err)
	}

	// set bencher fields
	rb.scenarioID = scenarioID
	rb.scenario = scenario
	if scenario.BenchConf.Operations <= 0 {
		rb.opsEmpty = true
		scenario.BenchConf.Operations = defaultOperations
	}
	// set up data aggregation interval
	var ok bool
	rb.aggregationInterval, ok = ResultOptions[scenario.BenchConf.Output]
	if !ok {
		rb.aggregationInterval = -1
	}

	// generate data
	for i := 0; i < scenario.BenchConf.Operations; i++ {
		rb.keys = append(rb.keys, generatedata(scenario.BenchConf.KeySize))
	}
	rb.value = generatedata(scenario.BenchConf.ValueSize)

	// initializing client
	config.SetupPolicy(&scenario.Policy)
	rb.client, err = client.New(scenario.Policy)
	if err != nil {
		return nil, fmt.Errorf("Failed creating client: %v", err)
	}
	// set testdata to client
	var wg sync.WaitGroup
	errCh := make(chan error)
	defer close(errCh)
	for _, key := range rb.keys {
		wg.Add(1)
		select {
		case <-errCh:
			return nil, err
		default:
			go func(key []byte) {
				_, err := rb.client.Write(key, rb.value, nil)
				if err != nil {
					errCh <- err
				}
				wg.Done()
			}(key)
		}

	}
	wg.Wait()

	return rb, nil
}

// RunBenchmark runs the read benchmarker
func (rb *ReadBencher) RunBenchmark() (*Result, error) {
	if rb.client == nil {
		return nil, fmt.Errorf("zstor client is nil")
	}

	var timeout <-chan time.Time
	if rb.scenario.BenchConf.Duration <= 0 {
		timeout = nil
	} else {
		timeout = time.After(time.Duration(rb.scenario.BenchConf.Duration) * time.Second)
	}

	signal := make(chan struct{})
	var wg sync.WaitGroup
	var start time.Time

	wg.Add(1)
	go func() {
		dataAggregator(&rb.result, rb.aggregationInterval, signal)
		wg.Done()
	}()

	defer func() {
		// set elapsed time
		rb.result.Duration.T = time.Since(start)

		// release test data
		rb.cleanup()

		// close signal
		close(signal)

		// wait for data aggregator to return
		wg.Wait()
	}()

	start = time.Now()
	for {
		for _, key := range rb.keys {
			select {
			case <-timeout:
				return &rb.result, nil
			default:
				_, _, err := rb.client.Read(key)
				if err != nil {
					log.Error(err)
				}
				signal <- struct{}{}
			}
		}
		if !rb.opsEmpty {
			return &rb.result, nil
		}
	}
}

func (rb *ReadBencher) cleanup() {
	rb.keys = nil
	rb.value = nil
}
