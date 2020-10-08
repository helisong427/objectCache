package internal

import (
	"sync"
)

const (
	chanCacheSize = 100
	sliceSize     = 100
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

//SliceChannel 是一个不限定容量的channel（非阻塞方式存、取）。这样做只为应对存入对象超级密集的情况（如性能测试）。
type SliceChannel struct {
	Channels    []nodeChan
	chanelCache *chanCache
	lock        sync.Mutex

	head, tail *nodeChan
}

func NewSliceChannel() (s *SliceChannel) {

	s = &SliceChannel{
		Channels:    make([]nodeChan, 0, sliceSize),
		chanelCache: newChanCache(),
	}

	s.Channels = append(s.Channels, make(chan *Node, channelSize))

	s.head = &s.Channels[0]
	s.tail = &s.Channels[0]
	return s
}

func (s *SliceChannel) GetNode() (node *Node, ok bool) {

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
				if len(s.Channels) > 1{
					s.chanelCache.set(s.Channels[0])
					s.Channels = s.Channels[1:]
					s.head = &s.Channels[0]
				}else {
					s.lock.Unlock()
					return nil, false
				}
			}
			s.lock.Unlock()
		}
	}

}

func (s *SliceChannel) SetNode(node *Node) {

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
				s.Channels = append(s.Channels, s.chanelCache.get())
				s.tail = &s.Channels[len(s.Channels) - 1]
			}
			s.lock.Unlock()
		}
	}

}
