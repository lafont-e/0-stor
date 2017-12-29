package benchers

import (
	"fmt"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/paulbellamy/ratecounter"
	"github.com/zero-os/0-stor/benchmark/config"
	"github.com/zero-os/0-stor/client"
)

//WriteBencher represents a writing benchmarker
type WriteBencher struct {
	client     *client.Client
	scenario   *config.Scenario
	scenarioID string
	keys       [][]byte
	value      []byte
}

// NewWriteBencher returns a new WriteBencher
func NewWriteBencher(scenarioID string, scenario *config.Scenario) (Benchmarker, error) {
	wb := new(WriteBencher)

	err := scenario.Validate()
	if err != nil {
		log.Errorf("Error validating scenario: %v", err)
		return nil, fmt.Errorf("Scenario %s failed: %v", scenarioID, err)
	}
	wb.scenarioID = scenarioID
	wb.scenario = scenario
	if scenario.BenchConf.Operations <= 0 {
		// wb.opsEmpty = true
		scenario.BenchConf.Operations = defaultOperations
	}

	// generate data
	for i := 0; i < scenario.BenchConf.Operations; i++ {
		wb.keys = append(wb.keys, generateData(scenario.BenchConf.KeySize))
	}
	wb.value = generateData(scenario.BenchConf.ValueSize)

	// initializing client
	config.SetupPolicy(&scenario.Policy)
	wb.client, err = client.New(scenario.Policy)
	if err != nil {
		log.Errorf("Error creating client: %v", err)
		return nil, fmt.Errorf("Failed creating client: %v", err)
	}

	return wb, nil
}

//RunBenchmark implements Method.RunBenchmark
func (wb *WriteBencher) RunBenchmark() (*Result, error) {
	if wb.client == nil {
		log.Error("zstor client is nil when trying to run a write bencher")
		return nil, fmt.Errorf("zstor client is nil")
	}

	var timeout <-chan time.Time
	if wb.scenario.BenchConf.Duration <= 0 {
		timeout = nil
	} else {
		timeout = time.After(time.Duration(wb.scenario.BenchConf.Duration) * time.Second)
	}

	// set up data aggregation interval
	interval, ok := ResultOptions[wb.scenario.BenchConf.Output]
	if !ok {
		interval = time.Second
	}

	var (
		tick         = time.Tick(interval * 1)
		start        time.Time
		counter      int64
		rc           = ratecounter.NewRateCounter(interval)
		result       = &Result{}
		maxIteration = len(wb.keys)
	)

	start = time.Now()
	for i := 0; i < maxIteration; i++ {
		// loop over the available keys
		key := wb.keys[i%maxIteration]

		select {
		case <-timeout:
			i = maxIteration
		case <-tick:
			result.PerInterval = append(result.PerInterval, rc.Rate())

		default:
			_, err := wb.client.Write(key, wb.value, nil)
			if err != nil {
				log.Errorf("Error write request to client: %v", err)
				return nil, err
			}
			rc.Incr(1)
			counter++
		}
	}
	result.Duration = Duration{time.Since(start)}
	result.Count = counter
	result.PerInterval = append(result.PerInterval, rc.Rate())

	wb.cleanup()

	return result, nil
}

func (wb *WriteBencher) cleanup() {
	wb.keys = nil
	wb.value = nil
}
