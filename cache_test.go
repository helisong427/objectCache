package objectCache_test

import (
	"fmt"
	"objectCache"
	"time"
)

type testData struct {
	id   int
	name string
}

func ExampleNewCache() {
	objectCache.InitObjectCache(1e5)
	objectCache.InitDefaultObjectCache()
	fmt.Println(objectCache.GetObjCount())
	// Output:
	// 0
}

func ExampleCache_SetAndGet() {
	objectCache.InitDefaultObjectCache()

	d := testData{id: 100, name: "test1"}

	objectCache.Set([]byte("test1"), d, 0)

	obj, ok := objectCache.Get([]byte("test1"))
	if ok {
		fmt.Println(obj.(testData).id)
	}

	// Output:
	// 100

}

func ExampleCache_SetIntAndGetInt() {
	objectCache.InitDefaultObjectCache()

	d := testData{id: 1001, name: "SetIntAndGetInt"}

	objectCache.SetInt(1001, d, 0)

	obj, ok := objectCache.GetInt(1001)
	if ok {
		fmt.Println(obj.(testData).name)
	}

	// Output:
	// SetIntAndGetInt

}

func ExampleCache_SetExpire() {
	objectCache.InitDefaultObjectCache()

	d := testData{id: 1002, name: "SetExpire"}

	objectCache.SetInt(1002, d, 5)

	time.Sleep(time.Second * 6)
	obj, ok := objectCache.GetInt(1002)
	if ok {
		fmt.Println(obj.(testData).name)
	} else {
		fmt.Println("get failed")
	}

	// Output:
	// get failed

}

func ExampleCache_SetAndGetByTopic() {
	objectCache.InitDefaultObjectCache()

	d := testData{id: 1005, name: "SetAndGetByTopic"}

	objectCache.SetByTopic("exampleTest", []byte("SetAndGetByTopic"), d, 0)

	obj, ok := objectCache.GetByTopic("exampleTest", []byte("SetAndGetByTopic"))
	if ok {
		fmt.Println(obj.(testData).id)
	}

	// Output:
	// 1005

}

func ExampleCache_SetIntAndGetIntByTopic() {
	objectCache.InitDefaultObjectCache()

	d := testData{id: 1004, name: "SetIntAndGetIntByTopic"}

	objectCache.SetIntByTopic("exampleTest", 1004, d, 0)

	obj, ok := objectCache.GetIntByTopic("exampleTest", 1004)
	if ok {
		fmt.Println(obj.(testData).name)
	}

	// Output:
	// SetIntAndGetIntByTopic

}

func ExampleCache_SetExpireByTopic() {
	objectCache.InitDefaultObjectCache()

	d := testData{id: 1003, name: "SetExpireByTopic"}

	objectCache.SetIntByTopic("exampleTest", 1003, d, 5)

	time.Sleep(time.Second * 6)
	obj, ok := objectCache.GetIntByTopic("exampleTest", 1003)
	if ok {
		fmt.Println(obj.(testData).name)
	} else {
		fmt.Println("get failed")
	}

	// Output:
	// get failed

}
