package internal

import "github.com/cespare/xxhash/v2"



//HashFunc 返回一个由二进制字符串进行哈希计算出来的数字
func HashFunc(data []byte) uint64 {
	return xxhash.Sum64(data)
}

