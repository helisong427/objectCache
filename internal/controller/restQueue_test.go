package controller

import (
	"objectCache/internal"
	"testing"
	"time"
)

func Test_restQueue_Total(t *testing.T) {

	rq := newRestQueue(10)
	node := &internal.Node{Hash: 1}
	node.UpdateNodeData(0)
	rq.addNode(node)

	nodes := make([]*internal.Node, 0, 10)

	nodes = rq.getExpireNodes(uint32(time.Now().Unix()), nodes)
	if len(nodes) > 0 {
		t.Error("失败1")
	}

	rq.setRestTime(1)

	time.Sleep(time.Second)
	nodes = nodes[0:0]
	nodes = rq.getExpireNodes(uint32(time.Now().Unix()), nodes)
	if len(nodes) != 1 {
		t.Error("失败2")
	}

}

func Test_restQueue_Total1(t *testing.T) {

	rq := newRestQueue(10)

	nodes := make([]*internal.Node, 0, 10)

	nodes = rq.getExpireNodes(uint32(time.Now().Unix()), nodes)
	if len(nodes) > 0 {
		t.Error("失败1")
	}

	for i := 0; i < 10000; i++ {
		node := &internal.Node{Hash: uint64(i)}
		node.UpdateNodeData(0)
		rq.addNode(node)
	}

	if rq.count != 10000 {
		t.Error("失败2")
	}

	rq.setRestTime(1)
	time.Sleep(time.Second)

	nodes = nodes[0:0]
	nodes = rq.getExpireNodes(uint32(time.Now().Unix()), nodes)
	if len(nodes) != 10000 {
		t.Error("失败3")
	}

}
