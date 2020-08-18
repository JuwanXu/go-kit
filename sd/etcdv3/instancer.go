package etcdv3

import (
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/sd"
	"github.com/go-kit/kit/sd/internal/instance"
)

// 客户端持有的实例维护器
type Instancer struct {
	cache  *instance.Cache // 缓存
	client Client          // 封装的etcd客户端
	prefix string          // 监听目录前缀
	logger log.Logger      // 日志
	quitC  chan struct{}   // 退出chan
}

func NewInstancer(c Client, prefix string, logger log.Logger) (*Instancer, error) {
	s := &Instancer{
		client: c,
		prefix: prefix,
		cache:  instance.NewCache(),
		logger: logger,
		quitC:  make(chan struct{}),
	}
	// 先批量获取所有的实例
	instances, err := s.client.GetEntries(s.prefix)
	if err == nil {
		logger.Log("prefix", s.prefix, "instances", len(instances))
	} else {
		logger.Log("prefix", s.prefix, "err", err)
	}
	// 更新缓存
	s.cache.Update(sd.Event{Instances: instances, Err: err})
	// 开始loop监听前缀变更
	go s.loop()
	return s, nil
}

func (s *Instancer) Stop() {
	close(s.quitC)
}

func (s *Instancer) Register(ch chan<- sd.Event) {
	s.cache.Register(ch)
}

func (s *Instancer) Deregister(ch chan<- sd.Event) {
	s.cache.Deregister(ch)
}

func (s *Instancer) loop() {
	ch := make(chan struct{})
	// 监听前缀变化
	go s.client.WatchPrefix(s.prefix, ch)
	for {
		select {
		case <-ch:
			// 有变更，则更新缓存
			instances, err := s.client.GetEntries(s.prefix)
			if err != nil {
				s.logger.Log("msg", "failed to retrieve entries", "err", err)
				s.cache.Update(sd.Event{Err: err})
				continue
			}
			s.cache.Update(sd.Event{Instances: instances})
		case <-s.quitC:
			return
		}
	}
}
