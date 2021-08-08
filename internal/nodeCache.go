package internal

import (
	"sync"
	"time"
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

func NewNodeCache(size int32) (n *NodeCache) {
	n = &NodeCache{
		nodeChan:   make(chan *Node, size),
		dirtyNodes: make([]*Node, 100),
	}

	go n.recoverNode()

	return n
}

func (c *NodeCache) GetNode() (n *Node) {
	select {
	case n = <-c.nodeChan:
	default:
		n = &Node{}
	}

	return n
}

// 直接存入缓存链中
func (c *NodeCache) SaveNode(n *Node) {

	select {
	case c.nodeChan <- n:
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
//
func (c *NodeCache) recoverNode() {

	// 定时5分钟回收一次node
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
