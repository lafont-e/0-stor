//Package config defines a packeg used to set up client config.
package config

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"

	"github.com/zero-os/0-stor/client"
	validator "gopkg.in/validator.v2"
	yaml "gopkg.in/yaml.v2"
)

// Scenario represents a scenario
type Scenario struct {
	Policy    client.Policy   `yaml:"zstor_config" validate:"nonzero"`
	BenchConf BenchmarkConfig `yaml:"bench_conf" validate:"nonzero"`
}

// Validate validates a scenario
func (sc *Scenario) Validate() error {
	err := sc.Policy.Validate()
	if err != nil {
		return err
	}

	return sc.BenchConf.validate()
}

// BenchmarkConfig represents benchmark configuration
type BenchmarkConfig struct {
	Method     string `yaml:"method" validate:"nonzero"`
	Output     string `yaml:"result_output"`
	Duration   int    `yaml:"duration"`
	Operations int    `yaml:"operations"`
	KeySize    int    `yaml:"key_size" validate:"nonzero"`
	ValueSize  int    `yaml:"ValueSize" validate:"nonzero"`
}

func (bc *BenchmarkConfig) validate() error {
	if bc.Duration == 0 && bc.Operations == 0 {
		return fmt.Errorf("One of duration or operations should be given")
	}

	return validator.Validate(bc)
}

// ClientConf represents a client banchmark config
type ClientConf struct {
	Scenarios map[string]Scenario `yaml:"scenarios" validate:"nonzero"`
}

func (clientConf *ClientConf) Validate() error {
	if len(clientConf.Scenarios) == 0 {
		return errors.New("Client config is empty")
	}

	return nil
}

// FromReader returns client config from a given reader
func FromReader(r io.Reader) (*ClientConf, error) {
	clientConf := &ClientConf{}

	b, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}

	// unmarshal
	if err := yaml.Unmarshal(b, clientConf); err != nil {
		return nil, err
	}

	// validate
	if err := clientConf.Validate(); err != nil {
		return nil, err
	}

	return clientConf, nil
}
