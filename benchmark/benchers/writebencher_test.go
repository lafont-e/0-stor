package benchers

import (
	"math"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/zero-os/0-stor/benchmark/config"
)

const (
	testID = "test"

	// test benchmark duration in seconds
	duration = 2
)

func TestWriteBencherRuns(t *testing.T) {
	require := require.New(t)

	// setup test servers
	meta, err := newEmbeddedMetaServer()
	require.NoError(err, "fail to start embedded meta server")
	defer meta.Stop()
	servers, serverClean := newTestZstorServers(t, 4)
	defer serverClean()

	shards := make([]string, len(servers))
	for i, server := range servers {
		shards[i] = server.Address()
	}

	clientConfig := newDefaultZstorConfig(shards, []string{meta.ListenAddr()}, 64)

	const runs = 5
	sc := config.Scenario{
		ZstorConf: clientConfig,
		BenchConf: config.BenchmarkConfig{
			Method:     "write",
			Operations: runs,
			KeySize:    5,
			ValueSize:  25,
			Output:     "per_second",
		},
	}

	// run limited benchmark
	wb, err := NewWriteBencher(testID, &sc)
	require.NoError(err)

	res, err := wb.RunBenchmark()
	require.NoError(err)
	require.Equal(int64(runs), res.Count)
}

func TestWriteBencherDuration(t *testing.T) {
	require := require.New(t)

	// setup test servers
	meta, err := newEmbeddedMetaServer()
	require.NoError(err, "fail to start embedded meta server")
	defer meta.Stop()
	servers, serverClean := newTestZstorServers(t, 4)
	defer serverClean()

	shards := make([]string, len(servers))
	for i, server := range servers {
		shards[i] = server.Address()
	}

	clientConfig := newDefaultZstorConfig(shards, []string{meta.ListenAddr()}, 64)

	sc := config.Scenario{
		ZstorConf: clientConfig,
		BenchConf: config.BenchmarkConfig{
			Method:    "write",
			Duration:  duration,
			KeySize:   5,
			ValueSize: 25,
			Output:    "per_second",
		},
	}

	// run limited benchmark
	wb, err := NewWriteBencher(testID, &sc)
	require.NoError(err)

	r, err := wb.RunBenchmark()
	require.NoError(err)

	// check if it ran for about requested duration
	runDur := r.Duration.Seconds()
	require.Equal(float64(duration), math.Floor(runDur), "rounded run duration should be equal to the requested duration")
}
