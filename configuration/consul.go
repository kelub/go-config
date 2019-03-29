package configuration

import (
	"fmt"
	"github.com/hashicorp/consul/api"
)

type consul struct {
	client *api.Client
}

func NewConsul() *consul {
	config := api.DefaultConfig()
	config.Address = "192.168.9.165:8500"
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

func (c *consul) Put(key string, value []byte) error {
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

func (c *consul) Get(key string) ([]byte, uint64, error) {
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

func (c *consul) Watch(key string, waitIndex uint64) ([]byte, uint64, error) {
	kv := c.client.KV()
	opts := &api.QueryOptions{
		WaitIndex: waitIndex,
		//WaitTime: time.Minute,
	}
	pair, queryMeta, err := kv.Get(key, opts)
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
