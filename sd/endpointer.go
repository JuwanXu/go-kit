package sd

import (
	"time"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
)

type Endpointer interface {
	Endpoints() ([]endpoint.Endpoint, error)
}

func NewEndpointer(src Instancer, f Factory, logger log.Logger, options ...EndpointerOption) *DefaultEndpointer {
	opts := endpointerOptions{}
	for _, opt := range options {
		opt(&opts)
	}
	se := &DefaultEndpointer{
		cache:     newEndpointCache(f, logger, opts),
		instancer: src,
		ch:        make(chan Event),
	}
	go se.receive()
	src.Register(se.ch)
	return se
}

type EndpointerOption func(*endpointerOptions)

type endpointerOptions struct {
	invalidateOnError bool
	invalidateTimeout time.Duration
}

type DefaultEndpointer struct {
	cache     *endpointCache
	instancer Instancer
	ch        chan Event
}

func (de *DefaultEndpointer) Close() {
	de.instancer.Deregister(de.ch)
	close(de.ch)
}

func (de *DefaultEndpointer) Endpoints() ([]endpoint.Endpoint, error) {
	return de.cache.Endpoints()
}

func (de *DefaultEndpointer) receive() {
	for event := range de.ch {
		de.cache.Update(event)
	}
}
