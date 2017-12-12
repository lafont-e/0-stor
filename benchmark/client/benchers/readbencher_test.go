package benchers

import (
	"math"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/zero-os/0-stor/benchmark/client/config"
	"github.com/zero-os/0-stor/client"
)

func TestReadBencherRuns(t *testing.T) {
	require := require.New(t)

	// setup test servers
	etcd, err := embedserver.New()
	require.NoError(err, "fail to start embedded etcd server")
	defer etcd.Stop()
	servers, serverClean := testServer(t, 4)
	defer serverClean()

	shards := make([]string, len(servers))
	for i, server := range servers {
		shards[i] = server.Address()
	}

	policy := client.Policy{
		Organization: "testorg",
		Namespace:    "namespace1",
		DataShards:   shards,
		MetaShards:   []string{etcd.ListenAddr()},
		IYOAppID:     "id",
		IYOSecret:    "secret",
	}

	sc := config.Scenario{
		Policy: policy,
		BenchConf: config.BenchmarkConfig{
			Method:     "read",
			Operations: runs,
			KeySize:    5,
			ValueSize:  25,
		},
	}

	// run limited benchmark
	rb, err := NewReadBencher(testID, &sc)
	require.NoError(err)

	res, err := rb.RunBenchmark()
	require.NoError(err)
	require.Equal(runs, res.Count)
}

func TestReadBencherDuration(t *testing.T) {
	require := require.New(t)

	// setup test servers
	etcd, err := embedserver.New()
	require.NoError(err, "fail to start embedded etcd server")
	defer etcd.Stop()
	servers, serverClean := testServer(t, 4)
	defer serverClean()

	shards := make([]string, len(servers))
	for i, server := range servers {
		shards[i] = server.Address()
	}

	policy := client.Policy{
		DataShards: shards,
		MetaShards: []string{etcd.ListenAddr()},
	}
	config.SetupPolicy(&policy)

	sc := config.Scenario{
		Policy: policy,
		BenchConf: config.BenchmarkConfig{
			Method:     "read",
			Operations: 50,
			Duration:   duration,
			KeySize:    5,
			ValueSize:  25,
		},
	}

	// run limited benchmark
	rb, err := NewReadBencher(testID, &sc)
	require.NoError(err)

	// set opsEmpty so that the benchmark wont stop after the set Operations in the config
	// 100 ops is set so the test wouldn't set default amount of keys
	// which would take too long for the test
	readBencher := rb.(*ReadBencher)
	readBencher.opsEmpty = true
	r, err := readBencher.RunBenchmark()
	require.NoError(err)

	// check if it ran for about requested duration
	runDur := r.Duration.Seconds()
	require.Equal(float64(duration), math.Floor(runDur),
		"rounded run duration should be equal to the requested duration")
}
