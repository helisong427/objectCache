package internal

import (
	"sync"
	"time"
)

var (
	nodeCache *NodeCache
)

// NodeCache 用于缓存internal.Node对象，减少动态分配给GC造成压力
type NodeCache struct {
	// 用于缓存node的channel
	nodeChan chan *Node

	// 脏数据，storage删除后而controller仍然管理着这个node的时候暂存于此处。
	// 恢复脏数据逻辑：用户删除对象时，此node存与dirtyNodes里面，并标记node.Obj=nil；当controller检查到此node.Obj==nil，则标记node.Hash=0，并放弃
	// 对此node的管理；recoverNode()协程检查到node.Hash==0则将此node加入到缓存的channel里面。
	dirtyNodes []*Node

	dirtyLock sync.Mutex
}

func NewNodeCache(size int32) (nodeCache *NodeCache) {
	nodeCache = &NodeCache{
		nodeChan:   make(chan *Node, size),
		dirtyNodes: make([]*Node, 0, 100),
	}

	go nodeCache.recoverNode()

	return nodeCache
}

func (cache *NodeCache) GetNode() (node *Node) {
	select {
	case node = <-cache.nodeChan:
	default:
		node = &Node{}
	}

	return node
}

// 直接存入缓存链中
func (cache *NodeCache) SaveNode(node *Node) {

	select {
	case cache.nodeChan <- node:
	default:
		node = nil
	}

}

// 存入垃圾脏数据链（dirtyNodes），当数据干净后存入缓存链
func (cache *NodeCache) SaveDirtyNode(node *Node) {

	cache.dirtyLock.Lock()
	defer cache.dirtyLock.Unlock()
	var done bool
	for k, _ := range cache.dirtyNodes {
		if cache.dirtyNodes[k] == nil {
			cache.dirtyNodes[k] = node
			done = true
			break
		}
	}

	if !done {
		cache.dirtyNodes = append(cache.dirtyNodes, node)
	}

}

// 恢复脏数据
func (cache *NodeCache) recoverNode() {

	// 定时5分钟回收一次node
	var t = time.NewTicker(time.Second * 5)
	defer t.Stop()

	for range t.C {
		cache.dirtyLock.Lock()
		for k, _ := range cache.dirtyNodes {
			// hash == 0 则 controller完成的。
			if cache.dirtyNodes[k] != nil && cache.dirtyNodes[k].Hash == 0 {
				cache.SaveNode(cache.dirtyNodes[k])
				cache.dirtyNodes[k] = nil
			}
		}
		cache.dirtyLock.Unlock()
	}
}
