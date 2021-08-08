package controller

type queueCache struct {
	cache chan *queue
}

var queueCacheObj = queueCache{cache: make(chan *queue, 100)}

func (s *queueCache) getQueue() (q *queue) {

	select {
	case q = <-s.cache:
	default:
		q = &queue{}
	}

	return q
}

func (s *queueCache) setQueue(q *queue) {

	select {
	case s.cache <- q:
	default:
		q = nil
	}
}
