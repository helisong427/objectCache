package internal

import (
	"testing"
	"time"
)

func TestNode_GetCurrentCount(t *testing.T) {
	var n = Node{}
	n.IncrementReadCount()
	n.AddCurrentCount(5)
	n.SetTotalCount(1000)

	if n.GetCurrentCount() != 6{
		t.Error("失败1")
	}
	if n.GetTotalCount() != 1000{
		t.Error("失败2")
	}

	n.UpdateNodeData(600)

	if n.GetCurrentCount() != 0{
		t.Error("失败3")
	}
	if n.GetTotalCount() != 1006{
		t.Error("失败4")
	}

	n.AddCurrentCount(9)

	n.SetTotalCount(0x1FFFFF)

	n.UpdateNodeData(600)

	if n.GetCurrentCount() != 0 {
		t.Error("失败5")
	}
	if n.GetTotalCount() !=  (0x1FFFFF / 2) + 9{
		t.Error("失败6")
	}

}

func TestNode_IncrementReadCount(t *testing.T) {
	var n = Node{}
	n.LastReadTime = uint16(time.Now().Unix() - CacheBaseTime)
	n.IncrementReadCount()

	time.Sleep(time.Second * 10)

	n.IncrementReadCount()
	n.IncrementReadCount()

	nn := n.GetCurrentCount()

	if nn != 1{
		t.Error("失败")
	}
}