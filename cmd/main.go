// 模拟随机读写场景，观察淘汰情况

package main

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	_ "net/http/pprof"
	"objectCache"
	"runtime"
	"time"
)

const (
	maxSize = 9000000
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

type Data struct {
	Id int64
	// Name       string
}

type segment struct {
	begin, end       int64
	averageSleepTime int

	data     [1000000]int64
	dataTail int
}

func newSegment(begin, end int64, averageSleepTime int) (se *segment) {

	se = &segment{
		begin:            begin,
		end:              end,
		averageSleepTime: averageSleepTime,
	}
	se.setData()
	go se.start()
	return se
}

func (s *segment) setData() (se *segment) {

	var i int64
	for i = s.begin; i <= s.end; i++ {
		data := Data{Id: i}
		objectCache.SetInt(i, data, 0)
		s.data[i-s.begin] = i
		s.dataTail++
	}

	return s
}

func (s *segment) start() {
	go func(s *segment) {
		var loopSize = 800000
		for {

			for i := 0; i < loopSize; i++ {
				if s.dataTail == 0 {
					fmt.Printf("\nsleep(%d~%d)s 全部淘汰 \n", s.averageSleepTime-5, s.averageSleepTime+5)
					return
				}

				index := int64(rand.Intn(s.dataTail))

				key := s.data[index]

				_, ok := objectCache.GetInt(key)
				if !ok {
					s.dataTail--
					s.data[index] = s.data[s.dataTail]
					s.data[s.dataTail] = 0
					loopSize--
				}
				if loopSize > s.dataTail {
					loopSize = s.dataTail
				}
			}

			sleepTime := s.averageSleepTime + rand.Intn(10) - 5
			time.Sleep(time.Second * time.Duration(sleepTime))
		}

	}(s)

}

// 模拟随机读写场景，观察淘汰情况
func main() {

	go func() {
		http.ListenAndServe("localhost:13001", nil)
	}()
	// go func() {
	//	statsviz.RegisterDefault()
	//	log.Println(http.ListenAndServe("localhost:8080", nil))
	// }()

	objectCache.InitObjectCache(maxSize)

	var segments [10]*segment

	var begin, end int64
	for i := 0; i < 10; i++ {
		begin = int64(1000000*i) + 1
		end = int64(1000000 * (i + 1))
		sleepTime := (2*i + 1) * 15
		segments[i] = newSegment(begin, end, sleepTime)
		time.Sleep(time.Second * 2)
	}

	// var i int64
	// for i = 0; i < randSize; i++ {
	//	ok := c.SetInt(i, Data{Id: i}, 0)
	//	if !ok {
	//		log.Println("setInt失败：", i)
	//	}
	// }
	//
	// go func() {
	//	for{
	//		var ii int64
	//		for ii = 0; ii < randSize; ii++ {
	//			_, ok := c.GetInt(ii)
	//			if !ok {
	//				log.Println("GetInt失败：", i)
	//			}
	//		}
	//
	//		time.Sleep(time.Second*30)
	//	}
	// }()

	fmt.Println("----------------------------------------------------------------------")
	fmt.Print("输入数字：1删除的详细信息；2打印队列信息；3打印内存使用情况。请输入：")

	var number int
	for {
		fmt.Scanln(&number)

		switch number {
		case 1:
			// var count int
			// var totalTime, totalGetCount uint32
			// m := c.GetDeleteNode()
			// m.Range(func(key, value interface{}) bool {
			//	d := value.(*internal.Node)
			//	//var readRate = uint32(0)
			//	//if d.TotalTime != 0 {
			//	//	readRate = d.TotalCount * 10000 / d.TotalTime
			//	//}
			//	//fmt.Printf("    访问总数and时间:%d-%d 当前访问次数:%d 平均访问频率(1w分钟访问次数):%d hash:%d \n",
			//	//	d.TotalCount, d.TotalTime, d.GetCurrentCount() ,readRate, d.Hash)
			//	count++
			//	totalTime += d.TotalTime
			//	totalGetCount += d.TotalCount
			//	return true
			// })
			// var getRate = uint32(0)
			// if totalTime != 0 {
			//	getRate = totalGetCount * 10000 / totalTime
			// }
			// fmt.Printf(" ==================================== \n")
			// fmt.Printf("    被删除的数量：%d 总的访问频率(1w分钟访问次数):%d (%d-%d) \n", count, getRate,totalGetCount, totalTime)
			// fmt.Printf(" ==================================== \n")
		case 2:
			fmt.Println(objectCache.GetQueueCount())
		case 3:
			traceMemStats()
		default:
			fmt.Println("输入错误")
		}
		fmt.Println("----------------------------------------------------------------------")
		fmt.Print("输入数字：1删除的详细信息；2打印队列信息；3打印内存使用情况。请输入：")
	}
}

func traceMemStats() {
	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)
	// log.Printf("Alloc:%d(bytes) HeapIdle:%d(bytes) HeapReleased:%d(bytes)", ms.Alloc, ms.HeapIdle, ms.HeapReleased)
	log.Printf("Sys:%d(kb) lookups:%d Alloc:%d(kb) HeapIdle:%d(kb) HeapReleased:%d(kb)",
		ms.Sys/1024, ms.Lookups, ms.Alloc/1024, ms.HeapIdle/1024, ms.HeapReleased/1024)
}
