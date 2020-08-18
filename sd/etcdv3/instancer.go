package etcdv3

import (
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/sd"
	"github.com/go-kit/kit/sd/internal/instance"
)

type Instancer struct {
	cache  *instance.Cache
	client Client
	prefix string
	logger log.Logger
	quitc  chan struct{}
}

func NewInstancer(c Client, prefix string, logger log.Logger) (*Instancer, error) {
	s := &Instancer{
		client: c,
		prefix: prefix,
		cache:  instance.NewCache(),
		logger: logger,
		quitc:  make(chan struct{}),
	}

	instances, err := s.client.GetEntries(s.prefix)
	if err == nil {
		logger.Log("prefix", s.prefix, "instances", len(instances))
	} else {
		logger.Log("prefix", s.prefix, "err", err)
	}
	s.cache.Update(sd.Event{Instances: instances, Err: err})

	go s.loop()
	return s, nil
}

func (s *Instancer) Stop() {
	close(s.quitc)
}

func (s *Instancer) Register(ch chan<- sd.Event) {
	s.cache.Register(ch)
}

func (s *Instancer) Deregister(ch chan<- sd.Event) {
	s.cache.Deregister(ch)
}

func (s *Instancer) loop() {
	ch := make(chan struct{})
	go s.client.WatchPrefix(s.prefix, ch)

	for {
		select {
		case <-ch:
			instances, err := s.client.GetEntries(s.prefix)
			if err != nil {
				s.logger.Log("msg", "failed to retrieve entries", "err", err)
				s.cache.Update(sd.Event{Err: err})
				continue
			}
			s.cache.Update(sd.Event{Instances: instances})

		case <-s.quitc:
			return
		}
	}
}
