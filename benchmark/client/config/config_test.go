package config_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/zero-os/0-stor/benchmark/client/config"
)

const (
	validFile              = "configTest/validConf.yaml"
	invalidBenchConfFile   = "configTest/invalidBenchConf.yaml"
	invalidKeySizeConfFile = "configTest/invalidKeySizeConf.yaml"
	invalidMethodConfFile  = "configTest/invalidMethodConf.yaml"
)

func TestClientConfig(t *testing.T) {
	require := require.New(t)

	yamlFile, err := os.Open(validFile)
	require.NoError(err)

	clientConf, err := config.FromReader(yamlFile)
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
	var clientConf config.ClientConf
	require.Error(clientConf.Validate())

	yamlFile, err := os.Open(invalidBenchConfFile)
	require.NoError(err)

	cc, err := config.FromReader(yamlFile)
	require.NoError(err)
	sc := cc.Scenarios["bench1"]
	require.Error(sc.Validate())

	yamlFile, err = os.Open(invalidKeySizeConfFile)
	require.NoError(err)

	cc, err = config.FromReader(yamlFile)
	require.NoError(err)
	sc = cc.Scenarios["bench1"]
	require.Error(sc.Validate())
}
