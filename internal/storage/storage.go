package storage

import (
	"cache/internal"
	"math"
	"sync"
	"time"
)

const (
	MaxSegmentSize = 256
)

//存储段
type Storage struct {
	sync.RWMutex
	NodeMap map[uint64]*internal.Node
}

func (s *Storage) Set(obj interface{}, hash uint64, expire int, n *internal.Node) (ok bool) {
	s.Lock()
	var now = time.Now()
	if _, ok = s.NodeMap[hash]; !ok {
		n.Hash = hash
		n.Obj = obj
		n.RestBeginTime = 0
		n.TotalTime = 0
		n.InitReadCount()
		n.LastReadTime = uint32(now.Unix()) - internal.NodeUnitRestTime
		if expire > 0 {
			n.Expire = uint32(now.Add(time.Second * time.Duration(expire)).Unix())
		}else{
			n.Expire = math.MaxUint32 // 2106-02-07 14:28:15 +0800 CST
		}
		s.NodeMap[n.Hash] = n
	}
	s.Unlock()
	return !ok
}

func (s *Storage) Get(hash uint64) (n *internal.Node, ok bool) {
	s.RLock()
	n, ok = s.NodeMap[hash]
	if ok {
		_ = n.IncrementReadCount()
	}
	s.RUnlock()
	return
}

func (s *Storage) Del(hash uint64) (n *internal.Node, ok bool) {
	s.Lock()
	if n, ok = s.NodeMap[hash]; ok {
		delete(s.NodeMap, hash)
		//释放存储对象，controller会对其进行检查，判断此对象是否被主动删除
		n.Obj = nil
	}
	s.Unlock()
	return
}
