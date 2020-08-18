package sd

import (
	"io"
	"sort"
	"sync"
	"time"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
)

type endpointCache struct {
	options            endpointerOptions
	mtx                sync.RWMutex
	factory            Factory
	cache              map[string]endpointCloser
	err                error
	endpoints          []endpoint.Endpoint
	logger             log.Logger
	invalidateDeadline time.Time
	timeNow            func() time.Time
}

type endpointCloser struct {
	endpoint.Endpoint
	io.Closer
}

func newEndpointCache(factory Factory, logger log.Logger, options endpointerOptions) *endpointCache {
	return &endpointCache{
		options: options,
		factory: factory,
		cache:   map[string]endpointCloser{},
		logger:  logger,
		timeNow: time.Now,
	}
}

func (c *endpointCache) Update(event Event) {
	c.mtx.Lock()
	defer c.mtx.Unlock()

	if event.Err == nil {
		c.updateCache(event.Instances)
		c.err = nil
		return
	}

	c.logger.Log("err", event.Err)
	if !c.options.invalidateOnError {
		return
	}
	if c.err != nil {
		return
	}
	c.err = event.Err
	c.invalidateDeadline = c.timeNow().Add(c.options.invalidateTimeout)
	return
}

func (c *endpointCache) Endpoints() ([]endpoint.Endpoint, error) {
	c.mtx.RLock()
	if c.err == nil || c.timeNow().Before(c.invalidateDeadline) {
		defer c.mtx.RUnlock()
		return c.endpoints, nil
	}
	c.mtx.RUnlock()
	c.mtx.Lock()
	defer c.mtx.Unlock()
	if c.err == nil || c.timeNow().Before(c.invalidateDeadline) {
		return c.endpoints, nil
	}
	c.updateCache(nil) // close any remaining active endpoints
	return nil, c.err
}

func (c *endpointCache) updateCache(instances []string) {
	sort.Strings(instances)
	cache := make(map[string]endpointCloser, len(instances))
	for _, instance := range instances {
		if sc, ok := c.cache[instance]; ok {
			cache[instance] = sc
			delete(c.cache, instance)
			continue
		}
		service, closer, err := c.factory(instance)
		if err != nil {
			c.logger.Log("instance", instance, "err", err)
			continue
		}
		cache[instance] = endpointCloser{service, closer}
	}
	for _, sc := range c.cache {
		if sc.Closer != nil {
			sc.Closer.Close()
		}
	}
	endpoints := make([]endpoint.Endpoint, 0, len(cache))
	for _, instance := range instances {
		if _, ok := cache[instance]; !ok {
			continue
		}
		endpoints = append(endpoints, cache[instance].Endpoint)
	}
	c.endpoints = endpoints
	c.cache = cache
}
