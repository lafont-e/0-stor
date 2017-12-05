package config

import (
	"fmt"
	"io"
	"io/ioutil"

	"github.com/zero-os/0-stor/client"
	"github.com/zero-os/0-stor/profile/benchers"
	validator "gopkg.in/validator.v2"
	yaml "gopkg.in/yaml.v2"
)

// Scenario represents a scenario
type Scenario struct {
	Policy    ClientPolicy    `yaml:"zstor_config" validate:"nonzero"`
	BenchConf BenchmarkConfig `yaml:"bench_conf" validate:"nonzero"`
}

// ClientPolicy represents a 0-stor client policy
// TODO: use client.policy directly
type ClientPolicy struct {
	Organization string `yaml:"organization" validate:"nonzero"`
	// Namespace label
	Namespace string `yaml:"namespace" validate:"nonzero"`

	// ItsYouOnline oauth2 application ID
	IYOAppID string `yaml:"iyo_app_id" validate:"nonzero"`
	// ItsYouOnline oauth2 application secret
	IYOSecret string `yaml:"iyo_app_secret" validate:"nonzero"`

	// Addresses to the 0-stor used to store date
	DataShards []string `yaml:"data_shards" validate:"nonzero"`
	// Addresses of the etcd cluster
	MetaShards []string `yaml:"meta_shards" validate:"nonzero"`

	// If the data written to the store is bigger then BlockSize, the data is splitted into
	// blocks of size BlockSize
	// set to 0 to never split data
	BlockSize int `yaml:"block_size"`

	// Number of replication to create when writting
	ReplicationNr int `yaml:"replication_nr"`
	// if data size is smaller than ReplicationMaxSize then data
	// will be replicated ReplicationNr time
	// if data is bigger, distribution will be used if configured
	ReplicationMaxSize int `yaml:"replication_max_size"`

	// Number of data block to create during distribution
	DistributionNr int `yaml:"distribution_data"`
	// Number of parity block to create during distribution
	DistributionRedundancy int `yaml:"distribution_parity"`

	// Enable compression
	Compress bool `yaml:"compress"`
	// Enable encryption, if true EncryptKey need to be set
	Encrypt bool `yaml:"encrypt"`
	// Key used during encryption
	EncryptKey string `yaml:"encrypt_key"`
}

func (p ClientPolicy) validate() error {
	if err := validator.Validate(p); err != nil {
		return err
	}

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

func (sc *Scenario) validate() error {
	err := sc.Policy.validate()
	if err != nil {
		return err
	}

	return sc.BenchConf.validate()
}

// BenchmarkConfig represents benchmark configuration
type BenchmarkConfig struct {
	Method     string   `yaml:"method" validate:"nonzero"`
	Output     []string `yaml:"result_output"`
	Duration   int      `yaml:"duration"`
	Operations int      `yaml:"operations"`
	KeySize    int      `yaml:"key_size" validate:"nonzero"`
	ValueSize  int      `yaml:"ValueSize" validate:"nonzero"`
}

func (bc *BenchmarkConfig) validate() error {
	if bc.Duration == 0 && bc.Operations == 0 {
		return fmt.Errorf("One of duration or operations should be given")
	}
	_, ok := benchers.MethodsMap[bc.Method]
	if !ok {
		return fmt.Errorf("%s is unknown method", bc.Method)
	}

	return validator.Validate(bc)
}

// ClientConf represents a client banchmark config
type ClientConf struct {
	Scenarios map[string]Scenario `yaml:"scenarios" validate:"nonzero"`
}

func (clientConf *ClientConf) validate() error {
	for scName, scBody := range clientConf.Scenarios {
		err := scBody.validate()
		if err != nil {
			return fmt.Errorf("Failed parsing %s with error %v", scName, err)
		}
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
	fmt.Println("cl Conf", clientConf)

	tmp := ClientPolicy{}
	fmt.Printf("Policy, %T\n", tmp.Namespace)

	// validate
	if err := clientConf.validate(); err != nil {
		return nil, err
	}

	return clientConf, nil
}
