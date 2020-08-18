package instance

import (
	"reflect"
	"sort"
	"sync"

	"github.com/go-kit/kit/sd"
)

type Cache struct {
	mtx   sync.RWMutex
	state sd.Event
	reg   registry
}

func NewCache() *Cache {
	return &Cache{
		reg: registry{},
	}
}

func (c *Cache) Update(event sd.Event) {
	c.mtx.Lock()
	defer c.mtx.Unlock()
	sort.Strings(event.Instances)
	if reflect.DeepEqual(c.state, event) {
		return
	}
	c.state = event
	c.reg.broadcast(event)
}

func (c *Cache) State() sd.Event {
	c.mtx.RLock()
	event := c.state
	c.mtx.RUnlock()
	eventCopy := copyEvent(event)
	return eventCopy
}

func (c *Cache) Stop() {}

func (c *Cache) Register(ch chan<- sd.Event) {
	c.mtx.Lock()
	defer c.mtx.Unlock()
	c.reg.register(ch)
	event := c.state
	eventCopy := copyEvent(event)
	ch <- eventCopy
}

func (c *Cache) Deregister(ch chan<- sd.Event) {
	c.mtx.Lock()
	defer c.mtx.Unlock()
	c.reg.deregister(ch)
}

type registry map[chan<- sd.Event]struct{}

func (r registry) broadcast(event sd.Event) {
	for c := range r {
		eventCopy := copyEvent(event)
		c <- eventCopy
	}
}

func (r registry) register(c chan<- sd.Event) {
	r[c] = struct{}{}
}

func (r registry) deregister(c chan<- sd.Event) {
	delete(r, c)
}

func copyEvent(e sd.Event) sd.Event {
	if e.Instances == nil {
		return e
	}
	instances := make([]string, len(e.Instances))
	copy(instances, e.Instances)
	e.Instances = instances
	return e
}
