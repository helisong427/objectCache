package controller

import (
	_ "net/http/pprof"
	"objectCache/internal"
	"objectCache/internal/storage"
	"testing"
	"time"
)

type objData struct {
	id   int
	name string
}

var c *Controller

func TestController_1(t *testing.T) {
	// go func() {
	//	http.ListenAndServe("localhost:13001", nil)
	// }()

	var segments [storage.MaxSegmentSize]*storage.Storage
	for i := 0; i < storage.MaxSegmentSize; i++ {
		segments[i] = &storage.Storage{NodeMap: make(map[uint64]*internal.Node)}
	}
	var nodeCache = internal.NewNodeCache(1e6 / 4)
	c = NewController(1e6, &segments, nodeCache)

	node := c.nodeCache.GetNode()
	var hash = uint64(1)
	ok := c.segment[hash%storage.MaxSegmentSize].Set(objData{id: 1, name: "1"}, hash, node)
	if ok {
		c.AddNode(node)
	}

	time.Sleep(time.Microsecond * 10)

	if c.initialQueue.count != 1 {
		t.Error("失败1", c.initialQueue.count)
	}

	// select {
	//
	// }
}
