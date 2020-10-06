package internal

import (
	"testing"
)


func TestSliceChannel_Total(t *testing.T) {

	sc := NewSliceChannel()

	for i := 0; i <100000; i++ {
		sc.SetNode(&Node{
			Hash: uint64(i),
		})
	}

	for i := 0; i < 100000; i++ {
		node, ok := sc.GetNode()
		if !ok {
			t.Error("失败1")
		}
		if node.Hash != uint64(i){
			t.Error("失败2")
		}
	}
	_, ok := sc.GetNode()
	if ok {
		t.Error("失败3")
	}

	for i := 0; i <100000; i++ {
		sc.SetNode(&Node{
			Hash: uint64(i),
		})
	}


}
