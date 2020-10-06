package internal

import (
	"sync"
	"time"
)


//NodeCache 用于缓存Node对象，减少动态分配给GC造成压力
type NodeCache struct {
	cache chan *Node

	dirtyNodes []*Node //脏数据，storage删除后而controller仍然管理着这个node的时候，暂存于此处。
	dirtyLock  sync.Mutex
}

func NewNodeCache(size int32) (n *NodeCache) {
	n = &NodeCache{
		cache:      make(chan *Node, size),
		dirtyNodes: make([]*Node, 100),
	}

	go n.recoverNode()

	return n
}

func (c *NodeCache) GetNode() (n *Node) {
	select {
	case n = <-c.cache:
	default:
		n = &Node{}
	}

	return n
}


// 直接存入缓存链中
func (c *NodeCache) SaveNode(n *Node) {

	select {
	case c.cache <- n:
	default:
		n = nil
	}

}


// 存入垃圾脏数据链（dirtyNodes），当数据干净后存入缓存链
func (c *NodeCache) SaveDirtyNode(n *Node) {

	c.dirtyLock.Lock()
	var done bool
	for k, _ := range c.dirtyNodes {
		if c.dirtyNodes[k] == nil {
			c.dirtyNodes[k] = n
			done = true
			break
		}
	}

	if !done {
		c.dirtyNodes = append(c.dirtyNodes, n)
	}
	c.dirtyLock.Unlock()
}


// 恢复脏数据
func (c *NodeCache) recoverNode() {

	//定时5分钟回收一次node
	var t = time.NewTicker(time.Second * 5)
	defer t.Stop()

	for range t.C {
		c.dirtyLock.Lock()
		for k, _ := range c.dirtyNodes {
			// hash == 0 则 controller完成的。
			if c.dirtyNodes[k] != nil && c.dirtyNodes[k].Hash == 0 {
				c.SaveNode(c.dirtyNodes[k])
				c.dirtyNodes[k] = nil
			}
		}
		c.dirtyLock.Unlock()
	}
}