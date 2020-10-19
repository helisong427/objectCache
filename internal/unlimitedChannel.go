package internal

import (
	"container/list"
	"sync"
)

const (
	chanCacheSize = 100
	channelSize   = 10000
)

type nodeChan chan *Node

type chanCache struct {
	cache chan nodeChan
	lock  sync.Mutex
}

func newChanCache() (c *chanCache) {
	c = &chanCache{
		cache: make(chan nodeChan, chanCacheSize),
	}
	return c
}

func (c *chanCache) get() (ch nodeChan) {
	select {
	case ch = <-c.cache:
	default:
		ch = make(chan *Node, channelSize)
	}

	return ch
}

func (c *chanCache) set(nc nodeChan) {
	select {
	case c.cache <- nc:
	default:
		c.lock.Lock()
		if nc != nil {
			close(nc)
			nc = nil
		}
		c.lock.Unlock()
	}
}

//UnlimitedChannel 是一个不限定容量的channel。这样做只为应对存入对象超级密集的情况（如性能测试）。
type UnlimitedChannel struct {
	chanelCache *chanCache

	channelList *list.List
	lock        sync.Mutex

	head, tail *nodeChan
}

func NewUnlimitedChannel() (s *UnlimitedChannel) {

	s = &UnlimitedChannel{
		channelList: list.New(),
		chanelCache: newChanCache(),
	}

	nc := s.chanelCache.get()
	s.lock.Lock()
	s.channelList.PushBack(nc)
	s.lock.Unlock()
	s.head = &nc
	s.tail = &nc
	return s
}

func (s *UnlimitedChannel) GetNode() (node *Node, ok bool) {

	for{
		select {
		case node = <- *s.head:
			return node, true
		default:
			s.lock.Lock()
			select {
			case node = <- *s.head:
				s.lock.Unlock()
				return node, true
			default:
				if s.channelList.Len() > 1{
					nc := s.channelList.Remove(s.channelList.Front()).(nodeChan)
					s.chanelCache.set(nc)
					nc = s.channelList.Front().Value.(nodeChan)
					s.head = &nc
				}else{
					s.lock.Unlock()
					return nil, false
				}
			}
			s.lock.Unlock()
		}
	}

}

func (s *UnlimitedChannel) SetNode(node *Node) {

	for{
		select {
		case *s.tail <- node:
			return
		default:
			s.lock.Lock()
			select {
			case *s.tail <- node:
				s.lock.Unlock()
				return
			default:
				nc := s.chanelCache.get()
				s.channelList.PushBack(nc)
				s.tail = &nc
			}
			s.lock.Unlock()
		}
	}

}
