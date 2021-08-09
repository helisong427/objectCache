package internal

import (
	"fmt"
	"math"
	"testing"
	"time"
	"unsafe"
)

func TestNode_GetCurrentCount(t *testing.T) {
	var n = Node{}
	n.IncrementReadCount()
	n.AddCurrentCount(5)
	n.TotalCount = 1000

	if n.GetCurrentCount() != 6 {
		t.Error("失败1")
	}

	n.UpdateData(600)

	if n.GetCurrentCount() != 0 {
		t.Error("失败3")
	}
	if n.TotalCount != 1006 {
		t.Error("失败4")
	}

	n.AddCurrentCount(9)

	n.UpdateData(600)

	if n.GetCurrentCount() != 0 {
		t.Error("失败5")
	}
	if n.TotalCount != 1006+9 {
		t.Error("失败6")
	}

}

func TestNode_IncrementReadCount(t *testing.T) {
	var n = Node{}
	n.LastReadTime = uint32(time.Now().Unix())
	n.IncrementReadCount()

	time.Sleep(time.Second * 10)

	n.IncrementReadCount()
	n.IncrementReadCount()

	nn := n.GetCurrentCount()

	if nn != 1 {
		t.Error("失败")
	}
}

func Test11(t *testing.T) {
	fmt.Println(unsafe.Sizeof(uint32(1)))
	fmt.Println(math.MaxUint16)
}