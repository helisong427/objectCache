package internal

import (
	"github.com/cespare/xxhash/v2"
	"unsafe"
)

// HashFunc 返回一个由二进制字符串进行哈希计算出来的数字
func HashFunc(data []byte) uint64 {
	return xxhash.Sum64(data)
}

func Bytes2String(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}

func String2Bytes(s string) []byte {
	x := (*[2]uintptr)(unsafe.Pointer(&s))
	h := [3]uintptr{x[0], x[1], x[1]}
	return *(*[]byte)(unsafe.Pointer(&h))
}
