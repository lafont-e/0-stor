package benchers

import (
	"errors"
	"fmt"
	"time"

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
	return nil, errors.New("not implemented yet")
}
