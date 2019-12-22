package consul

import (
	"fmt"
	"github.com/Sirupsen/logrus"
	"github.com/hashicorp/consul/api"
	"github.com/hashicorp/consul/api/watch"
	"sync"
)

type watchHandle func(idx uint64, raw interface{})

// Watch watch 基础结构
type Watch struct {
	// client consul client
	client api.Client
	// WatchKey
	WatchKey string
	// WatchType
	WatchType string

	// Plan watch plan
	Plan *watch.Plan
	// Handle watch 监听改变 处理器
	Handler watchHandle
	// KeyValueCh Key 变化值 chan
	KeyValueCh chan []byte
	// KeyPrefixValueCh KeyPrefix 变化值 chan
	KeyPrefixValueCh chan map[string][]byte

	// stop Stop 并发安全控制
	stop     bool
	stopCh   chan struct{}
	stopLock sync.Mutex
}

// NewKeyWatch
func NewKeyWatch(watchKey string) (*Watch, error) {
	params := make(map[string]interface{})
	params["type"] = "key"
	params["key"] = watchKey
	plan, err := watch.Parse(params)
	if err != nil {
		return nil, err
	}
	return &Watch{
		WatchKey:  watchKey,
		WatchType: "key",
		Plan:      plan,
	}, nil
}

// NewKeyPrefixWatch
func NewKeyPrefixWatch(keyprefix string) (*Watch, error) {
	params := make(map[string]interface{})
	params["type"] = "keyprefix"
	params["key"] = keyprefix
	plan, err := watch.Parse(params)
	if err != nil {
		return nil, err
	}
	return &Watch{
		WatchKey:  keyprefix,
		WatchType: "keyprefix",
		Plan:      plan,
	}, nil
}

// KeyWatch 运行监听 变化值由 KeyValueCh 传递
func (w *Watch) KeyWatch(done chan<- struct{}) {
	logEntry := logrus.WithFields(logrus.Fields{
		"func_name": "KeyWatch",
		"key":       w.WatchKey,
	})
	logEntry.Debugln("Start KeyWatch...")
	defer close(done)
	if w.WatchType != "key" {
		logEntry.Errorf("type must be key ")
		return
	}
	w.KeyValueCh = make(chan []byte, 1)
	defer close(w.KeyValueCh)
	w.Plan.Handler = func(idx uint64, raw interface{}) {
		var v *api.KVPair
		if raw == nil { // nil is a valid return value
			//v = nil
			w.KeyValueCh <- nil
			logEntry.Debugln("value: nil")

		} else {
			var ok bool
			if v, ok = raw.(*api.KVPair); !ok {
				return
			}
			w.KeyValueCh <- v.Value
			logEntry.Debugf("value: %s", v.Value)
		}
	}

	err := w.run()
	if err != nil {
		logEntry.Error("KeyWatch run stop", err)
	}
}

// KeyPrefixWatch 运行监听 变化值由 KeyPrefixValueCh 传递
func (w *Watch) KeyPrefixWatch(done chan<- struct{}) <-chan map[string][]byte {
	logEntry := logrus.WithFields(logrus.Fields{
		"func_name": "KeyPrefixWatch",
		"key":       w.WatchKey,
	})
	defer close(done)
	if w.WatchType != "keyprefix" {
		logEntry.Errorf("type must be keyprefix ")
		return nil
	}
	w.KeyPrefixValueCh = make(chan map[string][]byte, 1)
	defer close(w.KeyPrefixValueCh)
	w.Plan.Handler = func(idx uint64, raw interface{}) {
		var v api.KVPairs
		if raw == nil { // nil is a valid return value
			//v = nil
			w.KeyPrefixValueCh <- nil
		} else {
			var ok bool
			if v, ok = raw.(api.KVPairs); !ok {
				return
			}
			kvpairs := make(map[string][]byte)
			for _, pair := range v {
				kvpairs[pair.Key] = pair.Value
			}
			w.KeyPrefixValueCh <- kvpairs
		}
	}
	go func() {
		err := w.run()
		if err != nil {
			logEntry.Error("KeyWatch run", err)
		}
	}()
	return w.KeyPrefixValueCh
}

// run 运行监听 Plan
func (w *Watch) run() error {
	err := w.Plan.Run(consulAddr)
	if err != nil {
		return err
	}
	defer w.Plan.Stop()
	return nil
}

// Stop 停止 Watch
// goruntine safe
func (w *Watch) Stop() {
	w.stopLock.Lock()
	defer w.stopLock.Unlock()
	if w.stop {
		return
	} else {
		w.Plan.Stop()
		w.stop = true
	}
}

type WatchMgr struct {
	WatchMap sync.Map // key: *plan
}

func (wm *WatchMgr) KeyWatch(key string) error {
	_, ok := wm.WatchMap.Load(key)
	if ok {
		return fmt.Errorf("has exist key watch key: %s", key)
	}
	w, err := NewKeyWatch(key)
	if err != nil {
		return err
	}
	done := make(chan struct{})
	defer close(done)
	w.KeyWatch(done)
	select {
	case <-done:
		wm.WatchMap.Delete(key)
		return nil
	}
}

func (wm *WatchMgr) KeyPrefixWatch(key string) error {
	_, ok := wm.WatchMap.Load(key)
	if ok {
		return fmt.Errorf("has exist key watch key: %s", key)
	}
	w, err := NewKeyPrefixWatch(key)
	if err != nil {
		return err
	}
	done := make(chan struct{})
	defer close(done)
	w.KeyPrefixWatch(done)
	select {
	case <-done:
		wm.WatchMap.Delete(key)
		return nil
	}
}
