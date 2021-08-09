package controller

import (
	"objectCache/internal"
	"testing"
	"time"
)

func Test_queue_Total(t *testing.T) {

	var q = queue{}

	// pushBack
	for i := 0; i < queueNodeSize/2; i++ {
		n := &internal.Node{Hash: uint64(i)}
		n.UpdateData(0)

		ok := q.pushBack(n)
		if !ok {
			t.Error("失败1")
		}
	}
	time.Sleep(time.Second * 5)

	for i := queueNodeSize / 2; i < queueNodeSize; i++ {
		n := &internal.Node{Hash: uint64(i)}
		n.UpdateData(0)

		ok := q.pushBack(n)
		if !ok {
			t.Error("失败2")
		}
	}
	ok := q.pushBack(&internal.Node{})
	if ok {
		t.Error("失败3")
	}

	// getExpireNodes
	n := make([]*internal.Node, 0, 100)
	expireTime := uint32(time.Now().Unix() - 5)
	n, ok = q.fronts(expireTime, n)
	if len(n) != queueNodeSize/2 {
		t.Error("失败4：", len(n))
	}
	time.Sleep(time.Second * 5)
	n = n[0:0]
	expireTime = uint32(time.Now().Unix() - 5)
	n, ok = q.fronts(expireTime, n)
	if len(n) != queueNodeSize/2 {
		t.Error("失败5：", len(n))
	}

	n = n[0:0]
	n, ok = q.fronts(expireTime, n)
	if !ok {
		t.Error("失败6")
	}

	ok = q.pushBack(&internal.Node{})
	if ok {
		t.Error("失败7")
	}

}
