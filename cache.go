package cache

import (
	"cache/internal"
	"cache/internal/controller"
	"cache/internal/storage"
	"encoding/binary"
	"sync"
)


type Cache struct {
	bloomFilter bloom
	//objCount      int32
	//objTotalCount int32
	segments   [storage.MaxSegmentSize]*storage.Storage
	nodeCache  *internal.NodeCache //node对象缓存，目的是优化GC
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

//set 缓存一个二进制串为KEY的对象，当存储的类型不匹配返回错误
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
// int型KEY是不会存在hash冲突的
func (c *Cache) SetInt(key int64, obj interface{}, expireSecond int) (ok bool) {
	var bKey [8]byte
	binary.LittleEndian.PutUint64(bKey[:], uint64(key))
	return c.Set(bKey[:], obj, expireSecond)
}

func (c *Cache) Get(key []byte) (obj interface{}, ok bool) {
	hashVal := internal.HashFunc(key)
	segID := hashVal % storage.MaxSegmentSize
	node, ok := c.segments[segID].Get(hashVal)
	if !ok {
		return nil, false
	}

	return node.Obj, ok
}

func (c *Cache) GetInt(key int64) (obj interface{}, ok bool) {
	var bKey [8]byte
	binary.LittleEndian.PutUint64(bKey[:], uint64(key))
	return c.Get(bKey[:])
}

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

func (c *Cache) DelInt(key int64) (ok bool) {
	var bKey [8]byte
	binary.LittleEndian.PutUint64(bKey[:], uint64(key))
	return c.Del(bKey[:])
}


// GetObjCount 获取当前时刻存储对象的个数（是一个瞬时值）。
func (c *Cache) GetObjCount() (count int32) {
	return c.controller.GetTotalCount()
}


// GetObjCount 获取当前时刻存储对象的个数（是一个瞬时值）。
func (c *Cache) GetQueueCount() (result string) {

	return c.controller.GetQueueCount()
}


func (c *Cache) GetDeleteNode() (m sync.Map) {

	return c.controller.GetDeleteNode()
}
