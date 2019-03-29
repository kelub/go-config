package configuration

import (
	"fmt"
	"testing"
)

func Test_Put(t *testing.T) {
	c := NewConsul()
	if c == nil {
		t.FailNow()
	}
	key := "test"
	value := []byte("test consul put")
	err := c.Put(key, value)
	if err != nil {
		t.FailNow()
	}

	err = c.Put("test/abc", []byte("test consul put abc"))
	if err != nil {
		t.FailNow()
	}

	err = c.Put("test/def", []byte("test consul put def"))
	if err != nil {
		t.FailNow()
	}

	err = c.Put("test/abc/a", []byte("test consul put abc a"))
	if err != nil {
		t.FailNow()
	}
}

func Test_Get(t *testing.T) {
	c := NewConsul()
	if c == nil {
		t.FailNow()
	}
	v, index, err := c.Get("test/abc/a")
	if err != nil {
		t.FailNow()
	}
	fmt.Println("value: ", string(v))
	fmt.Println("index: ", index)
	fmt.Println("err: ", err)
}

func Test_List(t *testing.T) {
	c := NewConsul()
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

func Test_Watch(t *testing.T) {
	c := NewConsul()
	if c == nil {
		t.FailNow()
	}
	waitIndex := uint64(6803508)
	v, index, err := c.Watch("test/abc/a",waitIndex)
	if err != nil {
		t.FailNow()
	}
	fmt.Println("value: ", string(v))
	fmt.Println("index: ", index)
	fmt.Println("err: ", err)
}