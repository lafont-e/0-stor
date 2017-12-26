package benchers

import (
	"math"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/zero-os/0-stor/benchmark/config"
	"github.com/zero-os/0-stor/client"
)

func TestReadBencherRuns(t *testing.T) {
	require := require.New(t)

	// setup test servers
	etcd, err := NewEmbeddedServer()
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

	const runs = 5
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
	require.Equal(int64(runs), res.Count)
}

func TestReadBencherDuration(t *testing.T) {
	require := require.New(t)

	// setup test servers
	etcd, err := NewEmbeddedServer()
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
			Operations: 100, // should take more then duration
			Duration:   duration,
			KeySize:    5,
			ValueSize:  25,
		},
	}

	// run limited benchmark
	rb, err := NewReadBencher(testID, &sc)
	require.NoError(err)

	r, err := rb.RunBenchmark()
	require.NoError(err)

	// check if it ran for about requested duration
	runDur := r.Duration.Seconds()
	require.Equal(float64(duration), math.Floor(runDur),
		"rounded run duration should be equal to the requested duration")
}
