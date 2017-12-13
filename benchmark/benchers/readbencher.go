package benchers

import (
	"fmt"
	"time"

	"github.com/paulbellamy/ratecounter"
	"github.com/zero-os/0-stor/benchmark/config"
	"github.com/zero-os/0-stor/client"
)

//ReadBencher represents a reading benchmarker
type ReadBencher struct {
	client     *client.Client
	scenario   *config.Scenario
	scenarioID string
	keys       [][]byte
	value      []byte
}

// NewReadBencher returns a new ReadBencher
func NewReadBencher(scenarioID string, scenario *config.Scenario) (Benchmarker, error) {
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
		scenario.BenchConf.Operations = defaultOperations
	}

	// generate data
	for i := 0; i < scenario.BenchConf.Operations; i++ {
		rb.keys = append(rb.keys, generateData(scenario.BenchConf.KeySize))
	}
	rb.value = generateData(scenario.BenchConf.ValueSize)

	// initializing client
	config.SetupPolicy(&scenario.Policy)
	rb.client, err = client.New(scenario.Policy)
	if err != nil {
		return nil, fmt.Errorf("Failed creating client: %v", err)
	}

	// set testdata to client
	for _, key := range rb.keys {
		_, err := rb.client.Write(key, rb.value, nil)
		if err != nil {
			return nil, err
		}

	}

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

	// set up data aggregation interval
	interval, ok := ResultOptions[rb.scenario.BenchConf.Output]
	if !ok {
		interval = time.Second
	}

	var (
		tick         = time.Tick(interval * 1)
		start        time.Time
		counter      int64
		rc           = ratecounter.NewRateCounter(interval)
		result       = &Result{}
		maxIteration = len(rb.keys)
	)

	defer rb.cleanup()

	start = time.Now()
	for i := 0; ; i++ {
		// loop over the available keys
		key := rb.keys[i%maxIteration]

		select {
		case <-timeout:
			//timeout reached, make exit condition true
			timeout = nil
			i = maxIteration
		case <-tick:
			result.PerInterval = append(result.PerInterval, rc.Rate())
		default:
			_, _, err := rb.client.Read(key)
			if err != nil {
				return nil, err
			}
			rc.Incr(1)
			counter++
		}

		if timeout == nil && i >= maxIteration-1 {
			break
		}
	}
	result.Duration = Duration{time.Since(start)}
	result.Count = counter

	result.PerInterval = append(result.PerInterval, rc.Rate())
	return result, nil
}

func (rb *ReadBencher) cleanup() {
	rb.keys = nil
	rb.value = nil
}
