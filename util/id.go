package util

import "sync"

var preSec int64 = 0
var mu sync.Mutex

var gIdx int64 = 0

func getIdx(now int64) int64 {
	mu.Lock()
	defer mu.Unlock()
	if preSec != now {
		preSec = now
		gIdx = 0
	}
	gIdx++
	ret := gIdx
	return ret
}
