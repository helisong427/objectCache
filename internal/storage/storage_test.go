package storage

import (
	"cache/internal"
	"testing"
	"time"
)

type data struct {
	id int
	name string
}

func TestStorage_Total(t *testing.T) {

	nc := internal.NewNodeCache(1000)
	s := Storage{NodeMap: make(map[uint64]*internal.Node)}

	//set
	for i := 0; i < 100; i++ {
		if !s.Set(data{id:i, name:"aa"}, uint64(i), 0, nc.GetNode()) {
			t.Error("失败1")
		}
	}


	//get
	for i := 0; i < 100; i++ {
		n, ok := s.Get(uint64(i))
		if !ok {
			t.Error("失败2")
		}

		n, ok = s.Get(uint64(i))
		if !ok {
			t.Error("失败3")
		}

		if n.GetCurrentCount() != 1{
			t.Error("失败4")
		}
	}
	time.Sleep(time.Second*20)
	n, ok := s.Get(uint64(10))
	if !ok {
		t.Error("失败5")
	}
	if n.GetCurrentCount() != 2{
		t.Error("失败6")
	}


	//del
	n, ok = s.Del(10)
	if !ok{
		t.Error("失败7")
	}

	if n.Hash != 10 {
		t.Error("失败8")
	}

	nc.SaveDirtyNode(n)
	n.Hash = 0

	n, ok = s.Del(10)
	if ok{
		t.Error("失败9")
	}

}