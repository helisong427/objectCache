package objectCache_test

import (
	"fmt"
	"objectCache"
	"time"
)

type testData struct {
	id int
	name string
}

func ExampleNewCache() {
	_ = objectCache.NewCache(1e5)
	ca := objectCache.NewDefaultCache()
	fmt.Println(ca.GetObjCount())
	// Output:
	// 0
}


func ExampleCache_SetAndGet() {
	var ca = objectCache.NewDefaultCache()

	d := testData{id: 100,name: "test1"}

	ca.Set([]byte("test1"), d, 0)

	obj, ok := ca.Get([]byte("test1"))
	if ok {
		fmt.Println(obj.(testData).id)
	}

	// Output:
	// 100

}

func ExampleCache_SetIntAndGetInt() {
	var ca = objectCache.NewDefaultCache()

	d := testData{id: 100,name: "test1"}

	ca.SetInt(100, d, 0)

	obj, ok := ca.GetInt(100)
	if ok {
		fmt.Println(obj.(testData).name)
	}

	// Output:
	// test1

}

func ExampleCache_SetExpire() {
	var ca = objectCache.NewDefaultCache()

	d := testData{id: 100,name: "test1"}

	ca.SetInt(100, d, 5)

	time.Sleep(time.Second*6)
	obj, ok := ca.GetInt(100)
	if ok {
		fmt.Println(obj.(testData).name)
	}else{
		fmt.Println("get failed")
	}

	// Output:
	// get failed

}