package controller

import (
	"cache/internal"
)

const (
	queueNodeSize = 5000
)

// 底层用数组实现的队列
type queue struct {
	head, tail int
	queue      [queueNodeSize]*internal.Node
}

//pushBack 从队列尾部入队列。
func (q *queue) pushBack(n *internal.Node) (ok bool) {
	if q.tail >= queueNodeSize {
		return false
	}

	q.queue[q.tail] = n

	q.tail++

	return true
}

//getExpireNodes 从队列头部出队列，一次性把到期的都读出来。
// 返回值isEnd为是否读取到末尾
// 参数expireTime是到期时间
// 参数n是缓存结果的node切片，防止对象逃逸
func (q *queue) fronts(expireTime uint32, n []*internal.Node) (nodes []*internal.Node, isEnd bool) {

	var ii int
	for ; q.head < q.tail; {
		//fmt.Println(expireTime, q.queue[q.head].RestBeginTime)
		if expireTime >= q.queue[q.head].RestBeginTime {
			//fmt.Printf("fronts ==> now:%d > RestBeginTime:%d \n", expireTime, q.queue[q.head].RestBeginTime)
			n = append(n, q.queue[q.head])
			q.queue[q.head] = nil
			ii++
			q.head++
		} else {
			break
		}
	}

	if q.head == queueNodeSize {
		return n, true
	}

	return n, false
}
