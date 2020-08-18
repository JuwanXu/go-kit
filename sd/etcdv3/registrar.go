package etcdv3

import (
	"sync"
	"time"

	"github.com/go-kit/kit/log"
)

const minHeartBeatTime = 500 * time.Millisecond

// 服务注册结构体
type Registrar struct {
	client  Client        // etcd客户端
	service Service       // 服务信息
	logger  log.Logger    // 日志
	quitMtx sync.Mutex    // 退出锁
	quit    chan struct{} // 退出chan
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

// 生成一个服务注册结构体
func NewRegistrar(client Client, service Service, logger log.Logger) *Registrar {
	return &Registrar{
		client:  client,
		service: service,
		logger:  log.With(logger, "key", service.Key, "value", service.Value),
	}
}

// 注册服务
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

// 取消注册服务
func (r *Registrar) Deregister() {
	if err := r.client.Deregister(r.service); err != nil {
		r.logger.Log("err", err)
	} else {
		r.logger.Log("action", "deregister")
	}
	r.quitMtx.Lock()
	defer r.quitMtx.Unlock()
	if r.quit != nil {
		close(r.quit)
		r.quit = nil
	}
}
