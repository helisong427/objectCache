package storage

import (
	"math"
	"objectCache/internal"
	"sync"
	"time"
)

const (
	MaxSegmentSize = 256
)

// Storage存储对象的并发单元
// 持有一个读写锁和一个map，internal.Node直接存储与map中，读写锁就锁定这个map。
type Storage struct {
	sync.RWMutex
	NodeMap map[uint64]*internal.Node
}

func (s *Storage) Set(obj interface{}, hash uint64, expire int, node *internal.Node) (ok bool) {
	s.Lock()
	defer s.Unlock()
	var now = time.Now()
	if _, ok = s.NodeMap[hash]; !ok {
		node.Init(obj, hash, uint32(now.Unix()))
		s.NodeMap[hash] = node
	} else {
		s.NodeMap[hash].Obj = node.Obj
		_ = s.NodeMap[hash].IncrementReadCount()
	}

	if expire > 0 {
		s.NodeMap[hash].Expire = uint32(now.Add(time.Second * time.Duration(expire)).Unix())
	} else {
		// 不过期
		s.NodeMap[hash].Expire = math.MaxUint32 // 2106-02-07 14:28:15 +0800 CST
	}
	return !ok
}

func (s *Storage) Get(hash uint64) (node *internal.Node, ok bool) {
	s.RLock()
	defer s.RUnlock()
	if node, ok = s.NodeMap[hash]; ok {
		_ = node.IncrementReadCount()
	}
	return
}

func (s *Storage) Del(hash uint64) (node *internal.Node, ok bool) {
	s.Lock()
	defer s.Unlock()
	if node, ok = s.NodeMap[hash]; ok {
		delete(s.NodeMap, hash)
		// 释放存储对象，controller会对其进行检查，判断此对象是否被主动删除
		node.Obj = nil
	}
	return
}
