//Package config defines a packeg used to set up client config.
package config

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"strconv"
	"time"

	"github.com/zero-os/0-stor/client"
	validator "gopkg.in/validator.v2"
	yaml "gopkg.in/yaml.v2"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

// ClientConf represents a client banchmark config
type ClientConf struct {
	Scenarios map[string]Scenario `yaml:"scenarios" validate:"nonzero"`
}

// Validate validates a ClientConf
func (clientConf *ClientConf) Validate() error {
	if len(clientConf.Scenarios) == 0 {
		return errors.New("Client config is empty")
	}

	return nil
}

// Scenario represents a scenario
type Scenario struct {
	Policy    client.Policy   `yaml:"zstor_config" validate:"nonzero"`
	BenchConf BenchmarkConfig `yaml:"bench_conf" validate:"nonzero"`
}

// Validate validates a scenario
func (sc *Scenario) Validate() error {
	err := validatePolicy(sc.Policy)
	if err != nil {
		return err
	}

	return sc.BenchConf.Validate()
}

// validatePolicy validates a client.Policy specifically for benchmarking
func validatePolicy(p client.Policy) error {
	if len(p.DataShards) <= 0 {
		return client.ErrZeroDataShards
	}

	if len(p.MetaShards) <= 0 {
		return client.ErrZeroMetaShards
	}

	if p.ReplicationNr > len(p.DataShards) {
		return client.ErrNotEnoughReplicationShards
	}

	if p.ReplicationNr == 1 {
		return client.ErrReplicationMinimum
	}

	distributionNr := (p.DistributionNr + p.DistributionRedundancy)
	if distributionNr > len(p.DataShards) {
		return client.ErrNotEnoughDistributionShards
	}

	if p.Encrypt && p.EncryptKey == "" {
		return client.ErrNoEncryptionKey
	}

	return nil
}

// SetupPolicy sets up the client.Policy for a benchmark.
// Removes IYO fields.
// Sets default namespaceing if empty
func SetupPolicy(p *client.Policy) {
	// empty IYO fields
	p.IYOAppID = ""
	p.IYOSecret = ""

	// set namespaceing if not provided
	if p.Organization == "" {
		p.Organization = "zstor"
	}
	if p.Namespace == "" {
		p.Namespace = "b-" + randomSuffix(4)
	}
}

func randomSuffix(n int) string {
	var s string
	for i := 0; i < n; i++ {
		s = s + strconv.Itoa(rand.Intn(10))
	}

	return s
}

// BenchmarkConfig represents benchmark configuration
type BenchmarkConfig struct {
	Method     string `yaml:"method" validate:"nonzero"`
	Output     string `yaml:"result_output"`
	Duration   int    `yaml:"duration"`
	Operations int    `yaml:"operations"`
	Clients    int    `yaml:"clients"`
	KeySize    int    `yaml:"key_size" validate:"nonzero"`
	ValueSize  int    `yaml:"ValueSize" validate:"nonzero"`
}

// Validate validates a BenchmarkConfig
func (bc *BenchmarkConfig) Validate() error {
	if bc.Duration == 0 && bc.Operations == 0 {
		return fmt.Errorf("One of duration or operations should be given")
	}

	return validator.Validate(bc)
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
