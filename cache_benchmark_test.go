package objectCache

import (
	"fmt"
	"math/rand"
	"sync"
	"testing"
	"time"
)

type dataDemo struct {
	id   int
	name string
}

var mapDemo = make(map[string]*dataDemo)
var syncMapDemo sync.Map

const N = 6e7 // 6000w

func init() {
	// initCache()
	// initMap()
	// initSyncMap()
}

func initCache() {
	InitObjectCache(65535 * 200)
	for i := 0; i < 60000000; i++ {
		SetInt(int64(i), &dataDemo{
			id:   i,
			name: "haha",
		}, 10)

	}
	fmt.Println("init cache end")
	// timeGC()
}

func initMap() {
	for i := 0; i < 60000000; i++ {
		key := fmt.Sprintf("%d", i)
		mapDemo[key] = &dataDemo{
			id:   i,
			name: "haha",
		}
	}
	fmt.Println("init map end")
	// timeGC()
}

func initSyncMap() {
	for i := 0; i < 60000000; i++ {
		key := fmt.Sprintf("%d", i)
		syncMapDemo.Store(key, &dataDemo{
			id:   i,
			name: "haha",
		})
	}
	fmt.Println("init sync map end")
	// timeGC()
}

/**
bruce@:~/go/src/cache$ go test -bench=BenchmarkCache_Set -run=none -count=5
goos: linux
goarch: amd64
pkg: cache
BenchmarkCache_Set-12            3000000               411 ns/op              93 B/op          4 allocs/op
BenchmarkCache_Set-12            5000000               495 ns/op              97 B/op          4 allocs/op
BenchmarkCache_Set-12            5000000               311 ns/op              64 B/op          3 allocs/op
BenchmarkCache_Set-12            5000000               302 ns/op              64 B/op          3 allocs/op
BenchmarkCache_Set-12            5000000               299 ns/op              64 B/op          3 allocs/op
PASS
ok      cache   31.036s
*/
func BenchmarkCache_Set(b *testing.B) {
	InitObjectCache(65535 * 200)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("%d", i)
		Set([]byte(key), &dataDemo{
			id:   i,
			name: "haha",
		}, 10)
	}
}

/**

bruce@:~/go/src/cache$ go test -bench=BenchmarkMap_Set -run=none -count=5
goos: linux
goarch: amd64
pkg: cache
BenchmarkMap_Set1-12             3000000               349 ns/op              91 B/op          3 allocs/op
BenchmarkMap_Set1-12            10000000               437 ns/op             121 B/op          3 allocs/op
BenchmarkMap_Set1-12            10000000               246 ns/op              48 B/op          2 allocs/op
BenchmarkMap_Set1-12            10000000               241 ns/op              48 B/op          2 allocs/op
BenchmarkMap_Set1-12            10000000               241 ns/op              48 B/op          2 allocs/op
PASS
ok      cache   32.537s
*/
func BenchmarkMap_Set1(b *testing.B) {

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("%d", i)
		mapDemo[key] = &dataDemo{
			id:   i,
			name: "haha",
		}
	}
}

/**
bruce@:~/go/src/cache$ go test -bench=BenchmarkSyncCache_Set -run=none  -count=5
goos: linux
goarch: amd64
pkg: cache
BenchmarkSyncCache_Set-12        5000000               351 ns/op             144 B/op          2 allocs/op
BenchmarkSyncCache_Set-12        5000000               280 ns/op              96 B/op          2 allocs/op
BenchmarkSyncCache_Set-12        5000000               283 ns/op             222 B/op          2 allocs/op
BenchmarkSyncCache_Set-12        5000000               260 ns/op              99 B/op          2 allocs/op
BenchmarkSyncCache_Set-12        5000000               323 ns/op             356 B/op          2 allocs/op
PASS
ok      cache   180.037s
*/
func BenchmarkSyncCache_Set(b *testing.B) {
	InitObjectCache(65535 * 200)
	rand.Seed(time.Now().UnixNano())
	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			id := rand.Int63()
			// key := fmt.Sprintf("%d", id)
			SetInt(id, &dataDemo{
				id:   int(id),
				name: "haha",
			}, 10)

		}
	})
}

/**
bruce@:~/go/src/cache$ go test -bench=BenchmarkSyncMap_Set -run=none -count=5
goos: linux
goarch: amd64
pkg: cache
BenchmarkSyncMap_Set-12          1000000              1021 ns/op             250 B/op          7 allocs/op
BenchmarkSyncMap_Set-12          1000000              1019 ns/op             251 B/op          7 allocs/op
BenchmarkSyncMap_Set-12          3000000               983 ns/op             212 B/op          7 allocs/op
BenchmarkSyncMap_Set-12          1000000              1597 ns/op             608 B/op          7 allocs/op
BenchmarkSyncMap_Set-12           500000              3002 ns/op             127 B/op          7 allocs/op
PASS
ok      cache   32.021s

*/
func BenchmarkSyncMap_Set(b *testing.B) {
	rand.Seed(time.Now().UnixNano())
	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			id := rand.Int()
			key := fmt.Sprintf("%d", id)
			syncMapDemo.Store(key, &dataDemo{
				id:   id,
				name: "haha",
			})
		}
	})
}

/**
bruce@:~/go/src/cache$  go test -bench=BenchmarkCache_Get -v -run=none
init cache end
goos: linux
goarch: amd64
pkg: cache
BenchmarkCache_Get-12            5000000               369 ns/op               0 B/op          0 allocs/op
PASS
ok      cache   164.477s
*/
func BenchmarkCache_Get(b *testing.B) {
	InitObjectCache(65535 * 200)
	rand.Seed(time.Now().UnixNano())

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		key := rand.Int63n(60000000)
		_, ok := GetInt(key)
		if !ok {
			b.Errorf("获取错误 key:%d", key)
		}
	}
}

/**
bruce@:~/go/src/cache$ go test -bench=BenchmarkMap_Get -v -run=none
init map end
goos: linux
goarch: amd64
pkg: cache
BenchmarkMap_Get-12      3000000               424 ns/op              16 B/op          2 allocs/op
PASS
ok      cache   72.431s
*/
func BenchmarkMap_Get(b *testing.B) {
	rand.Seed(time.Now().UnixNano())
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("%d", rand.Int63n(60000000))
		_, ok := mapDemo[key]
		if !ok {
			b.Errorf(" key:%s", key)
		}
	}
}

/**
bruce@:~/go/src/cache$ go test -bench=BenchmarkSyncCache_Get -v -run=none
init cache end
goos: linux
goarch: amd64
pkg: cache
BenchmarkSyncCache_Get-12       10000000               222 ns/op               0 B/op          0 allocs/op
PASS
ok      cache   126.534s
*/
func BenchmarkSyncCache_Get(b *testing.B) {
	InitObjectCache(65535 * 200)
	rand.Seed(time.Now().UnixNano())
	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			// key := fmt.Sprintf("%d", rand.Int63n(60000000))
			_, ok := GetInt(rand.Int63n(60000000))
			if !ok {
				b.Error("获取错误")
			}
		}
	})
}

/**
bruce@:~/go/src/cache$ go test -bench=BenchmarkSyncMap_Get -v -run=none
init sync map end
goos: linux
goarch: amd64
pkg: cache
BenchmarkSyncMap_Get-12          2000000               841 ns/op              16 B/op          2 allocs/op
PASS
ok      cache   196.562s
*/
func BenchmarkSyncMap_Get(b *testing.B) {
	rand.Seed(time.Now().UnixNano())
	b.ReportAllocs()
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			key := fmt.Sprintf("%d", rand.Int63n(60000000))
			_, ok := syncMapDemo.Load(key)
			if !ok {
				b.Errorf(" key:%s", key)
			}
		}
	})
}
