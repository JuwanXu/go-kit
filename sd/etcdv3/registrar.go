package etcdv3

import (
	"sync"
	"time"

	"github.com/go-kit/kit/log"
)

const minHeartBeatTime = 500 * time.Millisecond

// 服务注册结构体
type Registrar struct {
	client  Client
	service Service
	logger  log.Logger

	quitmtx sync.Mutex
	quit    chan struct{}
}

type Service struct {
	Key   string
	Value string
	TTL   *TTLOption
}

type TTLOption struct {
	heartbeat time.Duration
	ttl       time.Duration
}

func NewTTLOption(heartbeat, ttl time.Duration) *TTLOption {
	if heartbeat <= minHeartBeatTime {
		heartbeat = minHeartBeatTime
	}
	if ttl <= heartbeat {
		ttl = 3 * heartbeat
	}
	return &TTLOption{
		heartbeat: heartbeat,
		ttl:       ttl,
	}
}

func NewRegistrar(client Client, service Service, logger log.Logger) *Registrar {
	return &Registrar{
		client:  client,
		service: service,
		logger:  log.With(logger, "key", service.Key, "value", service.Value),
	}
}

func (r *Registrar) Register() {
	if err := r.client.Register(r.service); err != nil {
		r.logger.Log("err", err)
		return
	}
	if r.service.TTL != nil {
		r.logger.Log("action", "register", "lease", r.client.LeaseID())
	} else {
		r.logger.Log("action", "register")
	}
}

func (r *Registrar) Deregister() {
	if err := r.client.Deregister(r.service); err != nil {
		r.logger.Log("err", err)
	} else {
		r.logger.Log("action", "deregister")
	}
	r.quitmtx.Lock()
	defer r.quitmtx.Unlock()
	if r.quit != nil {
		close(r.quit)
		r.quit = nil
	}
}
