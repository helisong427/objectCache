package controller

import (
	"container/list"
	"objectCache/internal"
)

// 休息队列，由多个queue构成一个可伸缩队列
type restQueue struct {
	restTime  uint32 // 休息时长，单位为秒
	count     int32
	queueList *list.List
}

func newRestQueue(restTime uint32) (q *restQueue) {

	q = &restQueue{
		restTime:  restTime,
		queueList: list.New(),
	}

	q.queueList.PushBack(queueCacheObj.getQueue())

	return q
}

func (s *restQueue) setRestTime(t uint32) {
	s.restTime = t
}

// 获取到期的所有到期的node
func (s *restQueue) getExpireNodes(now uint32, n []*internal.Node) (nodes []*internal.Node) {

	expireTime := now - s.restTime
	var isEnd bool
	var q *queue
	for i := s.queueList.Front(); i != nil; i = i.Next() {
		q = i.Value.(*queue)
		n, isEnd = q.fronts(expireTime, n)
		if isEnd {
			q.reset()

			if s.queueList.Len() == 1 {
				break
			}

			s.queueList.Remove(i)
			queueCacheObj.setQueue(q)
		} else {
			break
		}

	}
	s.count = s.count - int32(len(n))
	return n
}

// addNode 添加一个node到末尾
func (s *restQueue) addNode(n *internal.Node) {

	if !s.queueList.Back().Value.(*queue).pushBack(n) {
		s.queueList.PushBack(queueCacheObj.getQueue())
		s.addNode(n)
	} else {
		s.count++
	}

}
