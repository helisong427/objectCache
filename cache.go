package cache

import (
	"cache/internal"
	"cache/internal/controller"
	"cache/internal/storage"
	"encoding/binary"
	"math"
	"time"
)

var c *Cache

// 整个cache主要包含3个部分：
// segments: 用于存储对象，使用了256个storage.Storage组成，每一个storage.Storage持有一个读写锁，这样实现就减小了锁的粒度，整个cache就支持最大256个并发操作。
// nodeCache：是internal.Node(是存储对象用的，是cache存储的基本单元)的缓存池，避免动态创建internal.Node，整个cache就大幅减小对GC的压力。
// controller：对象控制器，用于对所有存储对象进行监控，根据对象的访问频率和访问的稳定性进行淘汰，还会删除到期的对象。
type Cache struct {
	segments   [storage.MaxSegmentSize]*storage.Storage
	nodeCache  *internal.NodeCache
	controller *controller.Controller
}

//NewCache 创建一个缓存集合
//objMaxCount 参数用于限制最大缓存数量，其范围为[1w ~ 10000w]，如果objMaxCount没有在这个范围，则采用默认值100w
func NewCache(objMaxCount int32) (cache *Cache) {
	cache = &Cache{
		nodeCache: internal.NewNodeCache(objMaxCount / 4),
	}

	for i := 0; i < storage.MaxSegmentSize; i++ {
		cache.segments[i] = &storage.Storage{NodeMap: make(map[uint64]*internal.Node)}
	}

	if objMaxCount > 1e8 || objMaxCount < 1e4 {
		objMaxCount = 1e6
	}
	cache.controller = controller.NewController(objMaxCount, &cache.segments, cache.nodeCache)

	return
}

//NewDefaultCache 返回一个cache对象，可以存储任意类型，最大缓存数量为默认值8*65535
func NewDefaultCache() (cache *Cache) {
	return NewCache(0)
}

//set 缓存字符切片为键值的对象，当存储的类型不匹配返回错误
// key为键值；obj为存储对象；expireSecond为过期时间（单位是秒），如果为0则不过期
func (c *Cache) Set(key []byte, obj interface{}, expireSecond int) (ok bool) {

	hashVal := internal.HashFunc(key)
	segID := hashVal % storage.MaxSegmentSize

	n := c.nodeCache.GetNode()
	ok = c.segments[segID].Set(obj, hashVal, expireSecond, n)
	if ok {
		c.controller.AddNode(n)
	} else {
		c.nodeCache.SaveNode(n)
	}
	return
}

// SetInt 缓存一个以int型KEY的对象，当存储的类型不匹配返回错误
// key为键值；obj为存储对象；expireSecond为过期时间（单位是秒），如果为0则不过期
func (c *Cache) SetInt(key int64, obj interface{}, expireSecond int) (ok bool) {
	var bKey [8]byte
	binary.LittleEndian.PutUint64(bKey[:], uint64(key))
	return c.Set(bKey[:], obj, expireSecond)
}

//Get 根据字符切片型键值获取对象
//ok 为是否获取成功，false则说明cache里面已经不存在此对象（可能被淘汰或者被Del()函数删除）
func (c *Cache) Get(key []byte) (obj interface{}, ok bool) {
	hashVal := internal.HashFunc(key)
	segID := hashVal % storage.MaxSegmentSize
	node, ok := c.segments[segID].Get(hashVal)
	if !ok {
		return nil, false
	}

	if node.Expire != math.MaxUint32 && uint32(time.Now().Unix()) > node.Expire {
		n, ok := c.segments[segID].Del(hashVal)
		if ok {
			c.nodeCache.SaveDirtyNode(n)
		}
		return nil, false
	}

	return node.Obj, ok
}

//GetInt 根据int型键值获取对象
//ok 为是否获取成功，false则说明cache里面已经不存在此对象（可能被淘汰或者被Del()函数删除）
func (c *Cache) GetInt(key int64) (obj interface{}, ok bool) {
	var bKey [8]byte
	binary.LittleEndian.PutUint64(bKey[:], uint64(key))
	return c.Get(bKey[:])
}


// Del 根据字符切片的键值删除对象
// ok返回为false则说明对象删除前已经不存在
func (c *Cache) Del(key []byte) (ok bool) {
	hashVal := internal.HashFunc(key)
	segID := hashVal % storage.MaxSegmentSize
	var n *internal.Node

	n, ok = c.segments[segID].Del(hashVal)
	if ok {
		c.nodeCache.SaveDirtyNode(n)
	}

	return ok
}


// DelInt 根据int型键值删除对象
// ok返回为false则说明对象删除前已经不存在
func (c *Cache) DelInt(key int64) (ok bool) {
	var bKey [8]byte
	binary.LittleEndian.PutUint64(bKey[:], uint64(key))
	return c.Del(bKey[:])
}


// GetObjCount 获取当前时刻存储对象的个数（是一个瞬时值，可能并不是你预期的值）。
func (c *Cache) GetObjCount() (count int32) {
	return c.controller.GetTotalCount()
}


// GetQueueCount 测试使用
func (c *Cache) GetQueueCount() (result string) {

	return c.controller.GetQueueCount()
}

// GetDeleteNode 测试使用
//func (c *Cache) GetDeleteNode() (m sync.Map) {
//
//	return c.controller.GetDeleteNode()
//}
