package benchers

import (
	"io/ioutil"
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/zero-os/0-stor/benchmark/client/config"
	"github.com/zero-os/0-stor/client"
	"github.com/zero-os/0-stor/client/meta/embedserver"
	"github.com/zero-os/0-stor/server"
)

const (
	testID = "test"
)

var (
	testBenchConf = config.BenchmarkConfig{}
)

func TestWriteBencherRuns(t *testing.T) {
	require := require.New(t)

	// setup test servers
	etcd, err := embedserver.New()
	require.NoError(err, "fail to start embedded etcd server")
	defer etcd.Stop()

	servers, serverClean := testGRPCServer(t, 4)
	defer serverClean()

	shards := make([]string, len(servers))
	for i, server := range servers {
		shards[i] = server.Addr()
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

func TestWriteBencherDuration(t *testing.T) {
	require := require.New(t)

	// setup test servers
	etcd, err := embedserver.New()
	require.NoError(err, "fail to start embedded etcd server")
	defer etcd.Stop()

	servers, serverClean := testGRPCServer(t, 4)
	defer serverClean()

	shards := make([]string, len(servers))
	for i, server := range servers {
		shards[i] = server.Addr()
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
			Method:    "write",
			Duration:  1,
			KeySize:   5,
			ValueSize: 25,
		},
	}

	// run limited benchmark
	wb, err := NewWriteBencher(testID, &sc)
	require.NoError(err)

	_, err = wb.RunBenchmark()
	require.NoError(err)
}

func testGRPCServer(t testing.TB, n int) ([]server.StoreServer, func()) {
	require := require.New(t)

	servers := make([]server.StoreServer, n)
	dirs := make([]string, n)

	for i := 0; i < n; i++ {

		tmpDir, err := ioutil.TempDir("", "0stortest")
		require.NoError(err)
		dirs[i] = tmpDir

		server, err := server.New(path.Join(tmpDir, "data"), path.Join(tmpDir, "meta"), nil, 4)
		require.NoError(err)

		_, err = server.Listen("localhost:0")
		require.NoError(err, "server failed to start listening")

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
