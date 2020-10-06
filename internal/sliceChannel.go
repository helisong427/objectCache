package internal

const(
	chanCacheSize = 100
	sliceSize = 100
	channelSize = 10000
)

type chanCache struct {
	cache chan chan *Node
}

func newChanCache() (c *chanCache){
	c = &chanCache{
		cache: make(chan chan *Node, chanCacheSize),
	}
	return c
}

func (c *chanCache) get() (ch chan *Node){
	select {
	case ch = <- c.cache:
	default:
		ch = make(chan *Node, channelSize)
	}

	return ch
}

func (c *chanCache) set(ch chan *Node){
	select {
	case c.cache <- ch:
	default:
		close(ch)
		ch = nil
	}
}

type SliceChannel struct {

	Channels []chan *Node
	chanelCache *chanCache
}

func NewSliceChannel() (s *SliceChannel) {

	s = &SliceChannel{
		Channels:    make([]chan *Node, 0, sliceSize),
		chanelCache: newChanCache(),
	}

	s.Channels = append(s.Channels, s.chanelCache.get())

	return s
}


func (s *SliceChannel) GetNode() (node *Node, ok bool){

	for{
		select {
		case node = <- s.Channels[0]:
			return node, true
		default:
			if len(s.Channels) > 1{
				s.chanelCache.set(s.Channels[0])
				s.Channels = s.Channels[1:]
			}else{
				return nil, false
			}
		}
	}
}

func (s *SliceChannel) SetNode(node *Node){

	for{
		select {
		case s.Channels[len(s.Channels) - 1] <- node:
			return
		default:
			s.Channels = append(s.Channels, s.chanelCache.get())
		}
	}

}










