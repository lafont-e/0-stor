package benchers

import (
	"fmt"
	"sync"
	"time"

	"github.com/zero-os/0-stor/benchmark/client/config"
	"github.com/zero-os/0-stor/client"
)

//WriteBencher writing benchmarker
type WriteBencher struct {
	client              *client.Client
	scenario            *config.Scenario
	scenarioID          string
	writekeys           [][]byte
	value               []byte
	opsEmpty            bool
	result              Result
	aggregationInterval time.Duration
}

// NewWriteBencher returns a new WriteBencher
func NewWriteBencher(scenarioID string, scenario *config.Scenario) (Method, error) {
	wb := new(WriteBencher)

	err := scenario.Validate()
	if err != nil {
		return nil, fmt.Errorf("Scenario %s failed: %v", scenarioID, err)
	}
	wb.scenarioID = scenarioID
	wb.scenario = scenario
	if scenario.BenchConf.Operations <= 0 {
		wb.opsEmpty = true
		scenario.BenchConf.Operations = defaultOperations
	}

	// set up data aggregation interval
	var ok bool
	wb.aggregationInterval, ok = ResultOptions[scenario.BenchConf.Output]
	if !ok {
		wb.aggregationInterval = -1
	}

	// generate data
	for i := 0; i < scenario.BenchConf.Operations; i++ {
		wb.writekeys = append(wb.writekeys, generatedata(scenario.BenchConf.KeySize))
	}
	wb.value = generatedata(scenario.BenchConf.ValueSize)

	// initializing client
	config.SetupPolicy(&scenario.Policy)
	wb.client, err = client.New(scenario.Policy)
	if err != nil {
		return nil, fmt.Errorf("Failed creating client: %v", err)
	}

	return wb, nil
}

//RunBenchmark implements Method.RunBenchmark
func (wb *WriteBencher) RunBenchmark() (*Result, error) {
	if wb.client == nil {
		return nil, fmt.Errorf("zstor client is nil")
	}

	var timeout <-chan time.Time
	if wb.scenario.BenchConf.Duration <= 0 {
		timeout = nil
	} else {
		timeout = time.After(time.Duration(wb.scenario.BenchConf.Duration) * time.Second)
	}

	signal := make(chan struct{})
	var wg sync.WaitGroup
	var start time.Time

	wg.Add(1)
	go func() {
		dataAggregator(&wb.result, wb.aggregationInterval, signal)
		wg.Done()
	}()

	defer func() {
		// set elapsed time
		wb.result.Duration = time.Since(start)

		// release test data
		wb.cleanup()

		// close signal
		close(signal)

		// wait for data aggregator to return
		wg.Wait()
	}()

	start = time.Now()
	for {
		for _, key := range wb.writekeys {
			select {
			case <-timeout:
				return &wb.result, nil
			default:
				wb.client.Write(key, wb.value, nil)
				signal <- struct{}{}
			}
		}
		if !wb.opsEmpty {
			return &wb.result, nil
		}
	}
}

func (wb *WriteBencher) cleanup() {
	wb.writekeys = nil
	wb.value = nil
}
