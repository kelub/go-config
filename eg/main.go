package main

import (
	"context"
	"fmt"
	conf "go-config/client"
	"sync"
	"time"
)

var (
	watchValue = make(chan map[string][]byte)
)

func main() {
	wg := sync.WaitGroup{}
	consulAddr := "127.0.0.1:8500"
	confs := conf.NewConsul(consulAddr)
	wg.Add(1)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	key := "test/c"
	go getconfLoop()
	wg.Add(1)
	go confs.WatchLoop(ctx, key, time.Second*60, watchValue)
	//time.Sleep(time.Second * 10)
	//cancel()
	//fmt.Println("退出循环......")
	wg.Wait()
}

func getconfLoop() {
	for {
		select {
		case value := <-watchValue:
			fmt.Printf("xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx\n")
			for k, v := range value {
				fmt.Printf("key: [%s],value: [%s]\n", k, string(v))
				time.Sleep(50 * time.Millisecond)
			}
		}
	}
}
