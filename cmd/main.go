package main

import (
	"fmt"
	_ "net/http/pprof"
)

const scaleFactor = 10000

func main() {
	var totalCount = uint32(10000000)
	aa := uint64(totalCount) * scaleFactor
	var countRatio = uint32(aa / 1e6)

	fmt.Println(aa, countRatio)
	return


}
