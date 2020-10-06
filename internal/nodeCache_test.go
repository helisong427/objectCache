package internal

import (
	"testing"
	"time"
)

type data struct {
	id int
	name string
}


func TestNodeCache_Total(t *testing.T) {

	nc := NewNodeCache(1000)
	for i := 0; i < 1001; i++{
		nc.SaveNode(&Node{
			Hash: uint64(i),
			Obj: data{id:i, name:"aa"},
		})
	}

	for i := 0; i < 1000; i++{
		n := nc.GetNode()
		if n.Hash != uint64(i) {
			t.Error("失败1")
		}
	}

	n := nc.GetNode()
	if n.Hash != 0 {
		t.Error("失败2")
	}

	n0 := &Node{
		Hash: 100,
		Obj: data{id:100, name:"aa"},
	}

	nc.SaveDirtyNode(n0)

	n1 := nc.GetNode()
	if n1.Obj != nil {
		t.Error("失败3")
	}

	n0.Hash = 0

	time.Sleep(time.Second*11)
	n2 := nc.GetNode()
	if n2.Obj == nil && n2.Obj.(data).id != 100 {
		t.Error("失败4")
	}


}