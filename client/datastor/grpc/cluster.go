package grpc

import (
	"errors"
	"sync"

	log "github.com/Sirupsen/logrus"
	"github.com/zero-os/0-stor/client/datastor"
)

// NewCluster creates a new cluster,
// and pre-loading it with a client for each of the listed (and thus known) shards.
// Unlisted shards's clients are also stored, bu those are loaded on the fly, only when needed.
func NewCluster(addresses []string, label string, jwtTokenGetter datastor.JWTTokenGetter) (*Cluster, error) {
	if len(addresses) == 0 {
		return nil, errors.New("no listed addresses given")
	}
	if label == "" {
		return nil, errors.New("no label given")
	}

	var (
		listedCount = len(addresses)
		listedSlice = make([]*Shard, 0, listedCount)
		listed      = make(map[string]*Shard, listedCount)
	)
	// create all shards, one by one
	for _, address := range addresses {
		client, err := NewClient(address, label, jwtTokenGetter)
		if err != nil {
			// close all shards already opened
			var closeErr error
			for address, shard := range listed {
				closeErr = shard.Close()
				if closeErr != nil {
					log.Errorf(
						"error while closing (because %v) listed shard (%s): %v",
						err, address, closeErr)
				}
			}
			return nil, err
		}
		shard := &Shard{
			Client:  client,
			address: address,
		}
		listedSlice = append(listedSlice, shard)
		listed[address] = shard
	}

	// return valid cluster, ready for usage
	return &Cluster{
		listed:         listed,
		listedSlice:    listedSlice,
		listedCount:    int64(listedCount),
		unlisted:       make(map[string]*Shard),
		label:          label,
		jwtTokenGetter: jwtTokenGetter,
	}, nil
}

// Cluster implements datastor.Cluster for
// clients which interface with zstordb using the GRPC interface.
type Cluster struct {
	listed      map[string]*Shard
	listedSlice []*Shard
	listedCount int64

	unlisted    map[string]*Shard
	unlistedMux sync.Mutex

	label          string
	jwtTokenGetter datastor.JWTTokenGetter
}

// GetShard implements datastor.Cluster.GetShard
func (cluster *Cluster) GetShard(address string) (datastor.Shard, error) {
	shard, ok := cluster.listed[address]
	if ok {
		// return the known listed client
		return shard, nil
	}

	cluster.unlistedMux.Lock()
	defer cluster.unlistedMux.Unlock()
	shard, ok = cluster.unlisted[address]
	if ok {
		// return the known unlisted client
		return shard, nil
	}

	// create and return an unknown unlisted client,
	// making it known for next time it is needed
	client, err := NewClient(address, cluster.label, cluster.jwtTokenGetter)
	if err != nil {
		return nil, err
	}
	shard = &Shard{
		Client:  client,
		address: address,
	}
	cluster.unlisted[address] = shard
	return shard, nil
}

// GetRandomShard implements datastor.Cluster.GetRandomShard
func (cluster *Cluster) GetRandomShard() (datastor.Shard, error) {
	// get a crypto/pseudo random index
	index := datastor.RandShardIndex(cluster.listedCount)

	// return the client with the random (valid) index
	return cluster.listedSlice[index], nil
}

// GetRandomShardIterator implements datastor.Cluster.GetRandomShardIterator
func (cluster *Cluster) GetRandomShardIterator(exceptShards []string) datastor.ShardIterator {
	slice := cluster.filteredSlice(exceptShards)
	return datastor.NewRandomShardIterator(slice)
}

func (cluster *Cluster) filteredSlice(exceptShards []string) []datastor.Shard {
	if len(exceptShards) == 0 {
		slice := make([]datastor.Shard, cluster.listedCount)
		for i := range slice {
			slice[i] = cluster.listedSlice[i]
		}
		return slice
	}

	fm := make(map[string]struct{}, len(exceptShards))
	for _, id := range exceptShards {
		fm[id] = struct{}{}
	}

	var (
		ok       bool
		filtered = make([]datastor.Shard, 0, cluster.listedCount)
	)
	for _, shard := range cluster.listedSlice {
		if _, ok = fm[shard.Identifier()]; !ok {
			filtered = append(filtered, shard)
		}
	}
	return filtered
}

// Close implements datastor.Cluster.Close
func (cluster *Cluster) Close() error {
	cluster.unlistedMux.Lock()
	defer cluster.unlistedMux.Unlock()

	var (
		err      error
		errCount int
	)

	// close all unlisted shards first
	for address, shard := range cluster.unlisted {
		err = shard.Close()
		if err != nil {
			errCount++
			log.Errorf(
				"error while closing unlisted shard (%s): %v", address, err)
		}
	}

	// close all listed shards next
	for address, shard := range cluster.listed {
		err = shard.Close()
		if err != nil {
			errCount++
			log.Errorf(
				"error while closing listed shard (%s): %v", address, err)
		}
	}

	// if at least one shard returned an error while closing,
	// return it as a generic error for now
	if errCount > 0 {
		return errors.New("one or multiple shards returned an error while closing")
	}
	return nil
}

// ListedShardCount implements datastor.Cluster.ListedShardCount
func (cluster *Cluster) ListedShardCount() int {
	return int(cluster.listedCount)
}

// Shard implements datastor.Shard for
// GRPC clients, to make those clients work within a cluster of other GRPC clients.
type Shard struct {
	*Client
	address string
}

// Identifier implements datastor.Shard.Identifier
func (shard *Shard) Identifier() string {
	return shard.address
}

var (
	_ datastor.Cluster = (*Cluster)(nil)
	_ datastor.Shard   = (*Shard)(nil)
)