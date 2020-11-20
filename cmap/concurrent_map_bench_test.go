package cmap_test

import (
	"math/rand"
	"strconv"
	"sync"
	"testing"
)

func generateSimpleMap()map[string]int{
	m := make(map[string]int)

	for i := 0 ; i < 10000 ; i++ {
		str := "test" + strconv.Itoa(i)
		m[str] = i
	}
	return m
}

func generateSyncMap() *sync.Map{
	var m *sync.Map
	m = new(sync.Map)
	for i := 0 ; i < 10000 ; i ++ {
		str := "test" + strconv.Itoa(i)
		m.Store(str, i)
	}
	return m
}
// >>>>>>>>>>>>>>   测试遍历map的效率 <<<<<<<<<<<<<<
func BenchmarkItems(b *testing.B){
	m := generateMap()

	for i := 0 ; i < b.N ; i ++ {
		m.Items()
	}
}
func BenchmarkSimpleMap(b *testing.B){
	m := generateSimpleMap()

	for i := 0 ; i < b.N ; i ++ {
		for _ , _ = range m{

		}
	}
}

func BenchmarkSyncMap(b *testing.B){
	m := generateSyncMap()

	for i := 0 ; i < b.N ; i ++ {
		m.Range(func(key , val interface{})bool{
			return true
		})
	}
}
// >>>>>>>>>>>>>>   测试获取元素的效率 <<<<<<<<<<<<<<
func BenchmarkConcurrentMap_Get(b *testing.B) {
	m := generateMap()

	for i := 0 ; i < b.N ; i ++ {
		x := rand.Intn(10000)
		m.Get("test" + strconv.Itoa(x))
	}
}

func BenchmarkSimpleMapGet(b *testing.B){
	m := generateSimpleMap()

	for i := 0 ; i < b.N ; i ++ {
		x := rand.Intn(10000)
		_ = m["test" + strconv.Itoa(x)]
	}
}

func BenchmarkSyncMapGet(b *testing.B){
	m := generateSyncMap()

	for i := 0 ; i < b.N ; i ++ {
		x := rand.Intn(10000)
		m.Load("test" + strconv.Itoa(x))
	}
}

// >>>>>>>>>>>>>>   测试设置元素的效率 <<<<<<<<<<<<<<

func BenchmarkConcurrentMap_Set(b *testing.B) {
	m := generateMap()

	for i := 0 ; i < b.N ; i ++ {
		x  := rand.Intn(20000)
		m.Set("test" + strconv.Itoa(x) , x)
	}
}

func BenchmarkSimpleMapSet(b *testing.B){
	m := generateSimpleMap()

	for i := 0 ; i < b.N ; i ++ {
		x  := rand.Intn(20000)
		m["test" + strconv.Itoa(x)] = x
	}
}

func BenchmarkSyncMapSet(b *testing.B){
	m := generateSyncMap()

	for i := 0 ; i < b.N ; i ++ {
		x := rand.Intn(20000)
		m.Store("test" + strconv.Itoa(x), i)
	}
}

// >>>>>>>>>>>>>>>> 测试并发set 元素 <<<<<<<<<<<<

func BenchmarkMultiSetMap(b *testing.B){
	m := generateMap()

	b.RunParallel(func (pb *testing.PB){
		for pb.Next(){
			m.Set("ket" , "value")
		}
	})
}

func BenchmarkMultiSetSyncMap(b *testing.B){
	m := generateSyncMap()

	b.RunParallel(func (pb *testing.PB){
		for pb.Next(){
			m.Store("key" , "value")
		}
	})
}

// >>>>>>>>>>>>>>>> 测试并发get 元素 <<<<<<<<<<<<

func BenchmarkMultiGetMap(b *testing.B){
	m := generateMap()

	b.RunParallel(func (pb *testing.PB){
		for pb.Next(){
			m.Get("key")
		}
	})
}

func BenchmarkMultiGetSyncMap(b *testing.B){
	m := generateSyncMap()

	b.RunParallel(func (pb *testing.PB){
		for pb.Next(){
			m.Load("key")
		}
	})
}

// >>>>>>>>>>>>>> 测试并发遍历 map <<<<<<<<<<<<<

func BenchmarkMultiRangeMap(b *testing.B){
	m := generateMap()

	b.RunParallel(func (pb *testing.PB){
		for pb.Next(){
			m.Items()
		}
	})
}

func BenchmarkMultiGetRangeMap(b *testing.B){
	m := generateSyncMap()

	b.RunParallel(func (pb *testing.PB){
		for pb.Next(){
			m.Range(func(key , val interface{})bool{
				return true
			})
		}
	})
}

