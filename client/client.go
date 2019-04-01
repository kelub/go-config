package client

import (
	"context"
	"fmt"
	"github.com/Sirupsen/logrus"
	"github.com/hashicorp/consul/api"
	"time"
)

type ConfigClienter interface {
	//监听 键前缀为 key 改动
	WatchLoop(ctx context.Context, key string, waitTime time.Duration, watchValue chan<- map[string][]byte)
}

type consul struct {
	client *api.Client
	//watchValue 		chan map[string][]byte
}

func NewConsul(addr string) ConfigClienter {
	config := api.DefaultConfig()
	config.Address = addr
	client, err := api.NewClient(config)
	if err != nil {
		fmt.Println("init consul client error", err)
		//logrus.Errorf("init consul client error", err)
		return nil
	}
	return &consul{
		client: client,
	}
}

func (c *consul) List(prefix string) (map[string][]byte, uint64, error) {
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

func (c *consul) WatchLoop(ctx context.Context, key string, waitTime time.Duration, watchValue chan<- map[string][]byte) {
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
			continue
		} else {
			watchValue <- Value
		}
	}
}

func (c *consul) watch(key string, waitTime time.Duration) map[string][]byte {
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
