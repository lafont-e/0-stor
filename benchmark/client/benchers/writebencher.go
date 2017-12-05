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
	writekeys           []string
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
	if scenario.BenchConf.Operations == 0 {
		wb.opsEmpty = true
		scenario.BenchConf.Operations = defaultOperations
	}

	// set up data aggregation interval
	var ok bool
	wb.aggregationInterval, ok = ResultOptions[scenario.BenchConf.Output]
	if !ok {
		wb.aggregationInterval = defaultAggregationInterval
	}

	// generate data
	for i := 0; i < scenario.BenchConf.Operations; i++ {
		wb.writekeys = append(wb.writekeys, string(generatedata(scenario.BenchConf.KeySize)))
	}
	wb.value = generatedata(scenario.BenchConf.ValueSize)

	// initializing client
	stripJWT(&scenario.Policy)
	wb.client, err = client.New(scenario.Policy)
	if err != nil {
		return nil, fmt.Errorf("Failed creating client: %v", err)
	}

	return wb, nil
}

//RunBenchmark implements Method.RunBenchmark
func (wb *WriteBencher) RunBenchmark() (*Result, error) {
	defer wb.cleanup()
	if wb.client == nil {
		return nil, fmt.Errorf("zstor client is nil")
	}

	var timeout <-chan time.Time
	if wb.scenario.BenchConf.Duration == 0 {
		timeout = nil
	} else {
		timeout = time.After(time.Duration(wb.scenario.BenchConf.Duration) * time.Second)
	}

	var wg sync.WaitGroup
	wg.Add(1)
	defer wg.Wait()

	signal := make(chan struct{})
	defer close(signal)

	go dataAggregator(&wb.result, wb.aggregationInterval, signal, &wg)

	start := time.Now()
	defer func(start time.Time, result *Result) {
		result.Duration = time.Since(start)
	}(start, &wb.result)

	for {
		for _, key := range wb.writekeys {
			select {
			case <-timeout:
				return &wb.result, nil
			default:
				wb.client.Write([]byte(key), wb.value, nil)
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
