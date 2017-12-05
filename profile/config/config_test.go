package config_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/zero-os/0-stor/profile/config"
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
	fmt.Println("cl Conf", clientConf)
	require.NoError(err)

	require.NotEmpty(clientConf.Scenarios, "scenarios should not be empty")

	tmp := clientConf.Scenarios["bench1"]
	tmp.BenchConf.KeySize = 0
	clientConf.Scenarios["bench1"] = tmp

}

func TestInvalidClientConfig(t *testing.T) {
	require := require.New(t)

	yamlFile, err := os.Open(invalidBenchConfFile)
	require.NoError(err)

	_, err = config.FromReader(yamlFile)
	require.Error(err)

	yamlFile, err = os.Open(invalidKeySizeConfFile)
	require.NoError(err)

	_, err = config.FromReader(yamlFile)
	require.Error(err)

	yamlFile, err = os.Open(invalidMethodConfFile)
	require.NoError(err)

	_, err = config.FromReader(yamlFile)
	require.Error(err)

}
