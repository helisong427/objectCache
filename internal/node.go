package internal

import (
	"sync/atomic"
	"time"
)

const ()

//存储的基本单元(sizeof = 48)
type Node struct {
	//最后被访问的时间，单位为秒
	LastReadTime uint32

	//淘汰检查开始时间，controller使用，单位为秒
	RestBeginTime uint32
	// 存储当前休眠时间内被访问的单位时间个数（单位时间内被访问则加1）
	currentCount uint32

	//总的单位时间访问次数
	TotalCount uint32
	// node存活时长，单位为秒
	TotalTime uint32

	//过期时间，单位为秒，Unix time
	Expire uint32

	//hash 值
	Hash uint64

	//存储的对象
	Obj interface{}
}

//ResetRestBeginTimeAndCurrentCount 重置restBeginTime、currentCount
//func (n *Node) ResetRestBeginTimeAndCurrentCount() (node *Node) {
//
//	n.RestBeginTime = uint32(time.Now().Unix())
//	atomic.StoreUint32(&n.currentCount, 0)
//	return n
//}

//UpdateNodeData 当node从休息队列中取出来后更新RestUnitCount、currentCount
func (n *Node) UpdateNodeData(CurrentTime uint32) {

	//设置totalCount
	n.TotalCount += n.GetCurrentCount()
	//设置TotalRestUnitTime
	n.TotalTime += CurrentTime

	// TotalTime、TotalCount是用于计算最近访问频率，这个最近的期限定为restQueue休息的最大时间，当超过这个时间就等比例缩放1倍
	if uint64(n.TotalTime) >= LevelRestStep*LevelSize {

		//nodeAverageQf := uint64(n.TotalCount) * 1000 * NodeUnitRestTime / uint64(n.TotalTime)

		//fmt.Printf("node等比例缩放%d(%d-%d) ==>", nodeAverageQf, n.TotalTime, n.TotalCount)
		n.TotalTime = n.TotalTime / 2
		n.TotalCount = n.TotalCount / 2
		//nodeAverageQf = uint64(n.TotalCount) * 1000 * NodeUnitRestTime / uint64(n.TotalTime)
		//fmt.Printf("%d(%d-%d) \n", nodeAverageQf, n.TotalTime, n.TotalCount)
	}

	n.RestBeginTime = uint32(time.Now().Unix())
	atomic.StoreUint32(&n.currentCount, 0)

}

// 获取当前休息时间内的读取次数
func (n *Node) GetCurrentCount() (count uint32) {
	return atomic.LoadUint32(&n.currentCount)
}

func (n *Node) AddCurrentCount(count uint32) {
	atomic.AddUint32(&n.currentCount, count)
}


func (n *Node) InitReadCount() {
	n.currentCount = 0
}

func (n *Node) IncrementReadCount() (ok bool) {

	// 在单位时间内，被访问多次只计算1次
	now := uint32(time.Now().Unix())
	if (now - n.LastReadTime) >= NodeUnitRestTime {
		atomic.AddUint32(&n.currentCount, 1)
		n.LastReadTime = now
		return true
	}

	return false
}
