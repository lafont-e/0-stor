package benchers

import (
	"fmt"
	"io/ioutil"
	"math"
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/zero-os/0-stor/benchmark/client/config"
	"github.com/zero-os/0-stor/client"
	"github.com/zero-os/0-stor/server/api"
	"github.com/zero-os/0-stor/server/api/grpc"
	"github.com/zero-os/0-stor/server/db/badger"
)

const (
	testID = "test"

	// test benchmark duration in seconds
	duration = 2
)

func testWriteBencherRuns(t *testing.T) {
	require := require.New(t)

	// setup test servers
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
		MetaShards:   []string{"testserver:123"},
		IYOAppID:     "id",
		IYOSecret:    "secret",
	}

	sc := config.Scenario{
		Policy: policy,
		BenchConf: config.BenchmarkConfig{
			Method:     "write",
			Operations: runs,
			KeySize:    5,
			ValueSize:  25,
		},
	}

	// run limited benchmark
	wb, err := NewWriteBencher(testID, &sc)
	require.NoError(err)

	res, err := wb.RunBenchmark()
	require.NoError(err)
	require.Equal(runs, res.Count)
}

func testWriteBencherDuration(t *testing.T) {
	require := require.New(t)

	// setup test servers
	servers, serverClean := testServer(t, 4)
	defer serverClean()

	shards := make([]string, len(servers))
	for i, server := range servers {
		shards[i] = server.Address()
		fmt.Println(server.Address())
	}

	policy := client.Policy{
		DataShards: shards,
		MetaShards: []string{"testserver"},
	}
	config.SetupPolicy(&policy)

	sc := config.Scenario{
		Policy: policy,
		BenchConf: config.BenchmarkConfig{
			Method:    "write",
			Duration:  duration,
			KeySize:   5,
			ValueSize: 25,
		},
	}

	// run limited benchmark
	wb, err := NewWriteBencher(testID, &sc)
	require.NoError(err)

	r, err := wb.RunBenchmark()
	require.NoError(err)

	// check if it ran for about requested duration
	runDur := r.Duration.Seconds()
	require.Equal(float64(duration), math.Floor(runDur),
		"rounded run duration should be equal to the requested duration")
}

// returns n amount of zstordb servers
func testServer(t testing.TB, n int) ([]api.Server, func()) {
	require := require.New(t)

	servers := make([]api.Server, n)
	dirs := make([]string, n)

	for i := 0; i < n; i++ {

		tmpDir, err := ioutil.TempDir("", "0stortest")
		require.NoError(err)
		dirs[i] = tmpDir

		db, err := badger.New(path.Join(tmpDir, "data"), path.Join(tmpDir, "meta"))
		require.NoError(err)

		server, err := grpc.New(db, nil, 4, 0)
		require.NoError(err)

		go func() {
			err := server.Listen("localhost:0")
			require.NoError(err, "server failed to start listening")
		}()

		servers[i] = server
	}

	clean := func() {
		for _, server := range servers {
			server.Close()
		}
		for _, dir := range dirs {
			os.RemoveAll(dir)
		}
	}

	return servers, clean
}
