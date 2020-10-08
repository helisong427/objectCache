package internal

import (
	"testing"
	"time"
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

var sc = NewSliceChannel()

func set(t *testing.T){
	t.Parallel()
	for i := 0; i < 10000000; i++ {
		sc.SetNode(&Node{})
		//time.Sleep(time.Microsecond*100)
	}
}

func get(t *testing.T){
	t.Parallel()
	var count uint64
	for{
		_, ok := sc.GetNode()
		if ok {
			count++
		}

		if count == 40000000{
			time.Sleep(time.Millisecond)
			_, ok := sc.GetNode()
			if ok {
				t.Errorf("失败1")
			}

			break
		}
		//time.Sleep(time.Microsecond*100)
	}
	t.Logf("count:%d", count)
}


func TestSliceChannel_SyncGetAndSet(t *testing.T) {

	t.Logf("start")

	t.Run("group", func(t *testing.T) {
		t.Run("set1", set)
		t.Run("set2", set)
		t.Run("set3", set)
		t.Run("set4", set)
		t.Run("get", get)
	})

	t.Logf("end")
}
