package controller

import (
	"objectCache/internal"
)

const (
	defaultQueueLen = 10
)

//休息队列，由多个queue构成一个可伸缩队列
type restQueue struct {
	restTime uint32 //休息时长，单位为秒
	count    int32
	queues   []*queue
}

func newRestQueue(restTime uint32) (q *restQueue) {

	q = &restQueue{
		restTime: restTime,
		queues:   make([]*queue, 0, defaultQueueLen),
	}

	q.queues = append(q.queues, &queue{})

	return q
}

func (s *restQueue) setRestTime(t uint32) {
	s.restTime = t
}

// 获取到期的所有到期的node
func (s *restQueue) getExpireNodes(now uint32, n []*internal.Node) (nodes []*internal.Node) {

	expireTime := now - s.restTime
	var isEnd bool
	for range s.queues {
		n, isEnd = s.queues[0].fronts(expireTime, n)

		if isEnd {

			if len(s.queues) == 1{
				s.queues[0].head = 0
				s.queues[0].tail = 0
				break
			}

			s.queues = s.queues[1:]
		} else {
			break
		}
	}

	s.count = s.count - int32(len(n))

	return n
}

// addNode 添加一个node到末尾
func (s *restQueue) addNode(n *internal.Node) {

	if !s.queues[len(s.queues)-1].pushBack(n) {
		s.queues = append(s.queues, &queue{})
		s.addNode(n)
	} else {
		s.count++
	}

}
