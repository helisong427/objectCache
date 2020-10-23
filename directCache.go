package objectCache

import (
	"encoding/binary"
	"math"
	"objectCache/internal"
	"objectCache/internal/storage"
	"time"
)

// setDirect 不纳入淘汰管理，直接存储
func setDirect(key []byte, obj interface{}, expireSecond int) {

	hashVal := internal.HashFunc(key)
	segID := hashVal % storage.MaxSegmentSize

	n := c.nodeCache.GetNode()
	ok := c.segments[segID].Set(obj, hashVal, expireSecond, n)
	if !ok {
		c.nodeCache.SaveNode(n)
	}
	return
}
// getDirect 不纳入淘汰管理，直接获取
func getDirect(key []byte) (obj interface{}, ok bool) {
	hashVal := internal.HashFunc(key)
	segID := hashVal % storage.MaxSegmentSize
	node, ok := c.segments[segID].Get(hashVal)
	if !ok {
		return nil, false
	}

	if node.Expire != math.MaxUint32 && uint32(time.Now().Unix()) > node.Expire {
		n, ok := c.segments[segID].Del(hashVal)
		if ok {
			c.nodeCache.SaveNode(n)
		}
		return nil, false
	}

	return node.Obj, ok
}

// delDirect 不纳入淘汰管理，直接删除
func delDirect(key []byte) (ok bool) {
	hashVal := internal.HashFunc(key)
	segID := hashVal % storage.MaxSegmentSize
	var n *internal.Node

	n, ok = c.segments[segID].Del(hashVal)
	if ok {
		c.nodeCache.SaveNode(n)
	}

	return ok
}


//set 缓存字符切片为键值的对象，不纳入淘汰管理。使用默认 _DefaultTopic_
// key为键值；obj为存储对象；expireSecond为过期时间（单位是秒），如果为0则不过期
func SetDirect(key []byte, obj interface{}, expireSecond int) {
	key = append(key, defaultTopic...)
	setDirect(key, obj, expireSecond)
}

// SetInt 缓存一个以int型KEY的对象，不纳入淘汰管理，不纳入淘汰管理。使用默认 _DefaultTopic_
// key为键值；obj为存储对象；expireSecond为过期时间（单位是秒），如果为0则不过期
func SetIntDirect(key int64, obj interface{}, expireSecond int) {
	var bKey [internal.DefaultKeySize]byte
	binary.LittleEndian.PutUint64(bKey[:], uint64(key))
	SetDirect(bKey[:], obj, expireSecond)
}

//Get 根据字符切片型键值获取对象，不纳入淘汰管理。使用默认 _DefaultTopic_
//ok 为是否获取成功，false则说明cache里面已经不存在此对象（可能被淘汰或者被Del()函数删除）
func GetDirect(key []byte) (obj interface{}, ok bool) {
	key = append(key, defaultTopic...)
	return getDirect(key)
}

//GetInt 根据int型键值获取对象，不纳入淘汰管理。使用默认 _DefaultTopic_
//ok 为是否获取成功，false则说明cache里面已经不存在此对象（可能被淘汰或者被Del()函数删除）
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

//SetByTopic 缓存字符切片为键值的对象，不纳入淘汰管理，当对象已经存在返回false。topic为空则使用默认 _DefaultTopic_
// key为键值；obj为存储对象；expireSecond为过期时间（单位是秒），如果为0则不过期
func SetDirectByTopic(topic string, key []byte, obj interface{}, expireSecond int) {
	if topic == ""{
		key = append(key, defaultTopic...)
	}else{
		key = append(key, internal.String2Bytes(topic)...)
	}

	setDirect(key, obj, expireSecond)
}

// SetInt 缓存一个以int型KEY的对象，不纳入淘汰管理，当对象已经存在返回false。topic为空则使用默认 _DefaultTopic_
// key为键值；obj为存储对象；expireSecond为过期时间（单位是秒），如果为0则不过期
func SetIntDirectByTopic(topic string, key int64, obj interface{}, expireSecond int) {
	var bKey [internal.DefaultKeySize]byte
	binary.LittleEndian.PutUint64(bKey[:], uint64(key))
	var hashKey []byte
	if topic == ""{
		hashKey = append(bKey[:], defaultTopic...)
	}else{
		hashKey = append(bKey[:], internal.String2Bytes(topic)...)
	}

	setDirect(hashKey, obj, expireSecond)
}

//Get 根据字符切片型键值获取对象，不纳入淘汰管理，当对象不存在返回false。topic为空则使用默认 _DefaultTopic_
//ok 为是否获取成功，false则说明cache里面已经不存在此对象（可能被淘汰或者被Del()函数删除）
func GetDirectByTopic(topic string, key []byte) (obj interface{}, ok bool) {
	if topic == ""{
		key = append(key, defaultTopic...)
	}else{
		key = append(key, internal.String2Bytes(topic)...)
	}

	return getDirect(key)
}

//GetInt 根据int型键值获取对象，不纳入淘汰管理。topic为空则使用默认 _DefaultTopic_
//ok 为是否获取成功，false则说明cache里面已经不存在此对象（可能被淘汰或者被Del()函数删除）
func GetIntDirectByTopic(topic string, key int64) (obj interface{}, ok bool) {
	var bKey [internal.DefaultKeySize]byte
	binary.LittleEndian.PutUint64(bKey[:], uint64(key))
	var hashKey []byte
	if topic == ""{
		hashKey = append(bKey[:], defaultTopic...)
	}else{
		hashKey = append(bKey[:], internal.String2Bytes(topic)...)
	}
	return getDirect(hashKey)
}


// Del 根据字符切片的键值删除对象，不纳入淘汰管理。topic为空则使用默认 _DefaultTopic_
// ok返回为false则说明对象删除前已经不存在
func DelDirectByTopic(topic string, key []byte) (ok bool) {
	if topic == ""{
		key = append(key, defaultTopic...)
	}else{
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
	if topic == ""{
		hashKey = append(bKey[:], defaultTopic...)
	}else{
		hashKey = append(bKey[:], internal.String2Bytes(topic)...)
	}
	return DelDirect(hashKey)
}