package etcd

import (
	"math"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/zero-os/0-stor/client/meta"
	"github.com/zero-os/0-stor/client/meta/embedserver"
)

func TestRoundTrip(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	etcd, err := embedserver.New()
	require.NoError(err)
	defer etcd.Stop()

	c, err := NewClient([]string{etcd.ListenAddr()})
	require.NoError(err)

	// prepare the data
	md := meta.Data{
		Key:   []byte("two"),
		Epoch: 123456789,
		Chunks: []*meta.Chunk{
			&meta.Chunk{
				Size:   math.MaxInt64,
				Key:    []byte("foo"),
				Shards: nil,
			},
			&meta.Chunk{
				Size:   1234,
				Key:    []byte("bar"),
				Shards: []string{"foo"},
			},
			&meta.Chunk{
				Size:   2,
				Key:    []byte("baz"),
				Shards: []string{"bar", "foo"},
			},
		},
		Next:     []byte("one"),
		Previous: []byte("three"),
	}

	// ensure metadata is not there yet
	_, err = c.GetMetadata(md.Key)
	require.Equal(meta.ErrNotFound, err)

	// set the metadata
	err = c.SetMetadata(md)
	require.NoError(err)

	// get it back
	storedMd, err := c.GetMetadata(md.Key)
	require.NoError(err)

	// check stored value
	assert.NotNil(storedMd)
	assert.Equal(md, *storedMd)

	err = c.DeleteMetadata(md.Key)
	require.NoError(err)
	// make sure we can't get it back
	_, err = c.GetMetadata(md.Key)
	require.Equal(meta.ErrNotFound, err)
}

// test that client can return gracefully when the server is not exist
func TestServerNotExist(t *testing.T) {
	_, err := NewClient([]string{"http://localhost:1234"})

	// make sure it is network error
	_, ok := err.(net.Error)
	require.True(t, ok)
}

// test that client can return gracefully when the server is down
func TestServerDown(t *testing.T) {
	require := require.New(t)

	etcd, err := embedserver.New()
	require.Nil(err)

	c, err := NewClient([]string{etcd.ListenAddr()})
	require.Nil(err)

	md := meta.Data{Key: []byte("key")}

	// make sure we can do some operation to server
	err = c.SetMetadata(md)
	require.Nil(err)

	outMD, err := c.GetMetadata(md.Key)
	require.Nil(err)
	require.NotNil(outMD)
	require.Equal(md, *outMD)

	// stop the server
	etcd.Stop()

	// test the SET
	done := make(chan struct{}, 1)
	go func() {
		err = c.SetMetadata(md)
		_, ok := err.(net.Error)
		require.True(ok)
		done <- struct{}{}
	}()

	select {
	case <-time.After(metaOpTimeout + (5 * time.Second)):
		// the put operation should be exited before the timeout
		t.Error("the operation should already returns with error")
	case <-done:
		t.Logf("operation exited successfully")
	}

	// test the GET
	done = make(chan struct{}, 1)
	go func() {
		_, err = c.GetMetadata(md.Key)
		_, ok := err.(net.Error)
		require.True(ok)
		done <- struct{}{}
	}()

	select {
	case <-time.After(metaOpTimeout + (5 * time.Second)):
		// the Get operation should be exited before the timeout
		t.Error("the operation should already returns with error")
	case <-done:
		t.Logf("operation exited successfully")
	}

}
