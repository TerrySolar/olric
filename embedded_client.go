// Copyright 2018-2022 Burak Sezer
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package olric

import (
	"context"
	"encoding/json"
	"time"

	"github.com/buraksezer/olric/internal/discovery"
	"github.com/buraksezer/olric/internal/dmap"
	"github.com/buraksezer/olric/internal/protocol"
	"github.com/buraksezer/olric/internal/util"
	"github.com/buraksezer/olric/stats"
)

// EmbeddedLockContext is returned by Lock and LockWithTimeout methods.
// It should be stored in a proper way to release the lock.
type EmbeddedLockContext struct {
	key   string
	token []byte
	dm    *EmbeddedDMap
}

// Unlock releases the lock.
func (l *EmbeddedLockContext) Unlock(ctx context.Context) error {
	err := l.dm.dm.Unlock(ctx, l.key, l.token)
	return convertDMapError(err)
}

// Lease takes the duration to update the expiry for the given Lock.
func (l *EmbeddedLockContext) Lease(ctx context.Context, duration time.Duration) error {
	err := l.dm.dm.Lease(ctx, l.key, l.token, duration)
	return convertDMapError(err)
}

// EmbeddedClient is an Olric client implementation for embedded-member scenario.
type EmbeddedClient struct {
	db *Olric
}

// EmbeddedDMap is an DMap client implementation for embedded-member scenario.
type EmbeddedDMap struct {
	config        *dmapConfig
	member        discovery.Member
	dm            *dmap.DMap
	client        *EmbeddedClient
	name          string
	storageEngine string
}

// Scan returns an iterator to loop over the keys.
//
// Available scan options:
//
// * Count
// * Match
func (dm *EmbeddedDMap) Scan(ctx context.Context, options ...ScanOption) (Iterator, error) {
	cc, err := NewClusterClient([]string{dm.client.db.rt.This().String()})
	if err != nil {
		return nil, err
	}
	cdm, err := cc.NewDMap(dm.name)
	if err != nil {
		return nil, err
	}
	i, err := cdm.Scan(ctx, options...)
	if err != nil {
		return nil, err
	}

	e := &EmbeddedIterator{
		client: dm.client,
		dm:     dm.dm,
	}

	clusterIterator := i.(*ClusterIterator)
	clusterIterator.scanner = e.scanOnOwners
	e.clusterIterator = clusterIterator
	return e, nil
}

// Lock sets a lock for the given key. Acquired lock is only for the key in
// this dmap.
//
// It returns immediately if it acquires the lock for the given key. Otherwise,
// it waits until deadline.
//
// You should know that the locks are approximate, and only to be used for
// non-critical purposes.
func (dm *EmbeddedDMap) Lock(ctx context.Context, key string, deadline time.Duration) (LockContext, error) {
	token, err := dm.dm.Lock(ctx, key, 0*time.Second, deadline)
	if err != nil {
		return nil, convertDMapError(err)
	}
	return &EmbeddedLockContext{
		key:   key,
		token: token,
		dm:    dm,
	}, nil
}

// LockWithTimeout sets a lock for the given key. If the lock is still unreleased
// the end of given period of time,
// it automatically releases the lock. Acquired lock is only for the key in
// this dmap.
//
// It returns immediately if it acquires the lock for the given key. Otherwise,
// it waits until deadline.
//
// You should know that the locks are approximate, and only to be used for
// non-critical purposes.
func (dm *EmbeddedDMap) LockWithTimeout(ctx context.Context, key string, timeout, deadline time.Duration) (LockContext, error) {
	token, err := dm.dm.Lock(ctx, key, timeout, deadline)
	if err != nil {
		return nil, convertDMapError(err)
	}
	return &EmbeddedLockContext{
		key:   key,
		token: token,
		dm:    dm,
	}, nil
}

// Destroy flushes the given DMap on the cluster. You should know that there
// is no global lock on DMaps. So if you call Put/PutEx and Destroy methods
// concurrently on the cluster, Put call may set new values to the DMap.
func (dm *EmbeddedDMap) Destroy(ctx context.Context) error {
	return dm.dm.Destroy(ctx)
}

// Expire updates the expiry for the given key. It returns ErrKeyNotFound if
// the DB does not contain the key. It's thread-safe.
func (dm *EmbeddedDMap) Expire(ctx context.Context, key string, timeout time.Duration) error {
	return dm.dm.Expire(ctx, key, timeout)
}

// Name exposes name of the DMap.
func (dm *EmbeddedDMap) Name() string {
	return dm.name
}

// GetPut atomically sets the key to value and returns the old value stored at key. It returns nil if there is no
// previous value.
func (dm *EmbeddedDMap) GetPut(ctx context.Context, key string, value interface{}) (*GetResponse, error) {
	e, err := dm.dm.GetPut(ctx, key, value)
	if err != nil {
		return nil, err
	}
	return &GetResponse{
		entry: e,
	}, nil
}

// Decr atomically decrements the key by delta. The return value is the new value
// after being decremented or an error.
func (dm *EmbeddedDMap) Decr(ctx context.Context, key string, delta int) (int, error) {
	return dm.dm.Decr(ctx, key, delta)
}

// Incr atomically increments the key by delta. The return value is the new value
// after being incremented or an error.
func (dm *EmbeddedDMap) Incr(ctx context.Context, key string, delta int) (int, error) {
	return dm.dm.Incr(ctx, key, delta)
}

// IncrByFloat atomically increments the key by delta. The return value is the new value after being incremented or an error.
func (dm *EmbeddedDMap) IncrByFloat(ctx context.Context, key string, delta float64) (float64, error) {
	return dm.dm.IncrByFloat(ctx, key, delta)
}

// Delete deletes values for the given keys. Delete will not return error
// if key doesn't exist. It's thread-safe. It is safe to modify the contents
// of the argument after Delete returns.
func (dm *EmbeddedDMap) Delete(ctx context.Context, keys ...string) (int, error) {
	return dm.dm.Delete(ctx, keys...)
}

// Get gets the value for the given key. It returns ErrKeyNotFound if the DB
// does not contain the key. It's thread-safe. It is safe to modify the contents
// of the returned value. See GetResponse for the details.
func (dm *EmbeddedDMap) Get(ctx context.Context, key string) (*GetResponse, error) {
	result, err := dm.dm.Get(ctx, key)
	if err != nil {
		return nil, convertDMapError(err)
	}

	return &GetResponse{
		entry: result,
	}, nil
}

// Put sets the value for the given key. It overwrites any previous value for
// that key, and it's thread-safe. The key has to be a string. value type is arbitrary.
// It is safe to modify the contents of the arguments after Put returns but not before.
func (dm *EmbeddedDMap) Put(ctx context.Context, key string, value interface{}, options ...PutOption) error {
	var pc dmap.PutConfig
	for _, opt := range options {
		opt(&pc)
	}
	err := dm.dm.Put(ctx, key, value, &pc)
	if err != nil {
		return convertDMapError(err)
	}
	return nil
}

func (e *EmbeddedClient) NewDMap(name string, options ...DMapOption) (DMap, error) {
	dm, err := e.db.dmap.NewDMap(name)
	if err != nil {
		return nil, convertDMapError(err)
	}

	var dc dmapConfig
	for _, opt := range options {
		opt(&dc)
	}

	return &EmbeddedDMap{
		config: &dc,
		dm:     dm,
		name:   name,
		client: e,
		member: e.db.rt.This(),
	}, nil
}

// Stats exposes some useful metrics to monitor an Olric node.
func (e *EmbeddedClient) Stats(ctx context.Context, address string, options ...StatsOption) (stats.Stats, error) {
	if err := e.db.isOperable(); err != nil {
		// this node is not bootstrapped yet.
		return stats.Stats{}, err
	}
	var cfg statsConfig
	for _, opt := range options {
		opt(&cfg)
	}

	if address == e.db.rt.This().String() {
		return e.db.stats(cfg), nil
	}

	statsCmd := protocol.NewStats()
	if cfg.CollectRuntime {
		statsCmd.SetCollectRuntime()
	}
	cmd := statsCmd.Command(ctx)
	rc := e.db.client.Get(address)
	err := rc.Process(ctx, cmd)
	if err != nil {
		return stats.Stats{}, processProtocolError(err)
	}

	if err = cmd.Err(); err != nil {
		return stats.Stats{}, processProtocolError(err)
	}
	data, err := cmd.Bytes()
	if err != nil {
		return stats.Stats{}, processProtocolError(err)
	}
	var s stats.Stats
	err = json.Unmarshal(data, &s)
	if err != nil {
		return stats.Stats{}, processProtocolError(err)
	}
	return s, nil
}

// Close stops background routines and frees allocated resources.
func (e *EmbeddedClient) Close(_ context.Context) error {
	return nil
}

// Ping sends a ping message to an Olric node. Returns PONG if message is empty,
// otherwise return a copy of the message as a bulk. This command is often used to test
// if a connection is still alive, or to measure latency.
func (e *EmbeddedClient) Ping(ctx context.Context, addr, message string) (string, error) {
	response, err := e.db.ping(ctx, addr, message)
	if err != nil {
		return "", err
	}
	return util.BytesToString(response), nil
}

// RoutingTable returns the latest version of the routing table.
func (e *EmbeddedClient) RoutingTable(ctx context.Context) (RoutingTable, error) {
	return e.db.routingTable(ctx)
}

// Members returns a thread-safe list of cluster members.
func (e *EmbeddedClient) Members(_ context.Context) ([]Member, error) {
	members := e.db.rt.Discovery().GetMembers()
	coordinator := e.db.rt.Discovery().GetCoordinator()
	var result []Member
	for _, member := range members {
		m := Member{
			Name:      member.Name,
			ID:        member.ID,
			Birthdate: member.Birthdate,
		}
		if coordinator.ID == member.ID {
			m.Coordinator = true
		}
		result = append(result, m)
	}
	return result, nil
}

// NewPubSub returns a new PubSub client with the given options.
func (e *EmbeddedClient) NewPubSub(options ...PubSubOption) (*PubSub, error) {
	return newPubSub(e.db.client, options...)
}

// NewEmbeddedClient creates and returns a new EmbeddedClient instance.
func (db *Olric) NewEmbeddedClient() *EmbeddedClient {
	return &EmbeddedClient{db: db}
}

var (
	_ Client = (*EmbeddedClient)(nil)
	_ DMap   = (*EmbeddedDMap)(nil)
)
