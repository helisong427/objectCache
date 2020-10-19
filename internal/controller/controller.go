package controller

import (
	"fmt"
	"objectCache/internal"
	"objectCache/internal/storage"
	"time"
)



//var InitDelete = 0
//var DeleteCount = 0
//var PrintFlag int
//var DeleteNodeMap sync.Map

/**
 名词解释：
 	访问频率qf（query frequency）：单位时间（UnitRestTime）内平均被访问的次数。
 	稳定性：node当前的在休息队列里期间的qf除以node在进段时间的qf

 淘汰模型：
	1、restQueue等级越高，则存储node的qf越稳定（波动小），用当前node的休息时间内的qf与此node的整个生命周期内的平均qf作比较得出是否稳定。
	2、淘汰原则是：淘汰qf低的node（与整个cache的平均qf作比较）。

 概述：
	淘汰算法：
	1、淘汰逻辑由12个休息队列，包括初始化队列（initialQueue）、稳定性等级队列（restQueue）、待删除队列（destroyQueue）。
	2、加入initialQueue、destroyQueue的node，都能在加入时开始持续10分钟不被检查; 而restQueue按等级依次是10分钟、20分钟、30分钟...
		最后是90分钟。
	3、刚加入进来的node都依次放入initialQueue，呆到10分钟后，判断此node过期则直接删除，没过期则直接存入restQueue的0级队列。
	4、node在restQueue的某一个队列里面呆相应时间后，判断此node过期则直接删除，没有过期则判断其稳定性和访问频率，根据稳定性和访问频率决定将
		node进行升级（放入上一级restQueue）、降级（放入下一级restQueue）、放入destroyQueue。
	5、node在destroyQueue里面呆相应时间后，判断此node过期则直接删除，没有过期则判断其稳定性和访问频率，根据稳定性和访问频率决定将node进行
		删除（淘汰）还是再次加入initialQueue。

	initialQueue：当新加入的node，给其一个固定的时间进行缓冲，避免刚加入进来就被淘汰，也为进入restQueue进行判断其稳定性和访问频率时积累
		数据。node在initialQueue期间只要有一次被访问就会放入restQueue，也给刚加入的node更多的机会。
	restQueue：当node的稳定性越高则放入的restQueue的等级越高，node的检查时间间隔越长，最终减少对CPU的占用。
	destroyQueue：当node在restQueue里面判断稳定性大幅下降且访问频率很低，没有直接淘汰而是存入destroyQueue，也是给次node最后一次机会，当
		次node在次期间稳定性大幅上升，则再次放入initialQueue，这样避免某些对象qf大幅波动导致被淘汰。
*/
type Controller struct {
	// 用于接收用户存储对象时的node，由于sliceChannel是一个不限定容量的channel，这样用户在高并发下也不会由于channel容量占满而被阻塞。
	unlimitedChannel *internal.UnlimitedChannel

	maxCount      int32 //用户设置的最大对象数量
	restNodeCount int32 //在restQueue队列中的对象数量

	TotalCount uint64 // 总的访问次数
	TotalTime  uint64 // 总的时长（每一个node的存活时长的总和）

	segment   *[storage.MaxSegmentSize]*storage.Storage
	nodeCache *internal.NodeCache

	//initialQueue 初始队列，刚存储的对象首先添加到初始队列，初始队列只会淘汰加入后没有被访问的node，
	//其他全部加入levelQueue的1级队列（为在1级队列中做做淘汰判断提供初始数据）。
	initialQueue *restQueue
	//restQueue 分等级的队列，等级越高则存储node的qf越稳定（波动小），休息时间越长。这样越稳定的数据，进行淘汰判断的频率就越低，减少对系统资源的消耗。
	restQueue [internal.LevelSize]*restQueue
	//destroyQueue 删除队列，不会做稳定性判断，如果访问率有增加则添加到levelQueue的1级队列，如果没有增加则确定淘汰对象。
	destroyQueue *restQueue



	updateTotalBeginTime int64
}

func NewController(maxCount int32, segment *[storage.MaxSegmentSize]*storage.Storage,
	nodeCache *internal.NodeCache) (c *Controller) {
	c = &Controller{
		unlimitedChannel:     internal.NewUnlimitedChannel(),
		maxCount:             maxCount,
		segment:              segment,
		nodeCache:            nodeCache,
		destroyQueue:         newRestQueue(uint32(internal.LevelRestStep)),
		initialQueue:         newRestQueue(uint32(internal.LevelRestStep)),
		updateTotalBeginTime: time.Now().Unix(),
	}

	var i uint16
	for i = 0; i < internal.LevelSize; i++ {
		c.restQueue[i] = newRestQueue(uint32(internal.LevelRestStep) * (uint32(i) + 1))
	}

	go c.handle()

	return c
}

func (c *Controller) AddNode(n *internal.Node) {
	c.unlimitedChannel.SetNode(n)
}

func (c *Controller) setTotalCountAndTotalTime(currentCount, currentTime uint32) {
	//
	//if (c.TotalTime + uint64(currentTime)) > 0xffffffffffffffff {
	//	cacheAverageQf := (c.TotalCount * internal.ScaleFactor * internal.NodeUnitRestTime) / c.TotalTime
	//	fmt.Printf("cache等比例缩放%d(%d-%d) ==>", cacheAverageQf, c.TotalTime, c.TotalCount)
	//	rate := float64(c.TotalTime) / float64(c.TotalCount)
	//
	//	c.TotalCount = uint64(float64(c.TotalCount)/rate) + uint64(currentCount)
	//	c.TotalTime = uint64(float64(c.TotalTime)/rate) + uint64(currentTime)
	//	cacheAverageQf = (uint64(c.TotalCount) * internal.ScaleFactor * internal.NodeUnitRestTime) / c.TotalTime
	//	fmt.Printf("%d(%d-%d) %s \n", cacheAverageQf, c.TotalTime, c.TotalCount,c.GetQueueCount())
	//} else {
	//	c.TotalCount += uint64(currentCount)
	//	c.TotalTime += uint64(currentTime)
	//}

	// 总的访问次数和总的访问qf次数（大于休息队列的最大休息时间则等比例缩放）
	now := time.Now().Unix()
	if now - c.updateTotalBeginTime >= int64(internal.LevelSize * internal.LevelRestStep){
		cacheAverageQf := (c.TotalCount * internal.ScaleFactor * internal.NodeUnitRestTime) / c.TotalTime
		fmt.Printf("cache等比例缩放%d(%d-%d) ==>", cacheAverageQf, c.TotalTime, c.TotalCount)

		c.TotalCount = c.TotalCount/2 + uint64(currentCount)
		c.TotalTime = c.TotalTime/2 + uint64(currentTime)
		cacheAverageQf = (c.TotalCount * internal.ScaleFactor * internal.NodeUnitRestTime) / c.TotalTime
		fmt.Printf("%d(%d-%d) %s \n", cacheAverageQf, c.TotalTime, c.TotalCount,c.GetQueueCount())
		c.updateTotalBeginTime = now
	}else{
		c.TotalCount += uint64(currentCount)
		c.TotalTime += uint64(currentTime)
	}
}

//eliminate 进行判断并做淘汰（淘汰算法在此）
// currentCount：当前访问次数； currentRestUnit：当前node此次睡眠期间的单位时间个数
func (c *Controller) eliminate(level int, currentCount, currentTime uint64, node *internal.Node) {

	//当前node在此次睡眠期间的访问频率
	currentQf := (currentCount * internal.ScaleFactor * internal.NodeUnitRestTime) / currentTime

	//当前node整个生命周期的访问频率
	var nodeAverageQf = uint64(0)
	if node.TotalTime != 0{
		nodeAverageQf = uint64(node.TotalCount) * internal.ScaleFactor * internal.NodeUnitRestTime / uint64(node.TotalTime)
	}

	//当前整个缓存的访问频率
	var cacheAverageQf = uint64(0)
	if c.TotalTime != 0{
		cacheAverageQf = (c.TotalCount * internal.ScaleFactor * internal.NodeUnitRestTime) / c.TotalTime
	}


	//计算当前node的稳定性（node在休息时间内的qf占此node的平均qf的比例）
	var nodeStability = uint64(0)
	if nodeAverageQf != 0 {
		nodeStability = currentQf * internal.ScaleFactor / nodeAverageQf
	}

	// 计算出淘汰比例
	eliminateRatio := uint64(c.restNodeCount+c.initialQueue.count+c.destroyQueue.count) * internal.ScaleFactor / uint64(c.maxCount)
	//eliminateRatio := uint64(c.restNodeCount+c.destroyQueue.count) * internal.ScaleFactor / uint64(c.maxCount)

	if eliminateRatio >= 950 {
		eliminateRatio = eliminateRatio - 950
	} else {
		eliminateRatio = 0
	}

	var nodeEliminateRatio = uint64(0)
	if cacheAverageQf != 0 {
		nodeEliminateRatio = (currentQf * internal.ScaleFactor) / (cacheAverageQf * 2)
	}

	//if PrintFlag % 500000 == 0 {
	//	fmt.Printf("==> level:%d; 当前频率 %d:(%d*20000 / %d); node频率 %d:(%d*20000 / %d); cache频率 %d:(%d*20000 / %d); " +
	//		"稳定性:%d 淘汰比例:%d \n", level,currentQf, currentCount,currentTime, nodeAverageQf, node.TotalCount, node.TotalTime,
	//		cacheAverageQf, c.TotalCount, c.TotalTime, nodeStability, eliminateRatio)
	//	PrintFlag++
	//}

	node.UpdateNodeData(uint32(currentTime))

	// nodeStability下降50%，则判断稳定性大幅下降，判断当前node的qf是否达到淘汰比例，达到移入destroyQueue队列。
	// 则当currentQf为0（即在当前休息时间内没有被访问），则必定移入destroyQueue队列。
	if nodeStability < 500 && nodeEliminateRatio <= eliminateRatio {
		//fmt.Printf("%s addNode: restQueue[%d] ==> destroy, key:%d\n", time.Now().Format("15:04:05"), level, node.Hash)
		c.destroyQueue.addNode(node)
		c.restNodeCount--
		//fmt.Printf("结果:destroyQueue \n")
	} else {
		//降级处理

		var levelTemp int
		if nodeStability >= 900 || nodeStability <= 1100 {
			//降级处理：波动在10%则上升1级
			if level < internal.LevelSize-1 {
				levelTemp = level + 1
			} else {
				levelTemp = level
			}
		} else if nodeStability < 800 {
			//降级处理：下降20%以上，则降级处理，多降10%则多降一级
			levelNum := int(800-nodeStability+90) / 100
			if level-levelNum > 0 {
				levelTemp = level - levelNum
			} else {
				levelTemp = 0
			}
		} else {
			//降级处理：下降10%到20%或者上升大于10%，则保留原级
			levelTemp = level
		}

		// 更新cache总数
		c.setTotalCountAndTotalTime(uint32(currentCount), uint32(currentTime))
		c.restQueue[levelTemp].addNode(node)
	}

}

// 动态调整restQueue队列休息的基本时间（第一级队列的休息时间）
// 达到在不同的使用场景下，对系统的压力趋于稳定：当缓存数量过大时，休息队列的休息时间增大，相同时间内缓存对象被检查的次数减少，反之则相反。
func (c *Controller) adjustEliminateParam() {


	var totalCount = c.restNodeCount + c.initialQueue.count + c.destroyQueue.count
	//var totalCount = c.restNodeCount + c.destroyQueue.count

	var countRatio = uint64(totalCount) * internal.ScaleFactor / internal.DefaultObjCount

	if countRatio > 1200 || (500 < countRatio && countRatio < 800) {
		stepTime := internal.LevelRestStep * countRatio / internal.ScaleFactor
		//fmt.Printf("setRestTime: stepTime=%d ", stepTime)
		for k, _ := range c.restQueue {
			//fmt.Printf("restQueue[%d]=%d; ",k, stepTime * uint32(k+1))
			c.restQueue[k].setRestTime(uint32(stepTime) * uint32(k+1))
		}
		//fmt.Printf("\n ")
	} else if countRatio <= 500 {
		stepTime := internal.LevelRestStep / 2
		//fmt.Printf("setRestTime: stepTime=%d ", stepTime)
		for k, _ := range c.restQueue {
			//	fmt.Printf("restQueue[%d]=%d; ",k, stepTime * (k+1))
			c.restQueue[k].setRestTime(uint32(stepTime * uint64(k+1)))
		}
		//fmt.Printf("\n ")
	}
}

func (c *Controller) handle() {

	var getTicker = time.NewTicker(time.Second)
	defer getTicker.Stop()

	var adjustLevelQueueTicker = time.NewTicker(time.Second * time.Duration(internal.LevelRestStep))
	defer adjustLevelQueueTicker.Stop()

	var nodes = make([]*internal.Node, 100)

	for {
		select {
		case t := <-getTicker.C:

			now := uint32(t.Unix())

			// 处理初始队列
			c.initialQueueHandle(nodes, now)
			// 处理休息队列
			c.restQueueHandle(nodes, now)
			//处理删除队列
			c.destroyQueueHandle(nodes, now)

		case <-adjustLevelQueueTicker.C:

			//c.adjustEliminateParam()
			fmt.Print("\n")
			fmt.Print(c.GetQueueCount())
			//fmt.Print("\n")
		default:

			node, ok := c.unlimitedChannel.GetNode()
			if ok {
				//fmt.Printf("%s addNode: user ==> init, key:%d\n",time.Now().Format("15:04:05"), node.Hash)
				node.UpdateNodeData(0)
				c.initialQueue.addNode(node)
				//c.restQueue[0].addNode(node)

			} else {
				time.Sleep(time.Millisecond)
			}
		}
	}
}

// 处理初始队列
func (c *Controller) initialQueueHandle(nodes []*internal.Node, now uint32) {
	//清空切片
	nodes = nodes[0:0]

	var currentCount uint32

	nodes = c.initialQueue.getExpireNodes(now, nodes)

	for k, _ := range nodes {

		if c.directEliminate(nodes[k], now) {
			//fmt.Printf("init\n")
			continue
		}

		// 在初始队列中没有被访问，则直接淘汰
		if nodes[k].GetCurrentCount() == 0{
			_, ok := c.segment[nodes[k].Hash%storage.MaxSegmentSize].Del(nodes[k].Hash)
			if ok {
				c.nodeCache.SaveNode(nodes[k])
			} else {
				nodes[k].Hash = 0
			}
			//InitDelete++
		}

		currentCount = nodes[k].GetCurrentCount()

		c.setTotalCountAndTotalTime(currentCount, c.initialQueue.restTime)

		nodes[k].UpdateNodeData(c.initialQueue.restTime)

		//fmt.Printf("%s addNode: init ==> restQueue[0], key:%d\n",time.Now().Format("15:04:05"), nodes[k].Hash)
		c.restQueue[0].addNode(nodes[k])

		c.restNodeCount++
	}
}

// 处理休息队列
func (c *Controller) restQueueHandle(nodes []*internal.Node, now uint32) {

	var currentCount, currentTime uint32

	// 处理休息队列
	for k, _ := range c.restQueue {

		//清空切片
		nodes = nodes[0:0]

		nodes = c.restQueue[k].getExpireNodes(now, nodes)

		for kk, _ := range nodes {
			if c.directEliminate(nodes[kk], now) {
				//fmt.Printf("%d\n", k)
				c.restNodeCount--
				continue
			}

			currentCount = nodes[kk].GetCurrentCount()

			currentTime = now - nodes[kk].RestBeginTime


			// 对当前的node进行淘汰ls
			c.eliminate(k, uint64(currentCount), uint64(currentTime), nodes[kk])
		}
	}
}

// 处理删除队列
func (c *Controller) destroyQueueHandle(nodes []*internal.Node, now uint32) {

	//清空切片
	nodes = nodes[0:0]

	nodes = c.destroyQueue.getExpireNodes(now, nodes)

	var deleteCount = c.maxCount - c.restNodeCount - c.initialQueue.count - c.destroyQueue.count
	//var deleteCount = c.maxCount - c.restNodeCount - c.destroyQueue.count
	for k, _ := range nodes {

		if c.directEliminate(nodes[k], now) {
			//fmt.Printf("destroy\n")
			continue
		}

		currentCount := uint64(nodes[k].GetCurrentCount())


		//当前node的访问频率
		averageQf := uint64(nodes[k].TotalCount) * internal.ScaleFactor * internal.NodeUnitRestTime / uint64(nodes[k].TotalTime)

		//当前node在此次睡眠期间的访问频率
		currentQf := currentCount * internal.ScaleFactor * internal.NodeUnitRestTime / uint64(c.destroyQueue.restTime)

		// node 的稳定性
		var nodeStability = uint64(0)
		if averageQf != 0 {
			nodeStability = currentQf * internal.ScaleFactor / averageQf
		}

		// 对于淘汰队列中到期，需要删除的node进行捡漏：
		//1、在destroyQueue队列中休息期间的访问率达到此node的平均访问率；
		//2、在destroyQueue队列中休息期间的访问率达到此node的平均访问率的70%，并且整个系统没有待淘汰的数量
		if nodeStability >= 1000 || (deleteCount <= 0 && nodeStability >= 700) {

			//fmt.Printf("%s addNode: destroy ==> restQueue[0], key:%d\n",time.Now().Format("15:04:05"), nodes[k].Hash)
			c.setTotalCountAndTotalTime(uint32(currentCount), c.destroyQueue.restTime)
			nodes[k].UpdateNodeData(c.destroyQueue.restTime)
			c.restQueue[0].addNode(nodes[k])
			c.restNodeCount++
		} else {

			// test
			//DeleteCount++
			//DeleteNodeMap.Store(nodes[k].Hash, nodes[k])

			//fmt.Println("delete the node: ", nodes[k].Hash)
			_, ok := c.segment[nodes[k].Hash%storage.MaxSegmentSize].Del(nodes[k].Hash)
			if ok {
				c.nodeCache.SaveNode(nodes[k])
			} else {
				nodes[k].Hash = 0
			}

			deleteCount--
		}
	}
}

// directEliminate 进行直接淘汰：1、被外部删除；2、对象过期。
func (c *Controller) directEliminate(node *internal.Node, now uint32) (ok bool) {
	// 被用户主动删除，直接丢弃
	if node.Obj == nil {
		//此处清除hash，作为recoverNode()进行判断的依据
		node.Hash = 0

		//fmt.Printf("directEliminate==> 用户删除 key: %d-", node.Hash)

		return true
	}
	// 过期，直接调用接口删除
	if now >= node.Expire {
		_, ok := c.segment[node.Hash%storage.MaxSegmentSize].Del(node.Hash)
		if ok {
			c.nodeCache.SaveNode(node)
		} else {
			node.Hash = 0
		}
		//fmt.Printf("directEliminate==> 过期    key: %d-", node.Hash)
		return true
	}

	return false
}

func (c *Controller) GetTotalCount() (count int32) {

	return c.restNodeCount + c.initialQueue.count + c.destroyQueue.count
}

func (c *Controller) GetQueueCount() (result string) {

	eliminateRatio := uint64(c.restNodeCount+c.initialQueue.count+c.destroyQueue.count) * internal.ScaleFactor / uint64(c.maxCount)

	if eliminateRatio >= 950 {
		eliminateRatio = eliminateRatio - 950
	} else {
		eliminateRatio = 0
	}
	var cacheAverageQf uint64
	if c.TotalCount != 0 {
		cacheAverageQf = (c.TotalCount * internal.ScaleFactor * internal.NodeUnitRestTime) / c.TotalTime
	}
	result = fmt.Sprintf("node count: %d-%d-%d-%d-%d-%d-%d-%d-%d-%d-%d-%d 总数量:%d 淘汰率：%d " +
		"平均访问频率:%d(%d * 20000 - %d)", c.initialQueue.count, c.restQueue[0].count,
		c.restQueue[1].count, c.restQueue[2].count, c.restQueue[3].count, c.restQueue[4].count, c.restQueue[5].count,
		c.restQueue[6].count, c.restQueue[7].count, c.restQueue[8].count, c.restQueue[9].count, c.destroyQueue.count,
		c.initialQueue.count+c.restNodeCount+c.destroyQueue.count, eliminateRatio, cacheAverageQf,
		c.TotalCount, c.TotalTime)


	return result
}

//func (c *Controller) GetDeleteNode() (m sync.Map) {
//
//	return DeleteNodeMap
//}
