package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zero-os/0-stor/client"
)

const (
	validFile              = "testconfigs/validConf.yaml"
	invalidBenchConfFile   = "testconfigs/invalidBenchConf.yaml"
	invalidKeySizeConfFile = "testconfigs/invalidKeySizeConf.yaml"
	invalidMethodConfFile  = "testconfigs/invalidMethodConf.yaml"
)

func TestClientConfig(t *testing.T) {
	require := require.New(t)

	yamlFile, err := os.Open(validFile)
	require.NoError(err)

	clientConf, err := FromReader(yamlFile)
	require.NoError(err)

	err = yamlFile.Close()
	require.NoError(err)

	err = clientConf.Validate()
	require.NoError(err)

	for _, sc := range clientConf.Scenarios {
		err = sc.Validate()
		require.NoError(err)
	}
}

func TestInvalidClientConfig(t *testing.T) {
	require := require.New(t)

	// empty config
	var clientConf ClientConf
	require.Error(clientConf.Validate())

	yamlFile, err := os.Open(invalidBenchConfFile)
	require.NoError(err)

	cc, err := FromReader(yamlFile)
	require.NoError(err)

	err = yamlFile.Close()
	require.NoError(err)

	sc := cc.Scenarios["bench1"]
	require.Error(sc.Validate())

	// invalid keysize
	yamlFile, err = os.Open(invalidKeySizeConfFile)
	require.NoError(err)

	cc, err = FromReader(yamlFile)
	require.NoError(err)

	err = yamlFile.Close()
	require.NoError(err)

	sc = cc.Scenarios["bench1"]
	require.Error(sc.Validate())
}

func TestSetupPolicy(t *testing.T) {
	require := require.New(t)
	p := client.Policy{
		IYOAppID:  "123",
		IYOSecret: "secret",
	}

	SetupPolicy(&p)

	require.Empty(p.IYOAppID, "IYOAppID should be empty")
	require.Empty(p.IYOSecret, "IYOSecret should be empty")
	require.NotEmpty(p.Organization, "Organization should be set")
	require.NotEmpty(p.Namespace, "Namespace should be set")
}

func TestCustomPolicyValidator(t *testing.T) {
	assert := assert.New(t)

	for _, p := range validPolicies {
		err := validatePolicy(p)
		assert.NoError(err, "Policy should be valid")
	}

	for _, p := range invalidPolicies {
		err := validatePolicy(p)
		assert.Error(err, "Policy should be invalid")
	}
}

var (
	validPolicies = []client.Policy{
		client.Policy{
			DataShards:     []string{"127.0.0.1:123", "127.0.0.1:456"},
			MetaShards:     []string{"127.0.0.1:987"},
			ReplicationNr:  2,
			DistributionNr: 1,
			Encrypt:        true,
			EncryptKey:     "123",
		},
	}

	invalidPolicies = []client.Policy{
		// No data shards
		client.Policy{
			MetaShards:     []string{"127.0.0.1:987"},
			ReplicationNr:  2,
			DistributionNr: 1,
			Encrypt:        true,
			EncryptKey:     "123",
		},

		// No meta shard
		client.Policy{
			DataShards:     []string{"127.0.0.1:123", "127.0.0.1:456"},
			ReplicationNr:  2,
			DistributionNr: 1,
			Encrypt:        true,
			EncryptKey:     "123",
		},

		// too low ReplicationNr
		client.Policy{
			DataShards:     []string{"127.0.0.1:123", "127.0.0.1:456"},
			MetaShards:     []string{"127.0.0.1:987"},
			ReplicationNr:  1,
			DistributionNr: 1,
			Encrypt:        true,
			EncryptKey:     "123",
		},

		// not enough datashards for ReplicationNr
		client.Policy{
			DataShards:     []string{"127.0.0.1:123"},
			MetaShards:     []string{"127.0.0.1:987"},
			ReplicationNr:  2,
			DistributionNr: 1,
			Encrypt:        true,
			EncryptKey:     "123",
		},

		// not enough datashards for distribution
		client.Policy{
			DataShards:     []string{"127.0.0.1:123", "127.0.0.1:456"},
			MetaShards:     []string{"127.0.0.1:987"},
			ReplicationNr:  2,
			DistributionNr: 3,
			Encrypt:        true,
			EncryptKey:     "123",
		},

		// no encryption key when requiring encryption
		client.Policy{
			DataShards:     []string{"127.0.0.1:123", "127.0.0.1:456"},
			MetaShards:     []string{"127.0.0.1:987"},
			ReplicationNr:  2,
			DistributionNr: 1,
			Encrypt:        true,
		},
	}
)
