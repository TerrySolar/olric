package dmap

import (
	"fmt"
	"sync"
	"time"

	"github.com/buraksezer/olric/config"
)

// TODO: accessLog should be moved to its own struct
// TODO: dmapConfig will be renamed dmap.Config after refactoring. See #70 for details.

// configuration keeps DMap config control parameters and access-log for keys in a dmap.
type configuration struct {
	sync.RWMutex // protects accessLog

	maxIdleDuration time.Duration
	ttlDuration     time.Duration
	maxKeys         int
	maxInuse        int
	accessLog       map[uint64]int64
	lruSamples      int
	evictionPolicy  config.EvictionPolicy
	storageEngine   string
}

func (c *configuration) load(cfg *config.Config, name string) error {
	// Try to set config configuration for this dmap.
	c.maxIdleDuration = cfg.DMaps.MaxIdleDuration
	c.ttlDuration = cfg.DMaps.TTLDuration
	c.maxKeys = cfg.DMaps.MaxKeys
	c.maxInuse = cfg.DMaps.MaxInuse
	c.lruSamples = cfg.DMaps.LRUSamples
	c.evictionPolicy = cfg.DMaps.EvictionPolicy
	if cfg.DMaps.StorageEngine != "" {
		c.storageEngine = cfg.DMaps.StorageEngine
	}

	if cfg.DMaps.Custom != nil {
		// config.DMap struct can be used for fine-grained control.
		cs, ok := cfg.DMaps.Custom[name]
		if ok {
			if c.maxIdleDuration != cs.MaxIdleDuration {
				c.maxIdleDuration = cs.MaxIdleDuration
			}
			if c.ttlDuration != cs.TTLDuration {
				c.ttlDuration = cs.TTLDuration
			}
			if c.evictionPolicy != cs.EvictionPolicy {
				c.evictionPolicy = cs.EvictionPolicy
			}
			if c.maxKeys != cs.MaxKeys {
				c.maxKeys = cs.MaxKeys
			}
			if c.maxInuse != cs.MaxInuse {
				c.maxInuse = cs.MaxInuse
			}
			if c.lruSamples != cs.LRUSamples {
				c.lruSamples = cs.LRUSamples
			}
			if c.evictionPolicy != cs.EvictionPolicy {
				c.evictionPolicy = cs.EvictionPolicy
			}
			if cs.StorageEngine != "" && c.storageEngine != cs.StorageEngine {
				c.storageEngine = cs.StorageEngine
			}
		}
	}

	if c.evictionPolicy == config.LRUEviction || c.maxIdleDuration != 0 {
		c.accessLog = make(map[uint64]int64)
	}

	// TODO: Create a new function to verify config config.
	if c.evictionPolicy == config.LRUEviction {
		if c.maxInuse <= 0 && c.maxKeys <= 0 {
			return fmt.Errorf("maxInuse or maxKeys have to be greater than zero")
		}
		// set the default value.
		if c.lruSamples == 0 {
			c.lruSamples = config.DefaultLRUSamples
		}
	}
	return nil
}