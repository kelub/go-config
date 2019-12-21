package consul

import (
	"fmt"
	"testing"
	"time"
)

func Test_Put(t *testing.T) {
	c := NewConsul("127.0.0.1:8500")
	if c == nil {
		t.FailNow()
	}
	//key := "test"
	//value := []byte("test consul put")
	//err := c.Put(key, value)
	//if err != nil {
	//	t.FailNow()
	//}

	err := c.Put("test/a", []byte("test consul put a"))
	if err != nil {
		t.FailNow()
	}

	err = c.Put("test/b", []byte("test consul put b"))
	if err != nil {
		t.FailNow()
	}

	err = c.Put("test/c/1", []byte("test consul put abc c 1"))
	if err != nil {
		t.FailNow()
	}
	time.Sleep(500 * time.Millisecond)
	err = c.Put("test/c/2", []byte("test consul put abc c 2"))
	if err != nil {
		t.FailNow()
	}
	time.Sleep(500 * time.Millisecond)
	err = c.Put("test/c/3", []byte("test consul put abc c 3"))
	if err != nil {
		t.FailNow()
	}
	time.Sleep(500 * time.Millisecond)
	err = c.Put("test/c/4", []byte("test consul put abc c 4"))
	if err != nil {
		t.FailNow()
	}
}

func Test_Get(t *testing.T) {
	c := NewConsul("127.0.0.1:8500")
	if c == nil {
		t.FailNow()
	}
	v, index, err := c.Get("test/c/1")
	if err != nil {
		t.FailNow()
	}
	fmt.Printf("value: %v", v)
	fmt.Println("index: ", index)
	fmt.Println("err: ", err)
}

func Test_List(t *testing.T) {
	c := NewConsul("127.0.0.1:8500")
	if c == nil {
		t.FailNow()
	}
	vs, index, err := c.List("test")
	if err != nil {
		t.FailNow()
	}
	for k, v := range vs {
		fmt.Printf("key: %s \nvalue: %s\n", k, string(v))
	}
	fmt.Println("index: ", index)
	fmt.Println("err: ", err)
}

//func Test_watch(t *testing.T) {
//	c := NewConsul("127.0.0.1:8500")
//	if c == nil {
//		t.FailNow()
//	}
//	key := "test/abc/a"
//	v := c.watch(key, time.Second*60)
//	fmt.Println("value: ", string(v[key]))
//}
