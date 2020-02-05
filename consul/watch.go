package consul

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/hashicorp/consul/api"
	"github.com/hashicorp/consul/api/watch"
	"sync"
)

var gFastWatch *FastWatch

type watchHandle func(idx uint64, raw interface{})

type keyValue []byte

type keyPrefixValue map[string][]byte

type serviceValue map[string][]byte

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
	//// KeyValueCh Key 变化值 chan
	//KeyValueCh chan []byte
	//// KeyPrefixValueCh KeyPrefix 变化值 chan
	//KeyPrefixValueCh chan map[string][]byte

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

// NewServiceWatch
func NewServiceWatch(service, tag string, passingonly bool) (*Watch, error) {
	params := make(map[string]interface{})
	params["type"] = "service"
	params["service"] = "service"
	params["tag"] = "tag"
	params["passingonly"] = passingonly
	plan, err := watch.Parse(params)
	if err != nil {
		return nil, err
	}
	return &Watch{
		WatchKey: tag,
		WatchType: "service",
		Plan:      plan,
	}, nil
}
// ServiceWatch key is tag
func (w *Watch) ServiceWatch(done chan<- struct{}) <-chan serviceValue {
	logEntry := logrus.WithFields(logrus.Fields{
		"func_name": "ServiceWatch",
		"key":       w.WatchKey,
	})
	logEntry.Debugln("Start ServiceWatch...")
	if w.WatchType != "service" {
		logEntry.Errorf("type must be service ")
		return nil
	}
	valueCh := make(chan serviceValue, 1)
	defer close(valueCh)
	w.Plan.Handler = func(idx uint64, raw interface{}){
		var v []*api.ServiceEntry
		if raw == nil{
			valueCh <- nil
			logEntry.Debugln("value: nil")

		}else{
			var ok bool
			if v, ok = raw.([]*api.ServiceEntry); !ok{
				return
			}
			//w.Plan.IsStopped()
			addrs := make(serviceValue, len(v))
			for i,_ := range v{
				addrs[w.WatchKey] = []byte(v[i].Service.Address)
			}
			valueCh <- addrs
		}
	}
	// 运行监听循环，退出 close done chan 通知父gorutine
	go func() {
		defer close(done)
		err := w.run()
		if err != nil {
			logEntry.Error("ServiceWatch run stop", err)
		}
	}()
	return valueCh
}

// KeyWatch 运行监听 变化值由 KeyValueCh 传递
func (w *Watch) KeyWatch(done chan<- struct{}) <-chan []byte {
	logEntry := logrus.WithFields(logrus.Fields{
		"func_name": "KeyWatch",
		"key":       w.WatchKey,
	})
	logEntry.Debugln("Start KeyWatch...")
	if w.WatchType != "key" {
		logEntry.Errorf("type must be key ")
		return nil
	}
	KeyValueCh := make(chan []byte, 1)
	defer close(KeyValueCh)
	w.Plan.Handler = func(idx uint64, raw interface{}) {
		var v *api.KVPair
		if raw == nil { // nil is a valid return value
			//v = nil
			KeyValueCh <- nil
			logEntry.Debugln("value: nil")

		} else {
			var ok bool
			if v, ok = raw.(*api.KVPair); !ok {
				return
			}
			KeyValueCh <- v.Value
			logEntry.Debugf("value: %s", v.Value)
		}
	}
	// 运行监听循环，退出 close done chan 通知父gorutine
	go func() {
		defer close(done)
		err := w.run()
		if err != nil {
			logEntry.Error("KeyWatch run stop", err)
		}
	}()
	return KeyValueCh
}

// KeyPrefixWatch 运行监听 变化值由 KeyPrefixValueCh 传递
func (w *Watch) KeyPrefixWatch(done chan<- struct{}) <-chan map[string][]byte {
	logEntry := logrus.WithFields(logrus.Fields{
		"func_name": "KeyPrefixWatch",
		"key":       w.WatchKey,
	})
	if w.WatchType != "keyprefix" {
		logEntry.Errorf("type must be keyprefix ")
		return nil
	}
	KeyPrefixValueCh := make(chan map[string][]byte, 1)
	defer close(KeyPrefixValueCh)
	w.Plan.Handler = func(idx uint64, raw interface{}) {
		var v api.KVPairs
		if raw == nil { // nil is a valid return value
			//v = nil
			KeyPrefixValueCh <- nil
		} else {
			var ok bool
			if v, ok = raw.(api.KVPairs); !ok {
				return
			}
			kvpairs := make(map[string][]byte)
			for _, pair := range v {
				kvpairs[pair.Key] = pair.Value
			}
			KeyPrefixValueCh <- kvpairs
		}
	}
	go func() {
		defer close(done)
		err := w.run()
		if err != nil {
			logEntry.Error("KeyWatch run", err)
		}
	}()
	return KeyPrefixValueCh
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

// 全局 FastWatch 控制
type FastWatch struct {
	WatchMap sync.Map // key: *plan
}

func CreateFastWatch() *FastWatch {
	gFastWatch = &FastWatch{}
	return gFastWatch
}

func GetFastWatch() *FastWatch {
	return gFastWatch
}

// KeyWatch 得到新的 KeyWatch 循环，并返回值的 chan
// done 通知父 goruntine Watch 循环退出。
// goruntine safe
func (fw *FastWatch) KeyWatch(done chan<- struct{}, key string) (value <-chan []byte, err error) {
	_, ok := fw.WatchMap.Load(key)
	if ok {
		return nil, fmt.Errorf("has exist key watch key: %s", key)
	}
	w, err := NewKeyWatch(key)
	if err != nil {
		return nil, err
	}
	value = w.KeyWatch(done)
	fw.WatchMap.Store(key, w)
	return value, nil
}

// KeyPrefixWatch 得到新的 KeyPrefixWatch 循环，并返回值的 chan
// done 通知父 goruntine Watch 循环退出。
// goruntine safe
func (fw *FastWatch) KeyPrefixWatch(done chan<- struct{}, key string) (value <-chan map[string][]byte, err error) {
	_, ok := fw.WatchMap.Load(key)
	if ok {
		return nil, fmt.Errorf("has exist key watch key: %s", key)
	}
	w, err := NewKeyPrefixWatch(key)
	if err != nil {
		return nil, err
	}
	value = w.KeyPrefixWatch(done)
	fw.WatchMap.Store(key, w)
	return value, nil
}

// Delete
func (fw *FastWatch) Delete(key string) {
	fw.WatchMap.Delete(key)
}
