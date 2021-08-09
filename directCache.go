package objectCache

import (
	"encoding/binary"
	"objectCache/internal"
	"objectCache/internal/storage"
)

// setDirect 不纳入淘汰管理，直接存储
func setDirect(key []byte, obj interface{}) {

	hashVal := internal.HashFunc(key)
	segID := hashVal % storage.MaxSegmentSize

	node := cache.nodeCache.GetNode()
	ok := cache.segments[segID].Set(obj, hashVal, node)
	if !ok {
		cache.nodeCache.SaveNode(node)
	}
	return
}

// getDirect 不纳入淘汰管理，直接获取
func getDirect(key []byte) (obj interface{}, ok bool) {
	hashVal := internal.HashFunc(key)
	segID := hashVal % storage.MaxSegmentSize
	node, ok := cache.segments[segID].Get(hashVal)
	if !ok {
		return nil, false
	}
	return node.Obj, ok
}

// delDirect 不纳入淘汰管理，直接删除
func delDirect(key []byte) (ok bool) {
	hashVal := internal.HashFunc(key)
	segID := hashVal % storage.MaxSegmentSize
	var node *internal.Node

	node, ok = cache.segments[segID].Del(hashVal)
	if ok {
		cache.nodeCache.SaveNode(node)
	}

	return ok
}

// set 缓存字符切片为键值的对象，不纳入淘汰管理。使用默认 _DefaultTopic_
func SetDirect(key []byte, obj interface{}) {
	key = append(key, defaultTopic...)
	setDirect(key, obj)
}

// SetInt 缓存一个以int型KEY的对象，不纳入淘汰管理，不纳入淘汰管理。使用默认 _DefaultTopic_
func SetIntDirect(key int64, obj interface{}) {
	var bKey [internal.DefaultKeySize]byte
	binary.LittleEndian.PutUint64(bKey[:], uint64(key))
	SetDirect(bKey[:], obj)
}

// Get 根据字符切片型键值获取对象，不纳入淘汰管理。使用默认 _DefaultTopic_
func GetDirect(key []byte) (obj interface{}, ok bool) {
	key = append(key, defaultTopic...)
	return getDirect(key)
}

// GetInt 根据int型键值获取对象，不纳入淘汰管理。使用默认 _DefaultTopic_
func GetIntDirect(key int64) (obj interface{}, ok bool) {
	var bKey [internal.DefaultKeySize]byte
	binary.LittleEndian.PutUint64(bKey[:], uint64(key))
	return GetDirect(bKey[:])
}

// Del 根据字符切片的键值删除对象，不纳入淘汰管理。使用默认 _DefaultTopic_
// ok返回为false则说明对象删除前已经不存在
func DelDirect(key []byte) (ok bool) {
	key = append(key, defaultTopic...)

	return delDirect(key)
}

// DelInt 根据int型键值删除对象，不纳入淘汰管理。使用默认 _DefaultTopic_
// ok返回为false则说明对象删除前已经不存在
func DelIntDirect(key int64) (ok bool) {
	var bKey [internal.DefaultKeySize]byte
	binary.LittleEndian.PutUint64(bKey[:], uint64(key))
	return DelDirect(bKey[:])
}

// SetByTopic 缓存字符切片为键值的对象，不纳入淘汰管理，当对象已经存在返回false。topic为空则使用默认 _DefaultTopic_
func SetDirectByTopic(topic string, key []byte, obj interface{}) {
	if topic == "" {
		key = append(key, defaultTopic...)
	} else {
		key = append(key, internal.String2Bytes(topic)...)
	}

	setDirect(key, obj)
}

// SetInt 缓存一个以int型KEY的对象，不纳入淘汰管理，当对象已经存在返回false。topic为空则使用默认 _DefaultTopic_
func SetIntDirectByTopic(topic string, key int64, obj interface{}) {
	var bKey [internal.DefaultKeySize]byte
	binary.LittleEndian.PutUint64(bKey[:], uint64(key))
	var hashKey []byte
	if topic == "" {
		hashKey = append(bKey[:], defaultTopic...)
	} else {
		hashKey = append(bKey[:], internal.String2Bytes(topic)...)
	}

	setDirect(hashKey, obj)
}

// Get 根据字符切片型键值获取对象，不纳入淘汰管理，当对象不存在返回false。topic为空则使用默认 _DefaultTopic_
func GetDirectByTopic(topic string, key []byte) (obj interface{}, ok bool) {
	if topic == "" {
		key = append(key, defaultTopic...)
	} else {
		key = append(key, internal.String2Bytes(topic)...)
	}

	return getDirect(key)
}

// GetInt 根据int型键值获取对象，不纳入淘汰管理。topic为空则使用默认 _DefaultTopic_
// ok 为是否获取成功，false则说明cache里面已经不存在此对象（可能被淘汰或者被Del()函数删除）
func GetIntDirectByTopic(topic string, key int64) (obj interface{}, ok bool) {
	var bKey [internal.DefaultKeySize]byte
	binary.LittleEndian.PutUint64(bKey[:], uint64(key))
	var hashKey []byte
	if topic == "" {
		hashKey = append(bKey[:], defaultTopic...)
	} else {
		hashKey = append(bKey[:], internal.String2Bytes(topic)...)
	}
	return getDirect(hashKey)
}

// Del 根据字符切片的键值删除对象，不纳入淘汰管理。topic为空则使用默认 _DefaultTopic_
// ok返回为false则说明对象删除前已经不存在
func DelDirectByTopic(topic string, key []byte) (ok bool) {
	if topic == "" {
		key = append(key, defaultTopic...)
	} else {
		key = append(key, internal.String2Bytes(topic)...)
	}

	return delDirect(key)

}

// DelInt 根据int型键值删除对象，不纳入淘汰管理。topic为空则使用默认 _DefaultTopic_
// ok返回为false则说明对象删除前已经不存在
func DelIntDirectByTopic(topic string, key int64) (ok bool) {
	var bKey [8]byte
	binary.LittleEndian.PutUint64(bKey[:], uint64(key))
	var hashKey []byte
	if topic == "" {
		hashKey = append(bKey[:], defaultTopic...)
	} else {
		hashKey = append(bKey[:], internal.String2Bytes(topic)...)
	}
	return DelDirect(hashKey)
}
