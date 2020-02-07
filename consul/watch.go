package consul

import (
	"context"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/hashicorp/consul/api"
	"github.com/hashicorp/consul/api/watch"
	"sync"
	"time"
)

var gFastWatch *FastWatch

type watchHandle func(idx uint64, raw interface{})

type keyValue []byte

type keyPrefixValue map[string][]byte

type serviceValue map[string][]byte

// Watch watch 基础结构
type watchBase struct {
	// client consul client
	client api.Client
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

	//[]byte		[]byte
	//keyPrefixValue 	map[string][]byte
	// stop Stop 并发安全控制
	stop     bool
	stopCh   chan struct{}
	stopLock sync.Mutex
}

type KeyWatch struct {
	// Watch
	watch		*watchBase
	// WatchKey
	watchKey 		string
	watchValue		[]byte
	// triggerSend trigger value to send
	triggerSend chan struct{}
	// updatert watch update
	updatert time.Duration
	syncCh 	chan []byte
	// mu only for watchKey
	mu 			sync.RWMutex
	// cancel the KeyWatch
	cancel 	context.CancelFunc
}

type KeyPrefixWatch struct {
	// Watch
	watch		*watchBase
	// WatchKey
	watchKey 		string
	watchValue		map[string][]byte
	// triggerSend trigger value to send
	triggerSend chan struct{}
	// updatert watch update
	updatert time.Duration
	syncCh 	chan map[string][]byte
	// mu only for watchKey
	mu 			sync.RWMutex
	// cancel the KeyWatch
	cancel 	context.CancelFunc
}

type ServiceWatch struct {
	// Watch
	watch		*watchBase
	// WatchKey
	watchKey 		string
	watchValue		serviceValue
	// triggerSend trigger value to send
	triggerSend chan struct{}
	// updatert watch update
	updatert time.Duration
	syncCh 	chan serviceValue
	// mu only for watchKey
	mu 			sync.RWMutex
	// cancel the KeyWatch
	cancel 	context.CancelFunc
}

// NewKeyWatch
func NewKeyWatch(watchKey string) (*KeyWatch, error) {
	params := make(map[string]interface{})
	params["type"] = "key"
	params["key"] = watchKey
	plan, err := watch.Parse(params)
	if err != nil {
		return nil, err
	}
	w := &KeyWatch{
		watch: &watchBase{
			WatchType: "key",
			Plan: plan,
		},
		triggerSend: make(chan struct{},1),
		updatert:       5 * time.Second,
		watchKey: watchKey,
		watchValue: make([]byte,1),
	}
	return w, nil
}
// KeyWatch 运行监听
func (w *KeyWatch) Run(ctx context.Context, done chan<- struct{})  {
	logEntry := logrus.WithFields(logrus.Fields{
		"func_name": "KeyWatch",
		"key":       w.watchKey,
	})
	logEntry.Debugln("Start KeyWatch...")
	if w.watch.WatchType != "key" {
		logEntry.Errorf("type must be key ")
		return
	}
	ctx, cancel := context.WithCancel(ctx)

	w.cancel = cancel
	w.watch.Plan.Handler = func(idx uint64, raw interface{}) {
		var v *api.KVPair
		if raw == nil { // nil is a valid return value
			//v = nil
			w.mu.Lock()
			w.watchValue = nil
			w.mu.Unlock()
			select {
			case <-ctx.Done():
				return
			case w.triggerSend <- struct{}{}:
			default:
			}
			logEntry.Debugln("value: nil")
		} else {
			var ok bool
			if v, ok = raw.(*api.KVPair); !ok {
				return
			}
			w.mu.Lock()
			w.watchValue = v.Value
			w.mu.Unlock()
			select {
			case <-ctx.Done():
				return
			case w.triggerSend <- struct{}{}:
			default:
			}
			logEntry.Debugf("value: %s", v.Value)
		}
	}
	// 运行监听循环，退出 close done chan 通知父gorutine
	go func() {
		defer close(done)
		err := w.watch.Run()
		if err != nil {
			logEntry.Error("KeyWatch run stop", err)
		}
	}()
	go w.sender(ctx)
}

func (w *KeyWatch) SyncCh() <-chan []byte{
	return w.syncCh
}

func (w *KeyWatch) Stop(){
	if w.cancel != nil{
		w.cancel()
	}
	w.stop()
}

func (w *KeyWatch) stop(){
	w.watch.Stop()
}

func (w *KeyWatch) sender(ctx context.Context){
	ticker := time.NewTicker(w.updatert)
	defer ticker.Stop()
	defer close(w.syncCh)
	for {
		select{
		case <-ctx.Done():
			w.stop()
			return
		case <-ticker.C:
		select{
			case <-w.triggerSend:
				select{
					case w.syncCh <- w.watchValue:
				default:
					select{
					case w.triggerSend <- struct{}{}:
					default:
					}
				}
		default:
		}
		}
	}
}
// NewKeyWatch
func NewKeyPrefixWatch(watchKey string) (*KeyPrefixWatch, error) {
	params := make(map[string]interface{})
	params["type"] = "key"
	params["key"] = watchKey
	plan, err := watch.Parse(params)
	if err != nil {
		return nil, err
	}
	w := &KeyPrefixWatch{
		watch: &watchBase{
			WatchType: "keyprefix",
			Plan: plan,
		},
		triggerSend: make(chan struct{},1),
		updatert:       5 * time.Second,
		watchKey: watchKey,
		watchValue: make(map[string][]byte,1),
	}
	return w, nil
}
// KeyWatch 运行监听
func (w *KeyPrefixWatch) Run(ctx context.Context, done chan<- struct{})  {
	logEntry := logrus.WithFields(logrus.Fields{
		"func_name": "KeyPrefixWatch",
		"key":       w.watchKey,
	})
	logEntry.Debugln("Start KeyPrefixWatch...")
	if w.watch.WatchType != "keyprefix" {
		logEntry.Errorf("type must be keyprefix ")
		return
	}
	ctx, cancel := context.WithCancel(ctx)

	w.cancel = cancel
	w.watch.Plan.Handler = func(idx uint64, raw interface{}) {
		var v api.KVPairs
		if raw == nil { // nil is a valid return value
			//v = nil
			w.mu.Lock()
			w.watchValue = nil
			w.mu.Unlock()
			select {
			case <-ctx.Done():
				return
			case w.triggerSend <- struct{}{}:
			default:
			}
			logEntry.Debugln("value: nil")
		} else {
			var ok bool
			if v, ok = raw.(api.KVPairs); !ok {
				return
			}
			temp := make(map[string][]byte,len(v))
			for _, pair := range v{
				temp[pair.Key] = pair.Value
			}
			w.mu.Lock()
			w.watchValue = temp
			w.mu.Unlock()
			select {
			case <-ctx.Done():
				return
			case w.triggerSend <- struct{}{}:
			default:
			}
			logEntry.Debugf("value: %s", temp)
		}
	}
	// 运行监听循环，退出 close done chan 通知父gorutine
	go func() {
		defer close(done)
		err := w.watch.Run()
		if err != nil {
			logEntry.Error("KeyWatch run stop", err)
		}
	}()
	go w.sender(ctx)
}

func (w *KeyPrefixWatch) SyncCh() <-chan map[string][]byte{
	return w.syncCh
}

func (w *KeyPrefixWatch) Stop(){
	if w.cancel != nil{
		w.cancel()
	}
	w.stop()
}

func (w *KeyPrefixWatch) stop(){
	w.watch.Stop()
}

func (w *KeyPrefixWatch) sender(ctx context.Context){
	ticker := time.NewTicker(w.updatert)
	defer ticker.Stop()
	defer close(w.syncCh)
	for {
		select{
		case <-ctx.Done():
			w.stop()
			return
		case <-ticker.C:
			select{
			case <-w.triggerSend:
				select{
				case w.syncCh <- w.watchValue:
				default:
					select{
					case w.triggerSend <- struct{}{}:
					default:
					}
				}
			default:
			}
		}
	}
}

// run 运行监听 Plan
func (w *watchBase) Run() error {
	err := w.Plan.Run(consulAddr)
	if err != nil {
		return err
	}
	defer w.Plan.Stop()
	return nil
}

// Stop 停止 Watch
// goruntine safe
func (w *watchBase) Stop() {
	w.stopLock.Lock()
	defer w.stopLock.Unlock()
	if w.stop || w.Plan.IsStopped(){
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
	fw.WatchMap.Store(key, w)
	go w.Run(context.TODO(), done)
	value = w.SyncCh()
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
	fw.WatchMap.Store(key, w)
	go w.Run(context.TODO(), done)
	value = w.SyncCh()
	return value, nil
}

// Delete
func (fw *FastWatch) Delete(key string) {
	fw.WatchMap.Delete(key)
}
