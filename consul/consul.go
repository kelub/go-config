package consul

import (
	"context"
	"fmt"
	"github.com/Sirupsen/logrus"
	"github.com/hashicorp/consul/api"
	"time"
)

var consulClient *api.Client
var consulAddr string
var consulMgr *Consul

type Consul struct {
	client *api.Client
	//watchValue 		chan map[string][]byte
}

func InitConsul(addr string) error {
	var err error
	consulAddr = addr
	consulClient, err = NewConsulClient(addr)
	if err != nil {
		return err
	}
	consulMgr = NewConsul(addr)
	return nil
}

func NewConsulClient(addr string) (*api.Client, error) {
	config := api.DefaultConfig()
	config.Address = addr
	client, err := api.NewClient(config)
	if err != nil {
		fmt.Println("init consul client error", err)
		//logrus.Errorf("init consul client error", err)
		return nil, err
	}
	return client, nil
}

func GetConsulClient() *api.Client {
	return consulClient
}

func GetConsulMgr() *Consul {
	return consulMgr
}

func GetConsulAddr() string {
	return consulAddr
}

func NewConsul(addr string) *Consul {
	return &Consul{
		client: consulClient,
	}
}

func (c *Consul) Put(key string, value []byte) error {
	kv := c.client.KV()
	p := &api.KVPair{
		Key:   key,
		Value: value,
	}
	_, err := kv.Put(p, &api.WriteOptions{})
	if err != nil {
		return err
	}
	return nil
}

func (c *Consul) Get(key string) ([]byte, uint64, error) {
	kv := c.client.KV()
	pair, queryMeta, err := kv.Get(key, &api.QueryOptions{})
	if err != nil {
		return nil, 0, err
	}
	var value []byte
	var lastIndex uint64
	if pair != nil {
		value = pair.Value
	} else {
		value = nil
	}
	if queryMeta != nil {
		lastIndex = queryMeta.LastIndex
	}
	return value, lastIndex, nil
}

func (c *Consul) List(prefix string) (map[string][]byte, uint64, error) {
	kv := c.client.KV()
	pairs, queryMeta, err := kv.List(prefix, &api.QueryOptions{})
	if err != nil {
		return nil, 0, err
	}
	if pairs == nil && queryMeta == nil {
		return nil, 0, err
	}
	kvpairs := make(map[string][]byte)
	for _, pair := range pairs {
		kvpairs[pair.Key] = pair.Value
	}
	return kvpairs, queryMeta.LastIndex, nil
}

func (c *Consul) Delete(key string) error {
	kv := c.client.KV()
	_, err := kv.Delete(key, &api.WriteOptions{})
	return err
}

func (c *Consul) DeleteTree(prefix string) error {
	kv := c.client.KV()
	_, err := kv.Delete(prefix, &api.WriteOptions{})
	return err
}

// WatchLoop Loop waitTime
func (c *Consul) WatchLoop(ctx context.Context, key string, waitTime time.Duration, watchValue chan<- map[string][]byte) {
	//refreshTimer := time.NewTimer(c.refreshInterval)
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}
		Value := c.watch(key, waitTime)
		if Value == nil {
			fmt.Println(waitTime, "no change......")
			time.Sleep(waitTime)
			continue
		} else {
			watchValue <- Value
		}
	}
}

func (c *Consul) watch(key string, waitTime time.Duration) map[string][]byte {
	_, lastIndex, err := c.List(key)
	if err != nil {
		logrus.WithError(err).Errorf("consul err. key=%s ", key)
	}
	kv := c.client.KV()
	opts := &api.QueryOptions{
		WaitIndex: lastIndex,
		WaitTime:  waitTime,
	}
	pairs, queryMeta, err := kv.List(key, opts)
	if pairs == nil && queryMeta == nil {
		logrus.WithError(err).Errorf("consul err. key=%s ", key)
	}
	if lastIndex == queryMeta.LastIndex {
		return nil
	}
	kvpairs := make(map[string][]byte)
	for _, pair := range pairs {
		kvpairs[pair.Key] = pair.Value
	}
	return kvpairs
}
