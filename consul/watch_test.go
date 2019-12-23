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

func Test_NewKeyWatch(t *testing.T) {
	s := makeClient(t)
	defer s.Stop()
	key1 := "service/1/config1/A/B/C"
	key2 := "service/1/config1/Q"
	key3 := "service/1/config1/W"
	key4 := "service/1/config1/E"
	key5 := "service/1/config2/M"
	key6 := "service/1/config3/N"

	key7 := "service/1/config1"

	keyMap := map[string]string{
		key1: "1",
		key2: "2",
		key3: "3",
		key4: "4",
		key5: "5",
		key6: "6",
	}
	for k, v := range keyMap {
		put(t, GetConsulClient(), k, []byte(v))
	}

	kv := GetConsulClient().KV()
	value1, _, err := kv.List(key1, nil)
	assert.NoError(t, err)
	value7, _, err := kv.List(key7, nil)
	assert.NoError(t, err)
	fmt.Println("value1", value1)
	fmt.Println("value7", value7)

	for _, v := range value1 {
		fmt.Println("value1", v)
	}
	for _, v := range value7 {
		fmt.Println("value7", v)
	}
}

func TestWatch_Stop(t *testing.T) {
	a := string([]byte("123456789")[3:])
	fmt.Println(a)
}
