package etcdv3

import (
	"context"
	"crypto/tls"
	"errors"
	"time"

	"go.etcd.io/etcd/clientv3"
	"go.etcd.io/etcd/pkg/transport"
)

var (
	// ErrNoKey indicates a client method needs a key but receives none.
	ErrNoKey = errors.New("no key provided")

	// ErrNoValue indicates a client method needs a value but receives none.
	ErrNoValue = errors.New("no value provided")
)

// etcd client的封装
type Client interface {
	// 返回给定前缀的所有值
	GetEntries(prefix string) ([]string, error)
	// 监听给定前缀的变化
	// 一旦监听目录发生变更，会通知ch
	WatchPrefix(prefix string, ch chan struct{})
	// 注册一个service到etcd
	Register(s Service) error
	// 取消注册一个service
	Deregister(s Service) error
	// 返回租约ID
	LeaseID() int64
}

type client struct {
	cli *clientv3.Client
	ctx context.Context
	kv  clientv3.KV
	// Watcher interface instance, used to leverage Watcher.Close()
	watcher clientv3.Watcher
	// watcher context
	wCtx context.Context
	// watcher cancel func
	wcf context.CancelFunc
	// 租约id
	leaseID clientv3.LeaseID
	// 续约心跳chan
	hbCh <-chan *clientv3.LeaseKeepAliveResponse
	// Lease interface instance, used to leverage Lease.Close()
	leaser clientv3.Lease
}

type ClientOptions struct {
	Cert          string
	Key           string
	CACert        string
	DialTimeout   time.Duration
	DialKeepAlive time.Duration
	Username      string
	Password      string
}

func NewClient(ctx context.Context, machines []string, options ClientOptions) (Client, error) {
	if options.DialTimeout == 0 {
		options.DialTimeout = 3 * time.Second
	}
	if options.DialKeepAlive == 0 {
		options.DialKeepAlive = 3 * time.Second
	}
	var err error
	var tlsCfg *tls.Config
	if options.Cert != "" && options.Key != "" {
		tlsInfo := transport.TLSInfo{
			CertFile:      options.Cert,
			KeyFile:       options.Key,
			TrustedCAFile: options.CACert,
		}
		tlsCfg, err = tlsInfo.ClientConfig()
		if err != nil {
			return nil, err
		}
	}
	cli, err := clientv3.New(clientv3.Config{
		Context:           ctx,
		Endpoints:         machines,
		DialTimeout:       options.DialTimeout,
		DialKeepAliveTime: options.DialKeepAlive,
		TLS:               tlsCfg,
		Username:          options.Username,
		Password:          options.Password,
	})
	if err != nil {
		return nil, err
	}
	return &client{
		cli: cli,
		ctx: ctx,
		kv:  clientv3.NewKV(cli),
	}, nil
}

func (c *client) LeaseID() int64 { return int64(c.leaseID) }

func (c *client) GetEntries(key string) ([]string, error) {
	resp, err := c.kv.Get(c.ctx, key, clientv3.WithPrefix())
	if err != nil {
		return nil, err
	}
	entries := make([]string, len(resp.Kvs))
	for i, kv := range resp.Kvs {
		entries[i] = string(kv.Value)
	}
	return entries, nil
}

func (c *client) WatchPrefix(prefix string, ch chan struct{}) {
	c.wCtx, c.wcf = context.WithCancel(c.ctx)
	c.watcher = clientv3.NewWatcher(c.cli)
	wch := c.watcher.Watch(c.wCtx, prefix, clientv3.WithPrefix(), clientv3.WithRev(0))
	ch <- struct{}{}
	for wr := range wch {
		if wr.Canceled {
			return
		}
		ch <- struct{}{}
	}
}

// 注册一个服务
func (c *client) Register(s Service) error {
	var err error
	if s.Key == "" {
		return ErrNoKey
	}
	if s.Value == "" {
		return ErrNoValue
	}
	if c.leaser != nil {
		c.leaser.Close()
	}
	c.leaser = clientv3.NewLease(c.cli)
	if c.watcher != nil {
		c.watcher.Close()
	}
	c.watcher = clientv3.NewWatcher(c.cli)
	if c.kv == nil {
		c.kv = clientv3.NewKV(c.cli)
	}
	if s.TTL == nil {
		s.TTL = NewTTLOption(time.Second*3, time.Second*10)
	}
	grantResp, err := c.leaser.Grant(c.ctx, int64(s.TTL.ttl.Seconds()))
	if err != nil {
		return err
	}
	c.leaseID = grantResp.ID
	_, err = c.kv.Put(
		c.ctx,
		s.Key,
		s.Value,
		clientv3.WithLease(c.leaseID),
	)
	if err != nil {
		return err
	}
	// 定时续约
	c.hbCh, err = c.leaser.KeepAlive(c.ctx, c.leaseID)
	if err != nil {
		return err
	}
	go func() {
		for {
			select {
			case r := <-c.hbCh:
				// 若chan已经关闭，则return
				if r == nil {
					return
				}
			case <-c.ctx.Done():
				return
			}
		}
	}()
	return nil
}

// 取消注册一个服务
func (c *client) Deregister(s Service) error {
	defer c.close()
	if s.Key == "" {
		return ErrNoKey
	}
	if _, err := c.cli.Delete(c.ctx, s.Key, clientv3.WithIgnoreLease()); err != nil {
		return err
	}
	return nil
}

func (c *client) close() {
	if c.leaser != nil {
		c.leaser.Close()
	}
	if c.watcher != nil {
		c.watcher.Close()
	}
	if c.wcf != nil {
		c.wcf()
	}
}
