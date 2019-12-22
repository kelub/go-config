package consul

import (
	"fmt"
	"github.com/hashicorp/consul/api"
	"github.com/hashicorp/consul/sdk/testutil"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func makeClient(t *testing.T) *testutil.TestServer {
	// Create server
	server, err := testutil.NewTestServerConfigT(t, nil)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	Address := server.HTTPAddr
	server.WaitForLeader(t)
	// Init Consul client
	err = InitConsul(Address)
	if err != nil {
		server.Stop()
		assert.NoError(t, err)
	}
	return server
}

func put(t *testing.T, c *api.Client, key string, value []byte) {
	kv := c.KV()
	pair := &api.KVPair{
		Key:   key,
		Value: value,
	}
	if _, err := kv.Put(pair, nil); err != nil {
		t.Fatalf("err: %v", err)
	}
}

func TestWatch_KeyWatch(t *testing.T) {
	s := makeClient(t)
	defer s.Stop()
	key := "service/2/timeout/get"
	//value := []byte("4s")
	//put(t, GetConsulClient(), key, value)
	w, err := NewKeyWatch(key)
	assert.NoError(t, err)
	done := make(chan struct{})
	go func() {
		defer w.Stop()
		w.KeyWatch(done)
		select {
		case <-done:
			fmt.Println("done")
		}
	}()
	time.Sleep(2 * time.Second)
	fmt.Println("waiting...")
	value := <-w.KeyValueCh
	assert.Equal(t, []byte(nil), value)

	Expect := []byte("7s")
	put(t, GetConsulClient(), key, Expect)

	value = <-w.KeyValueCh
	assert.Equal(t, Expect, value)
	w.Stop()
}
